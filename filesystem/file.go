package filesystem

import (
	"encoding/base32"
	"github.com/abates/bms/database"
	"github.com/spf13/afero"
	"os"
	"strings"
)

type File struct {
	afero.File
	metadata *Metadata
}

func NewFile(backend afero.Fs, name string, mode os.FileMode) *File {
	return &File{
		metadata: NewMetadata(name, mode),
	}
}

func (f *File) ID() database.ID { return f.metadata.ID }

func (f *File) MarshalBinary() ([]byte, error) {
	return f.metadata.MarshalBinary()
}

func (f *File) Mode() os.FileMode                { return f.metadata.Mode }
func (f *File) Name() string                     { return f.metadata.Name }
func (f *File) Owner() database.ID               { return f.metadata.Owner }
func (f *File) SetName(name string) error        { f.metadata.SetName(name); return nil }
func (f *File) SetOwner(owner database.ID) error { f.metadata.Owner = owner; return nil }

func (f *File) Stat() (os.FileInfo, error) {
	return NewFileInfo(f)
}

func (f *File) RealPath() string {
	path := make([]string, 0)
	idString := strings.TrimRight(base32.StdEncoding.EncodeToString(f.metadata.ID[:]), "=")
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
	f.metadata = &Metadata{}
	return f.metadata.UnmarshalBinary(data)
}

type FileInfo struct {
	os.FileInfo
	file *File
}

func NewFileInfo(file *File) (*FileInfo, error) {
	osfi, err := BackendFs.Stat(file.RealPath())
	fi := &FileInfo{
		FileInfo: osfi,
		file:     file,
	}
	return fi, err
}

func (fi *FileInfo) Name() string      { return fi.file.Name() }
func (fi *FileInfo) Mode() os.FileMode { return fi.file.Mode() }
func (fi *FileInfo) IsDir() bool       { return false }
