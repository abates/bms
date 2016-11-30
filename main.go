package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"net/http"
	"strconv"
	"strings"
)

var um *UserManager

func basicAuthHandler() gin.HandlerFunc {
	realm := "Basic realm=" + strconv.Quote("WebDav Realm")

	return func(c *gin.Context) {
		var err error
		var user *User
		if authString := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2); len(authString) == 2 {
			user, err = um.BasicAuthenticate(authString[1])
		}

		if user == nil || err != nil {
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.Set(gin.AuthUserKey, user)
		}
	}
}

func init() {
	um = NewUserManager(afero.NewBasePathFs(afero.NewOsFs(), "/Users/abates/bms"))
	if user, err := um.Add("user1", "1111"); err == nil {
		user.fs.Mkdir("/dir1", 0700)
		user.fs.Mkdir("/dir1/dir3", 0700)
		user.fs.Mkdir("/dir2", 0700)
		user.fs.Mkdir("/dir2/dir4", 0700)
		user.fs.Mkdir("/dir2/dir4/dir5", 0700)
	}

	if user, err := um.Add("user2", "2222"); err == nil {
		user.fs.Mkdir("dir6", 0700)
		user.fs.Mkdir("dir7", 0700)
		user.fs.Mkdir("dir6/dir8", 0700)
	}
}

func main() {
	r := gin.Default()

	r.Static("/js", "app/js")
	r.Static("/css", "app/css")
	r.StaticFile("/", "app/index.html")
	r.StaticFile("/index.html", "app/index.html")

	webdavRoute := r.Group("/webdav/", basicAuthHandler())
	for _, method := range []string{"OPTIONS", "GET", "HEAD", "POST", "DELETE", "PUT", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "PROPFIND", "PROPPATCH"} {
		webdavRoute.Handle(method, "*path", func(c *gin.Context) {
			u, _ := c.Get(gin.AuthUserKey)
			if user, ok := u.(*User); ok {
				wfs, err := NewWebdavFileSystem(user)
				if err == nil {
					wfs.ServeHTTP(c.Writer, c.Request)
				} else {
					c.AbortWithError(http.StatusInternalServerError, err)
				}
			} else {
				c.AbortWithStatus(http.StatusNotFound)
			}
		})
	}

	r.Run(":8080")
}
