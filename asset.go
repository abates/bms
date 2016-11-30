package main

import (
	"io"
	"os"
	"time"
)

type Asset interface {
	io.ReadWriteCloser
	io.Seeker
	ID() ID
	Mode() os.FileMode
	Name() string
	Owner() *User
	Readdir(count int) ([]os.FileInfo, error)
	SetOwner(*User) error
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
