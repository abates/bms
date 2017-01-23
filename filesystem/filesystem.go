package filesystem

import (
	"github.com/satori/go.uuid"
	"io"
	"os"
)

var (
	CreateMode = os.FileMode(0600)
)

func HasFlags(flag, search int) bool {
	return flag&search > 0
}

type ID []byte

func NewID() ID {
	id := uuid.NewV4()
	return ID(id[:])
}

func NewIDFromString(str string) (ID, error) {
	id, err := uuid.FromString(str)
	return ID(id[:]), err
}

func (id ID) String() string {
	i, _ := uuid.FromBytes(id)
	return i.String()
}

type File interface {
	io.ReadWriteCloser
	io.ReaderAt
	io.Seeker
	io.WriterAt

	Name() string
	Readdir(count int) ([]os.FileInfo, error)
	Readdirnames(n int) ([]string, error)
	Stat() (os.FileInfo, error)
	Sync() error
	Truncate(size int64) error
	WriteString(s string) (ret int, err error)
}

type Asset interface {
	File
	Metadata() *Metadata
}

type Backend interface {
	Mkdir(name string, perm os.FileMode) error
	MkdirAll(name string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
	RemoveAll(name string) error
	Rename(oldName, newName string) error
	Stat(name string) (os.FileInfo, error)
}

type FileSystem interface {
	Backend
	OpenAssetByID(id ID, flag int, perm os.FileMode) (File, error)
	OpenAsset(name string, flag int, perm os.FileMode) (Asset, error)
	StatByID(id ID) (os.FileInfo, error)
}
