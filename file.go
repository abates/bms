package main

import (
	"encoding/base32"
	"github.com/abates/bms/database"
	"github.com/spf13/afero"
	"os"
	"strings"
)

type File struct {
	afero.File
	header *AssetHeader
}

func NewFile(backend afero.Fs, name string, mode os.FileMode) *File {
	return &File{
		header: NewAssetHeader(name, mode),
	}
}

func (f *File) ID() database.ID { return f.header.ID }

func (f *File) MarshalBinary() ([]byte, error) {
	return f.header.MarshalBinary()
}

func (f *File) Mode() os.FileMode                { return f.header.Mode }
func (f *File) Name() string                     { return f.header.Name }
func (f *File) Owner() database.ID               { return f.header.Owner }
func (f *File) SetName(name string) error        { f.header.SetName(name); return nil }
func (f *File) SetOwner(owner database.ID) error { f.header.Owner = owner; return nil }

func (f *File) Stat() (os.FileInfo, error) {
	return NewFileInfo(f)
}

func (f *File) RealPath() string {
	path := make([]string, 0)
	idString := strings.TrimRight(base32.StdEncoding.EncodeToString(f.header.ID[:]), "=")
	for i, s := range strings.Split(idString, "") {
		if i%4 == 0 {
			path = append(path, s)
		} else {
			path[len(path)-1] += s
		}
	}
	return "/" + strings.Join(path, "/")
}

func (f *File) UnmarshalBinary(data []byte) error {
	f.header = &AssetHeader{}
	return f.header.UnmarshalBinary(data)
}
