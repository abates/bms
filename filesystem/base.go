package filesystem

import (
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var (
	ErrNotSupported = fmt.Errorf("Operation is not supported")
)

func pathForID(id ID) string {
	name := make([]string, 0)
	idString := strings.TrimRight(base32.StdEncoding.EncodeToString(id[:]), "=")
	for i, s := range strings.Split(idString, "") {
		if i%4 == 0 {
			name = append(name, s)
		} else {
			name[len(name)-1] += s
		}
	}
	return "/" + strings.Join(name, "/")
}

type BaseFile struct {
	File
	fs       *BaseFileSystem
	metadata *Metadata
}

func NewBaseFile(file File, metadata *Metadata) *BaseFile {
	return &BaseFile{
		File:     file,
		metadata: metadata,
	}
}

func (file *BaseFile) Metadata() *Metadata {
	return file.metadata
}

type BaseFileSystem struct {
	backend Backend
}

func NewBaseFileSystem(backend Backend) *BaseFileSystem {
	return &BaseFileSystem{
		backend: backend,
	}
}

func (fs *BaseFileSystem) loadMetadata(metadata *Metadata) error {
	file, err := fs.backend.OpenFile(pathForID(metadata.ID), os.O_RDONLY, CreateMode)
	if err == nil {
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err == nil {
			err = metadata.UnmarshalJSON(data)
		}
	}
	return err
}

func (fs *BaseFileSystem) Mkdir(name string, perm os.FileMode) error    { return ErrNotSupported }
func (fs *BaseFileSystem) MkdirAll(name string, perm os.FileMode) error { return ErrNotSupported }

func (fs *BaseFileSystem) OpenAsset(name string, flag int, perm os.FileMode) (Asset, error) {
	file, err := fs.OpenFile(name, flag, perm)
	return file.(Asset), err
}

func (fs *BaseFileSystem) OpenFile(name string, flag int, perm os.FileMode) (file File, err error) {
	id, err := NewIDFromString(name)

	if err == nil {
		file, err = fs.OpenAssetByID(id, flag, perm)
	}
	return file, err
}

func (fs *BaseFileSystem) OpenAssetByID(id ID, flag int, perm os.FileMode) (basefile Asset, err error) {
	var file File
	metadata := &Metadata{
		ID: id,
	}
	if id == nil {
		if HasFlags(flag, os.O_CREATE) {
			metadata.ID = NewID()
			metadata.FileID = NewID()
			metadata.Mode = os.ModePerm & perm
			fs.saveMetadata(metadata)
		} else {
			return nil, &os.PathError{"OpenFile", "", fmt.Errorf("invalid id")}
		}
	} else {
		err = fs.loadMetadata(metadata)
		if os.IsNotExist(err) {
			metadata.FileID = NewID()
			metadata.Mode = os.ModePerm & perm
			err = fs.saveMetadata(metadata)
		}
	}

	if err == nil {
		file, err = fs.backend.OpenFile(pathForID(metadata.FileID), flag, CreateMode)
		if err == nil {
			basefile = NewBaseFile(file, metadata)
		}
	}
	return basefile, err
}

func (fs *BaseFileSystem) Rename(oldName, newName string) (err error) { return ErrNotSupported }

func (fs *BaseFileSystem) saveMetadata(metadata *Metadata) error {
	file, err := fs.backend.OpenFile(pathForID(metadata.ID), os.O_CREATE|os.O_RDWR, CreateMode)
	if err == nil {
		defer file.Close()
		data, err := metadata.MarshalJSON()
		if err == nil {
			_, err = file.Write(data)
		}
	}
	return err
}

func (fs *BaseFileSystem) StatByID(id ID) (fi os.FileInfo, err error) {
	metadata := &Metadata{
		ID: id,
	}
	err = fs.loadMetadata(metadata)
	if err == nil {
		fi, err = fs.backend.Stat(pathForID(metadata.FileID))
		fi = &FileInfo{fi, metadata}
	}
	return fi, err
}

func (fs *BaseFileSystem) Stat(name string) (fi os.FileInfo, err error) {
	var id ID
	id, err = NewIDFromString(name)
	if err == nil {
		fi, err = fs.StatByID(id)
	}
	return fi, err
}
