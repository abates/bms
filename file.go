package main

import (
	"encoding/base32"
	"github.com/satori/go.uuid"
	"github.com/spf13/afero"
	"os"
	"strings"
)

type FileInfo struct {
	os.FileInfo
	name string
}

func (f FileInfo) Name() string { return f.name }

type File struct {
	Asset
	backend  afero.Fs
	name     string
	owner    *User
	perm     os.FileMode
	realPath string
}

func NewFile(backend afero.Fs, name string, perm os.FileMode) *File {
	return &File{
		backend:  backend,
		name:     name,
		realPath: newPath(),
		perm:     perm,
	}
}

func (f *File) IsDir() bool                { return false }
func (f *File) Name() string               { return f.name }
func (f *File) Owner() *User               { return f.owner }
func (f *File) SetOwner(owner *User) error { f.owner = owner; return nil }

func (f *File) Stat() (os.FileInfo, error) {
	var err error
	fi := FileInfo{
		name: f.Name(),
	}
	fi.FileInfo, err = f.backend.Stat(f.realPath)
	return fi, err
}

func (f *File) String() string { return f.name }

func newPath() string {
	path := make([]string, 0)
	id := uuid.NewV4()
	idString := strings.TrimRight(base32.StdEncoding.EncodeToString(id[:]), "=")
	for i, s := range strings.Split(idString, "") {
		if i%4 == 0 {
			path = append(path, s)
		} else {
			path[len(path)-1] += s
		}
	}
	return "/" + strings.Join(path, "/")
}
