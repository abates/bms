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

type Folder struct {
	name     string
	size     int64
	mode     os.FileMode
	modTime  time.Time
	owner    *User
	children map[string]Asset
}

func NewFolder(name string, perm os.FileMode) *Folder {
	return &Folder{
		name:     name,
		size:     0,
		mode:     os.ModeDir | perm,
		modTime:  time.Now(),
		children: make(map[string]Asset),
	}
}

func (folder *Folder) addAsset(asset Asset) error {
	if _, found := folder.children[asset.Name()]; found {
		return os.ErrExist
	}
	folder.children[asset.Name()] = asset
	folder.modTime = time.Now()
	logger.Infof("Adding %s to '%s'.  Children are now %v", asset.Name(), folder.String(), folder.children)
	return nil
}

func (folder *Folder) Close() error {
	return nil
}

func (folder *Folder) Find(name []string) (Asset, error) {
	if len(name) == 0 {
		return folder, nil
	}

	if asset, found := folder.children[name[0]]; found {
		if next, ok := asset.(*Folder); ok {
			return next.Find(name[1:])
		} else if len(name) == 1 {
			return asset, nil
		}
	}
	return nil, os.ErrNotExist
}

func (folder *Folder) IsDir() bool { return true }

func (folder *Folder) Mkfolder(path []string, perm os.FileMode) (*Folder, error) {
	if len(path) == 1 {
		newFolder := NewFolder(path[0], perm)
		return newFolder, folder.addAsset(newFolder)
	} else if len(path) > 1 {
		child := folder.children[path[0]]
		switch next := child.(type) {
		case *Folder:
			return next.Mkfolder(path[1:], perm)
		case nil:
			return nil, os.ErrNotExist
		}
	}
	return nil, os.ErrNotExist
}

func (folder *Folder) Mode() os.FileMode  { return folder.mode }
func (folder *Folder) ModTime() time.Time { return folder.modTime }
func (folder *Folder) Name() string       { return folder.name }
func (folder *Folder) Owner() *User       { return folder.owner }

func (folder *Folder) Readdir(count int) ([]os.FileInfo, error) {
	children := make([]os.FileInfo, 0)
	for name, child := range folder.children {
		if fi, err := child.Stat(); err == nil {
			logger.Infof("Readdir name: %s", fi.Name())
			children = append(children, fi)
		} else {
			logger.Infof("Readdir name %s/%s error %v", folder.Name(), name, err)
		}
	}
	logger.WithFields(map[string]interface{}{
		"children": children,
	}).Infof("Readdir %s", folder.name)
	return children, nil
}

func (folder *Folder) Read([]byte) (int, error)               { return 0, ErrIsFolder }
func (folder *Folder) RemoveAll(path []string) error          { return nil }
func (folder *Folder) Rename(oldName, newName []string) error { return nil }
func (folder *Folder) Seek(int64, int) (int64, error)         { return 0, ErrIsFolder }
func (folder *Folder) SetOwner(owner *User) error             { folder.owner = owner; return nil }

func (folder *Folder) Size() int64 {
	return 0
}

func (folder *Folder) Stat() (os.FileInfo, error) {
	return folder, nil
}

func (folder *Folder) String() string            { return fmt.Sprintf("%p:%s", folder, folder.name) }
func (folder *Folder) Sys() interface{}          { return nil }
func (folder *Folder) Write([]byte) (int, error) { return 0, ErrIsFolder }
