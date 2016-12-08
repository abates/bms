package filesystem

import (
	"github.com/abates/bms/database"
	"github.com/spf13/afero"
	"os"
	"path"
	"strings"
)

var (
	Db        database.Database
	BackendFs afero.Fs
)

func splitPath(name string) []string {
	if name == "/" {
		return []string{}
	} else if name[0] != '/' {
		name = "/" + name
	}
	name = path.Clean(name)
	return strings.Split(name, "/")[1:]
}

type FileSystem interface {
	Mkdir(name string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (Asset, error)
	RemoveAll(path string) error
	Rename(oldName, newName string) error
	Stat(name string) (os.FileInfo, error)
}
