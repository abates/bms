package filesystem

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/abates/bms/database"
	"golang.org/x/net/context"
	"golang.org/x/net/webdav"
	"net/http"
	"os"
)

type WebdavFileSystem struct {
	*webdav.Handler
	root FileSystem
}

var lockSystem webdav.LockSystem

func NewWebdavFileSystem(userid database.ID, root *Folder) *WebdavFileSystem {
	if lockSystem == nil {
		lockSystem = webdav.NewMemLS()
	}

	fs := NewFolderFileSystem(root, userid)
	wfs := &WebdavFileSystem{
		Handler: &webdav.Handler{
			Prefix:     "/webdav",
			LockSystem: lockSystem,
			Logger: func(r *http.Request, err error) {
				if err != nil {
					logger.WithFields(map[string]interface{}{
						"method": r.Method,
						"url":    r.URL,
						"error":  err,
					}).Warnf("%s Failed", r.Method)
				}
			},
		},
		root: fs,
	}
	wfs.Handler.FileSystem = wfs
	return wfs
}

func (fs *WebdavFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return fs.root.Mkdir(name, perm)
}

func (fs *WebdavFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return fs.root.OpenFile(name, flag, perm)
}

func (fs *WebdavFileSystem) RemoveAll(ctx context.Context, name string) error {
	return fs.root.RemoveAll(name)
}

func (fs *WebdavFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	return fs.root.Rename(oldName, newName)
}

func (fs *WebdavFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return fs.root.Stat(name)
}
