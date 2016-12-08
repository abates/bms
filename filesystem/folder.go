package filesystem

import (
	"bytes"
	"fmt"
	logger "github.com/Sirupsen/logrus"
	"github.com/abates/bms/database"
	"github.com/lunixbochs/struc"
	"io"
	"os"
	"path"
	"time"
)

var (
	ErrIsFolder  = fmt.Errorf("Asset is a folder")
	ErrNotFolder = fmt.Errorf("Asset is not a folder")
)

type FolderEntry struct {
	ID       database.ID `struc:"[16]byte"`
	NameLen  int         `struc:"uint16,sizeof=Name"`
	Name     string
	IsFolder bool
}

func NewFolderEntry(id database.ID, name string, isFolder bool) *FolderEntry {
	return &FolderEntry{
		ID:       id,
		NameLen:  len(name),
		Name:     name,
		IsFolder: isFolder,
	}
}

func (fe *FolderEntry) Asset() (asset Asset, err error) {
	if fe.IsFolder {
		asset = &Folder{}
		err = Db.Find(fe.ID, asset.(*Folder))
	} else {
		asset = &File{}
		err = Db.Find(fe.ID, asset.(*File))
	}
	return asset, err
}

func (fe *FolderEntry) Remove() (err error) {
	asset, _ := fe.Asset()
	switch file := asset.(type) {
	case *Folder:
		file.RemoveAll()
	case *File:
		BackendFs.Remove(file.RealPath())
	}

	if err == nil {
		err = Db.Delete(fe.ID)
	}
	return
}

func (fe *FolderEntry) SetName(name string) {
	fe.NameLen = len(name)
	fe.Name = name
}

type Folder struct {
	metadata *Metadata
	entries  map[string]*FolderEntry
}

func NewFolder(name string, perm os.FileMode) *Folder {
	return &Folder{
		metadata: NewMetadata(name, os.ModeDir|perm),
		entries:  make(map[string]*FolderEntry),
	}
}

func (folder *Folder) addAsset(asset Asset) error {
	if _, found := folder.entries[asset.Name()]; found {
		return os.ErrExist
	}

	switch file := asset.(type) {
	case *Folder:
		folder.entries[asset.Name()] = NewFolderEntry(file.ID(), asset.Name(), true)
	case *File:
		folder.entries[asset.Name()] = NewFolderEntry(file.ID(), asset.Name(), false)
	default:
		logger.Warnf("Unknown asset type %T", asset)
	}

	folder.metadata.ModTime = time.Now().Unix()
	Db.Save(folder.ID(), folder)
	return nil
}

func (folder *Folder) Close() error {
	return nil
}

func (folder *Folder) Find(name []string) (Asset, error) {
	if folder == nil {
		return nil, os.ErrNotExist
	}

	if len(name) == 0 {
		return folder, nil
	}

	if entry, found := folder.entries[name[0]]; found {
		asset, _ := entry.Asset()
		if asset == nil {
			logger.Warnf("Folder entry %s/%s points to non-existant folder %s", folder.Name(), name[0], entry.ID)
		} else if folder, ok := asset.(*Folder); ok {
			return folder.Find(name[1:])
		} else if len(name) == 1 {
			return asset, nil
		}
	}
	return nil, os.ErrNotExist
}

func (folder *Folder) ID() database.ID { return folder.metadata.ID }

func (folder *Folder) MarshalBinary() ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := folder.metadata.Pack(buffer)
	if err == nil {
		for _, entry := range folder.entries {
			err = struc.Pack(buffer, entry)
			if err != nil {
				break
			}
		}
	}
	return buffer.Bytes(), err
}

func (folder *Folder) Mkfolder(name string, perm os.FileMode) (newFolder *Folder, err error) {
	newFolder = NewFolder(name, perm)
	err = folder.addAsset(newFolder)
	err = Db.Save(newFolder.ID(), newFolder)
	return
}

func (folder *Folder) Mode() os.FileMode  { return folder.metadata.Mode }
func (folder *Folder) ModTime() time.Time { return time.Unix(folder.metadata.ModTime, 0) }

func (folder *Folder) Move(name string, newFolder *Folder) (err error) {
	if entry, found := folder.entries[name]; found {
		delete(folder.entries, name)
		newFolder.entries[name] = entry
		Db.Save(folder.ID(), folder)
		Db.Save(newFolder.ID(), newFolder)
	} else {
		err = os.ErrNotExist
	}
	return err
}

func (folder *Folder) Name() string       { return folder.metadata.Name }
func (folder *Folder) Owner() database.ID { return folder.metadata.Owner }

func (folder *Folder) Readdir(count int) ([]os.FileInfo, error) {
	entries := make([]os.FileInfo, 0)
	for name, entry := range folder.entries {
		asset, err := entry.Asset()
		if err != nil {
			return entries, err
		}
		if asset == nil {
			logger.Warnf("Folder entry %s/%s points to non-existant folder %s", folder.Name(), name, entry.ID)
		} else if fi, err := asset.Stat(); err == nil {
			entries = append(entries, fi)
		} else {
			logger.WithFields(map[string]interface{}{
				"parent": folder.Name(),
				"child":  name,
				"error":  err,
			}).Warn("Readdir failed to stat %s/%s", folder.Name(), name)
		}
	}
	return entries, nil
}

func (folder *Folder) Read([]byte) (int, error) { return 0, ErrIsFolder }

func (folder *Folder) Remove(name string) error {
	err := os.ErrNotExist
	if entry, found := folder.entries[name]; found {
		err = entry.Remove()
		if err == nil {
			delete(folder.entries, name)
			Db.Save(folder.ID(), folder)
		}
	}
	return err
}

func (folder *Folder) RemoveAll() (err error) {
	for _, entry := range folder.entries {
		err1 := entry.Remove()
		if err == nil {
			err = err1
		}
	}
	return
}

func (folder *Folder) Rename(oldName, newName string) error {
	err := os.ErrNotExist
	if entry, found := folder.entries[oldName]; found {
		asset, _ := entry.Asset()
		if asset != nil {
			err = asset.SetName(newName)
		}

		if err == nil {
			entry.SetName(newName)
			delete(folder.entries, oldName)
			folder.entries[newName] = entry
			Db.Save(folder.ID(), folder)
		}
	}
	return err
}

func (folder *Folder) UnmarshalBinary(data []byte) error {
	folder.metadata = &Metadata{}
	buffer := bytes.NewBuffer(data)
	err := folder.metadata.Unpack(buffer)
	folder.entries = make(map[string]*FolderEntry)
	for err == nil {
		entry := &FolderEntry{}
		err = struc.Unpack(buffer, entry)
		if err == nil {
			folder.entries[entry.Name] = entry
		}
	}

	if err == io.EOF {
		err = nil
	}
	return err
}

func (folder *Folder) Seek(int64, int) (int64, error)   { return 0, ErrIsFolder }
func (folder *Folder) SetName(name string) error        { folder.metadata.SetName(name); return nil }
func (folder *Folder) SetOwner(owner database.ID) error { folder.metadata.Owner = owner; return nil }

func (folder *Folder) Stat() (os.FileInfo, error) {
	return NewFolderInfo(folder)
}

func (folder *Folder) Write([]byte) (int, error) { return 0, ErrIsFolder }

type FolderInfo struct {
	folder *Folder
}

func NewFolderInfo(folder *Folder) (*FolderInfo, error) {
	return &FolderInfo{folder}, nil
}

func (fi *FolderInfo) Mode() os.FileMode  { return fi.folder.Mode() }
func (fi *FolderInfo) ModTime() time.Time { return fi.folder.ModTime() }
func (fi *FolderInfo) Name() string       { return fi.folder.Name() }
func (fi *FolderInfo) Size() int64        { return 0 }
func (fi *FolderInfo) Sys() interface{}   { return nil }

func (fi *FolderInfo) IsDir() bool { return true }

type FolderFileSystem struct {
	userid database.ID
	root   *Folder
}

func NewFolderFileSystem(root *Folder, userid database.ID) *FolderFileSystem {
	return &FolderFileSystem{
		userid: userid,
		root:   root,
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
			err = newFolder.SetOwner(fs.userid)
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
			file := NewFile(BackendFs, filename, perm)
			err = BackendFs.MkdirAll(path.Dir(file.RealPath()), 0700)
			if err == nil {
				asset = file
				err = base.addAsset(file)
				if err == nil {
					Db.Save(file.ID(), file)
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
		file.File, err = BackendFs.OpenFile(file.RealPath(), flag, 0600)
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
