package main

import (
	"bytes"
	"fmt"
	"github.com/abates/bms/database"
	"github.com/lunixbochs/struc"
	"io"
	"os"
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
		err = db.Find(fe.ID, asset.(*Folder))
	} else {
		asset = &File{}
		err = db.Find(fe.ID, asset.(*File))
	}
	return asset, err
}

func (fe *FolderEntry) Remove() (err error) {
	asset, _ := fe.Asset()
	switch file := asset.(type) {
	case *Folder:
		file.RemoveAll()
	case *File:
		backendFs.Remove(file.RealPath())
	}

	if err == nil {
		err = db.Delete(fe.ID)
	}
	return
}

func (fe *FolderEntry) SetName(name string) {
	fe.NameLen = len(name)
	fe.Name = name
}

type Folder struct {
	header  *AssetHeader
	entries map[string]*FolderEntry
}

func NewFolder(name string, perm os.FileMode) *Folder {
	return &Folder{
		header:  NewAssetHeader(name, os.ModeDir|perm),
		entries: make(map[string]*FolderEntry),
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

	folder.header.ModTime = time.Now().Unix()
	db.Save(folder.ID(), folder)
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

func (folder *Folder) ID() database.ID { return folder.header.ID }

func (folder *Folder) MarshalBinary() ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := folder.header.Pack(buffer)
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
	err = db.Save(newFolder.ID(), newFolder)
	return
}

func (folder *Folder) Mode() os.FileMode  { return folder.header.Mode }
func (folder *Folder) ModTime() time.Time { return time.Unix(folder.header.ModTime, 0) }

func (folder *Folder) Move(name string, newFolder *Folder) (err error) {
	if entry, found := folder.entries[name]; found {
		delete(folder.entries, name)
		newFolder.entries[name] = entry
		db.Save(folder.ID(), folder)
		db.Save(newFolder.ID(), newFolder)
	} else {
		err = os.ErrNotExist
	}
	return err
}

func (folder *Folder) Name() string       { return folder.header.Name }
func (folder *Folder) Owner() database.ID { return folder.header.Owner }

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
			db.Save(folder.ID(), folder)
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
			db.Save(folder.ID(), folder)
		}
	}
	return err
}

func (folder *Folder) UnmarshalBinary(data []byte) error {
	folder.header = &AssetHeader{}
	buffer := bytes.NewBuffer(data)
	err := folder.header.Unpack(buffer)
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
func (folder *Folder) SetName(name string) error        { folder.header.SetName(name); return nil }
func (folder *Folder) SetOwner(owner database.ID) error { folder.header.Owner = owner; return nil }

func (folder *Folder) Stat() (os.FileInfo, error) {
	return NewFolderInfo(folder)
}

func (folder *Folder) Write([]byte) (int, error) { return 0, ErrIsFolder }
