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
		id:      generateID(),
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
		if asset == nil {
			logger.Warnf("Folder entry %s/%s points to non-existant folder %s", folder.Name(), name[0], entry.id)
		} else if folder, ok := asset.(*Folder); ok {
			return folder.Find(name[1:])
		} else if len(name) == 1 {
			return asset, nil
		}
	}
	return nil, os.ErrNotExist
}

func (folder *Folder) ID() ID { return folder.id }

func (folder *Folder) Mkfolder(name string, perm os.FileMode) (newFolder *Folder, err error) {
	newFolder = NewFolder(name, perm)
	err = folder.addAsset(newFolder)
	if err == nil {
		db.Save(newFolder)
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
		asset := db.Find(entry.id)
		if asset == nil {
			logger.Warnf("Folder entry %s/%s points to non-existant folder %s", folder.Name(), name, entry.id)
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

func (folder *Folder) Read([]byte) (int, error)               { return 0, ErrIsFolder }
func (folder *Folder) RemoveAll(path []string) error          { return nil }
func (folder *Folder) Rename(oldName, newName []string) error { return nil }
func (folder *Folder) Seek(int64, int) (int64, error)         { return 0, ErrIsFolder }

func (folder *Folder) SetOwner(owner *User) error { folder.owner = owner; return nil }

func (folder *Folder) Stat() (os.FileInfo, error) {
	return NewFolderInfo(folder)
}

func (folder *Folder) Write([]byte) (int, error) { return 0, ErrIsFolder }
