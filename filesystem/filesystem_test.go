package filesystem

import (
	"github.com/spf13/afero"
	"os"
)

func id(str string) ID {
	id, _ := NewIDFromString(str)
	return id
}

type testfs struct {
	afero.Fs
}

func newTestfs() *testfs {
	return &testfs{
		Fs: afero.NewMemMapFs(),
	}
}

func (fs testfs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return fs.Fs.OpenFile(name, flag, perm)
}
