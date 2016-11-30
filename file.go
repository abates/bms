package main

import (
	"encoding/base32"
	"github.com/satori/go.uuid"
	"github.com/spf13/afero"
	"os"
	"strings"
)

type File struct {
	afero.File
	backend  afero.Fs
	name     string
	id       ID
	owner    *User
	mode     os.FileMode
	realPath string
}

func NewFile(backend afero.Fs, name string, mode os.FileMode) *File {
	return &File{
		backend:  backend,
		name:     name,
		realPath: newPath(),
		mode:     mode,
	}
}

func (f *File) ID() ID                     { return f.id }
func (f *File) Mode() os.FileMode          { return f.mode }
func (f *File) Name() string               { return f.name }
func (f *File) Owner() *User               { return f.owner }
func (f *File) SetOwner(owner *User) error { f.owner = owner; return nil }

func (f *File) Stat() (os.FileInfo, error) {
	return NewFileInfo(f)
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
