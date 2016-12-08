package filesystem

import (
	"github.com/abates/bms/database"
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
