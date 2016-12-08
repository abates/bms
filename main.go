package main

import (
	"fmt"
	"github.com/abates/bms/database"
	"github.com/abates/bms/filesystem"
	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"net/http"
	"strconv"
	"strings"
)

var (
	um                   *UserManager
	db                   database.Database
	ErrInvalidAuthString = fmt.Errorf("invalid authorization string")
)

func basicAuthHandler() gin.HandlerFunc {
	realm := "Basic realm=" + strconv.Quote("WebDav Realm")

	return func(c *gin.Context) {
		var user *User
		var err error
		authString := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)
		if len(authString) == 2 {
			user, err = um.BasicAuthenticate(authString[1])
		}

		if user == nil || err != nil {
			if err != nil {
				logger.Warnf("%v", err)
			}
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.Set(gin.AuthUserKey, user)
		}
	}
}

func init() {
	var err error
	db, err = database.OpenBoltDb("/Users/abates/bms/bolt.db")
	if err != nil {
		panic(err.Error())
	}
	filesystem.Db = db
	filesystem.BackendFs = afero.NewBasePathFs(afero.NewOsFs(), "/Users/abates/bms")
	um = NewUserManager()
	_, err = um.Add("user1", "1111")
	if err != nil {
		logger.Warn(err.Error())
	}

	_, err = um.Add("user2", "2222")
	if err != nil {
		logger.Warn(err.Error())
	}
}

func main() {
	r := gin.Default()

	r.Static("/js", "app/js")
	r.Static("/css", "app/css")
	r.StaticFile("/", "app/index.html")
	r.StaticFile("/index.html", "app/index.html")

	webdavRoute := r.Group("/webdav/", basicAuthHandler())
	handler := func(c *gin.Context) {
		u, _ := c.Get(gin.AuthUserKey)
		if user, ok := u.(*User); ok {
			rootFolder := &filesystem.Folder{}
			err := db.Find(user.RootFolderID, rootFolder)
			if err == nil {
				wfs := filesystem.NewWebdavFileSystem(user.ID, rootFolder)
				wfs.ServeHTTP(c.Writer, c.Request)
			} else {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
		} else {
			c.AbortWithStatus(http.StatusNotFound)
		}
	}
	for _, method := range []string{"OPTIONS", "GET", "HEAD", "POST", "DELETE", "PUT", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "PROPFIND", "PROPPATCH"} {
		webdavRoute.Handle(method, "*path", handler)
	}

	r.Run(":8080")
}
