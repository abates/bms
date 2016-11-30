package main

import (
	"github.com/spf13/afero"
	"golang.org/x/net/context"
	"golang.org/x/net/webdav"
	"os"
	"path"
	"strings"
)

type WebdavFileSystem struct {
	*webdav.Handler
	root FileSystem
}

func NewWebdavFileSystem(user *User) (wfs *WebdavFileSystem, err error) {
	wfs = &WebdavFileSystem{
		Handler: &webdav.Handler{
			Prefix:     "/webdav",
			LockSystem: webdav.NewMemLS(),
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

func cleanPath(name string) []string {
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
	user    *User
	root    *Folder
	backend afero.Fs
}

func NewFolderFileSystem(user *User, backend afero.Fs) *FolderFileSystem {
	return &FolderFileSystem{
		user:    user,
		root:    NewFolder("", 0700),
		backend: backend,
	}
}

func (fs *FolderFileSystem) Mkdir(name string, perm os.FileMode) error {
	folder, err := fs.root.Mkfolder(cleanPath(name), perm)
	if err == nil {
		err = folder.SetOwner(fs.user)
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
	name = path.Clean(name)
	dirname := path.Dir(name)
	filename := path.Base(name)

	if name == "/" {
		if hasFlags(flag, os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) {
			return nil, &os.PathError{"open", name, ErrIsFolder}
		}
		logger.Infof("OpenFile Returning root filesystem %s", fs.root.String())
		return fs.root, nil
	}

	asset, err = fs.root.Find(cleanPath(dirname))

	var base *Folder
	if folder, ok := asset.(*Folder); ok {
		base = folder
	} else {
		logger.Infof("OpenFile base path %s is not a folder", dirname)
		return nil, &os.PathError{"open", dirname, ErrNotFolder}
	}

	asset, err = base.Find([]string{filename})
	if os.IsNotExist(err) {
		if hasFlags(flag, os.O_CREATE) {
			err = nil
			file := NewFile(filename, perm)
			err = fs.backend.MkdirAll(path.Dir(file.realPath), 0700)
			if err == nil {
				asset = file
				err = base.addAsset(file)
			}
		}
	}

	logger.Infof("OpenFile found a %T named %s", asset, asset.Name())
	switch file := asset.(type) {
	case *Folder:
		if hasFlags(flag, os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) {
			return nil, &os.PathError{"open", name, ErrIsFolder}
		}
	case *File:
		file.File, err = fs.backend.OpenFile(file.realPath, flag, 0600)
		logger.Infof("Opening file %s File: %v flag: %s err: %v", file.realPath, file.File, flagString(flag), err)
	}
	return asset, err
}

func (fs *FolderFileSystem) RemoveAll(path string) error {
	return nil
}

func (fs *FolderFileSystem) Rename(oldName, newName string) error { return nil }

func (fs *FolderFileSystem) Stat(name string) (fi os.FileInfo, err error) {
	asset, err := fs.root.Find(cleanPath(name))
	if err == nil {
		fi, err = asset.Stat()
	}
	return
}
