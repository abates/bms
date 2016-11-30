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
	name     string
	owner    *User
	perm     os.FileMode
	realPath string
}

func NewFile(name string, perm os.FileMode) *File {
	return &File{
		name:     name,
		realPath: newPath(),
		perm:     perm,
	}
}

func (f *File) IsDir() bool                { return false }
func (f *File) Name() string               { return f.name }
func (f *File) Owner() *User               { return f.owner }
func (f *File) SetOwner(owner *User) error { f.owner = owner; return nil }
func (f *File) String() string             { return f.name }

func (f *File) Write(b []byte) (int, error) {
	logger.Infof("Writing %d bytes to %s", len(b), f.name)
	return f.File.Write(b)
}

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
