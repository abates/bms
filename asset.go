package main

import (
	"bytes"
	"github.com/abates/bms/database"
	"github.com/lunixbochs/struc"
	"io"
	"os"
	"time"
)

type Asset interface {
	io.ReadWriteCloser
	io.Seeker
	ID() database.ID
	Mode() os.FileMode
	Name() string
	Owner() database.ID
	Readdir(count int) ([]os.FileInfo, error)
	SetName(string) error
	SetOwner(database.ID) error
	Stat() (os.FileInfo, error)
}

type AssetInfo interface {
	IsDir() bool
	Mode() os.FileMode
	ModTime() time.Time
	Name() string
	Size() int64
	Sys() interface{}
}

type AssetHeader struct {
	ID      database.ID `struc:"[16]byte"`
	NameLen int         `struc:"int16,sizeof=Name"`
	Name    string
	Owner   database.ID `struc:"[16]byte"`
	Mode    os.FileMode `struc:"uint32"`
	ModTime int64       `struc:"int64"`
}

func NewAssetHeader(name string, mode os.FileMode) *AssetHeader {
	return &AssetHeader{
		ID:      database.NewID(),
		NameLen: len(name),
		Name:    name,
		Mode:    mode,
		ModTime: time.Now().Unix(),
	}
}

func (ah *AssetHeader) SetName(name string) {
	ah.NameLen = len(name)
	ah.Name = name
}

func (ah *AssetHeader) UnmarshalBinary(data []byte) error {
	return ah.Unpack(bytes.NewBuffer(data))
}

func (ah *AssetHeader) Unpack(reader io.Reader) error {
	return struc.Unpack(reader, ah)
}

func (ah *AssetHeader) MarshalBinary() ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := ah.Pack(buffer)
	return buffer.Bytes(), err
}

func (ah *AssetHeader) Pack(writer io.Writer) error {
	return struc.Pack(writer, ah)
}

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

type FileInfo struct {
	os.FileInfo
	file *File
}

func NewFileInfo(file *File) (*FileInfo, error) {
	osfi, err := backendFs.Stat(file.RealPath())
	fi := &FileInfo{
		FileInfo: osfi,
		file:     file,
	}
	return fi, err
}

func (fi *FileInfo) Name() string      { return fi.file.Name() }
func (fi *FileInfo) Mode() os.FileMode { return fi.file.Mode() }
func (fi *FileInfo) IsDir() bool       { return false }
