package main

import (
	"fmt"
	"os"
	"time"
)

var (
	ErrIsFolder  = fmt.Errorf("Asset is a folder")
	ErrNotFolder = fmt.Errorf("Asset is not a folder")
)

type FolderEntry struct {
	id       ID
	isFolder bool
}

type Folder struct {
	name    string
	entries map[string]FolderEntry
	id      ID
	mode    os.FileMode
	modTime time.Time
	owner   *User
}

func NewFolder(name string, perm os.FileMode) *Folder {
	return &Folder{
		name:    name,
		mode:    os.ModeDir | perm,
		modTime: time.Now(),
		entries: make(map[string]FolderEntry),
	}
}

func (folder *Folder) addAsset(asset Asset) error {
	if _, found := folder.entries[asset.Name()]; found {
		return os.ErrExist
	}

	switch file := asset.(type) {
	case *Folder:
		folder.entries[asset.Name()] = FolderEntry{file.ID(), true}
	case *File:
		folder.entries[asset.Name()] = FolderEntry{file.ID(), false}
	default:
		logger.Warnf("Unknown asset type %T", asset)
	}

	folder.modTime = time.Now()
	db.Save(folder)
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
		asset := db.Find(entry.id)
		if folder, ok := asset.(*Folder); ok {
			return folder.Find(name[1:])
		} else if len(name) == 1 {
			return asset, nil
		}
	}
	return nil, os.ErrNotExist
}

func (folder *Folder) ID() ID { return folder.id }

func (folder *Folder) Mkfolder(path []string, perm os.FileMode) (newFolder *Folder, err error) {
	err = os.ErrNotExist
	if len(path) == 1 {
		newFolder = NewFolder(path[0], perm)
		err = folder.addAsset(newFolder)
	} else if len(path) > 1 {
		if entry, found := folder.entries[path[0]]; found {
			if entry.isFolder {
				folder := db.Find(entry.id).(*Folder)
				return folder.Mkfolder(path[1:], perm)
			}
			err = ErrNotFolder
		}
	}
	return
}

func (folder *Folder) Mode() os.FileMode  { return folder.mode }
func (folder *Folder) ModTime() time.Time { return folder.modTime }
func (folder *Folder) Name() string       { return folder.name }

func (folder *Folder) Owner() *User { return folder.owner }

func (folder *Folder) Readdir(count int) ([]os.FileInfo, error) {
	entries := make([]os.FileInfo, 0)
	for name, entry := range folder.entries {
		fi, err := db.Find(entry.id).Stat()
		if err == nil {
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

func (folder *Folder) Read([]byte) (int, error)               { return 0, ErrIsFolder }
func (folder *Folder) RemoveAll(path []string) error          { return nil }
func (folder *Folder) Rename(oldName, newName []string) error { return nil }
func (folder *Folder) Seek(int64, int) (int64, error)         { return 0, ErrIsFolder }

func (folder *Folder) SetOwner(owner *User) error { folder.owner = owner; return nil }

/*func (folder *Folder) Size() int64 {
	return 0
}*/

func (folder *Folder) Stat() (os.FileInfo, error) {
	return NewFolderInfo(folder)
}

func (folder *Folder) String() string { return fmt.Sprintf("%p:%s", folder, folder.name) }

//func (folder *Folder) Sys() interface{}          { return nil }
func (folder *Folder) Write([]byte) (int, error) { return 0, ErrIsFolder }
