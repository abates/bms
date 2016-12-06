package main

import (
	"golang.org/x/net/context"
	"golang.org/x/net/webdav"
	"net/http"
	"os"
	"path"
	"strings"
)

type WebdavFileSystem struct {
	*webdav.Handler
	root FileSystem
}

var lockSystem webdav.LockSystem

func NewWebdavFileSystem(user *User) (wfs *WebdavFileSystem, err error) {
	if lockSystem == nil {
		lockSystem = webdav.NewMemLS()
	}

	wfs = &WebdavFileSystem{
		Handler: &webdav.Handler{
			Prefix:     "/webdav",
			LockSystem: lockSystem,
			Logger: func(r *http.Request, err error) {
				if err != nil {
					logger.WithFields(map[string]interface{}{
						"method": r.Method,
						"url":    r.URL,
						"error":  err,
					}).Warnf("%s Failed", r.Method)
				}
			},
		},
		root: user.fs,
	}
	wfs.Handler.FileSystem = wfs
	return
}

func (fs *WebdavFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return fs.root.Mkdir(name, perm)
}

func (fs *WebdavFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return fs.root.OpenFile(name, flag, perm)
}

func (fs *WebdavFileSystem) RemoveAll(ctx context.Context, name string) error {
	return fs.root.RemoveAll(name)
}

func (fs *WebdavFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	return fs.root.Rename(oldName, newName)
}

func (fs *WebdavFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return fs.root.Stat(name)
}

func splitPath(name string) []string {
	if name == "/" {
		return []string{}
	} else if name[0] != '/' {
		name = "/" + name
	}
	name = path.Clean(name)
	return strings.Split(name, "/")[1:]
}

type FileSystem interface {
	Mkdir(name string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (Asset, error)
	RemoveAll(path string) error
	Rename(oldName, newName string) error
	Stat(name string) (os.FileInfo, error)
}

type FolderFileSystem struct {
	user *User
	root *Folder
}

func NewFolderFileSystem(user *User) *FolderFileSystem {
	return &FolderFileSystem{
		user: user,
		root: NewFolder("", 0700),
	}
}

func cleanPath(name string) (string, string) {
	name = path.Clean(name)
	return path.Dir(name), path.Base(name)
}

func (fs *FolderFileSystem) Mkdir(name string, perm os.FileMode) error {
	dirname, filename := cleanPath(name)
	if name == "/" {
		return os.ErrExist
	}

	asset, err := fs.root.Find(splitPath(dirname))
	if parent, ok := asset.(*Folder); ok {
		var newFolder *Folder
		newFolder, err = parent.Mkfolder(filename, perm)
		if err == nil {
			err = newFolder.SetOwner(fs.user.ID)
		}
	}
	return err
}

func validatePermissions(asset Asset, perm os.FileMode) error {
	return nil
}

func flagString(flag int) string {
	s := ""

	for name, i := range map[string]int{
		"RDONLY": os.O_RDONLY,
		"WRONLY": os.O_WRONLY,
		"RDWR":   os.O_RDWR,
		"APPEND": os.O_APPEND,
		"CREATE": os.O_CREATE,
		"EXCL":   os.O_EXCL,
		"SYNC":   os.O_SYNC,
		"TRUNC":  os.O_TRUNC} {
		if flag&i > 0 {
			if s != "" {
				s += "|"
			}
			s += name
		}
	}
	return s
}

func hasFlags(flag, search int) bool {
	return flag&search > 0
}

func (fs *FolderFileSystem) OpenFile(name string, flag int, perm os.FileMode) (asset Asset, err error) {
	dirname, filename := cleanPath(name)

	if name == "/" {
		if hasFlags(flag, os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) {
			return nil, &os.PathError{"open", name, ErrIsFolder}
		}
		return fs.root, nil
	}

	asset, err = fs.root.Find(splitPath(dirname))

	var base *Folder
	if folder, ok := asset.(*Folder); ok {
		base = folder
	} else {
		return nil, &os.PathError{"open", dirname, ErrNotFolder}
	}

	asset, err = base.Find([]string{filename})
	if os.IsNotExist(err) {
		if hasFlags(flag, os.O_CREATE) {
			err = nil
			file := NewFile(backendFs, filename, perm)
			err = backendFs.MkdirAll(path.Dir(file.RealPath()), 0700)
			if err == nil {
				asset = file
				err = base.addAsset(file)
				if err == nil {
					db.Save(file)
				}
			}
		}
	}

	switch file := asset.(type) {
	case *Folder:
		if hasFlags(flag, os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) {
			return nil, &os.PathError{"open", name, ErrIsFolder}
		}
	case *File:
		file.File, err = backendFs.OpenFile(file.RealPath(), flag, 0600)
	}
	return asset, err
}

func (fs *FolderFileSystem) RemoveAll(name string) error {
	dirname, filename := cleanPath(name)
	asset, err := fs.root.Find(splitPath(dirname))
	if err == nil {
		switch file := asset.(type) {
		case *Folder:
			err = file.Remove(filename)
		case *File:
			err = &os.PathError{"removeall", dirname, ErrNotFolder}
		}
	}
	return err
}

func (fs *FolderFileSystem) Rename(oldName, newName string) (err error) {
	oldDirname, oldFilename := cleanPath(oldName)
	newDirname, newFilename := cleanPath(newName)

	asset, err := fs.root.Find(splitPath(oldDirname))
	if err != nil {
		return
	}

	if oldDir, ok := asset.(*Folder); ok {
		if oldName != newName {
			err = oldDir.Rename(oldFilename, newFilename)
		}

		if err == nil && oldDirname != newDirname {
			asset, err = fs.root.Find(splitPath(newDirname))
			if err != nil {
				return err
			}

			if newDir, ok := asset.(*Folder); ok {
				err = oldDir.Move(newFilename, newDir)
			} else {
				err = &os.PathError{"rename", newDirname, ErrNotFolder}
			}
		}
	} else {
		err = &os.PathError{"rename", newDirname, ErrNotFolder}
	}
	return
}

func (fs *FolderFileSystem) Stat(name string) (os.FileInfo, error) {
	asset, err := fs.root.Find(splitPath(name))
	if err == nil {
		return asset.Stat()
	}

	return nil, err
}
