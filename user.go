package main

import (
	"encoding/base64"
	"fmt"
	"github.com/abates/bms/database"
	"github.com/spf13/afero"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type User struct {
	ID       database.ID
	Username string
	Password string
	fs       FileSystem
}

func (u *User) Authenticate(password string) (err error) {
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return
}

func (u *User) ChangePassword(oldPassword, newPassword string) (err error) {
	err = u.Authenticate(oldPassword)
	if err == nil {
		u.SetPassword(newPassword)
	}
	return
}

func (u *User) SetPassword(password string) (err error) {
	var p []byte
	p, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err == nil {
		u.Password = string(p)
	}
	return
}

type UserManager struct {
	db map[string]*User
	fs afero.Fs
}

func NewUserManager() *UserManager {
	return &UserManager{
		db: make(map[string]*User),
	}
}

func (um *UserManager) Add(username, password string) (user *User, err error) {
	if _, found := um.db[username]; found {
		err = fmt.Errorf("%s already exists", username)
	} else {
		user = &User{
			Username: username,
		}
		user.fs = NewFolderFileSystem(user)
		user.SetPassword(password)
		um.db[username] = user
	}
	return
}

func (um *UserManager) Find(username string) (user *User, err error) {
	if u, found := um.db[username]; found {
		user = u
	} else {
		err = fmt.Errorf("Could not find %s in the database", username)
	}
	return
}

func (um *UserManager) Authenticate(username, password string) (user *User, err error) {
	user, err = um.Find(username)
	if err == nil {
		err = user.Authenticate(password)
	}
	return
}

func (um *UserManager) BasicAuthenticate(authString string) (user *User, err error) {
	var s []byte
	s, err = base64.StdEncoding.DecodeString(authString)
	if err == nil {
		authPair := strings.SplitN(string(s), ":", 2)
		if len(authPair) != 2 {
			err = fmt.Errorf("Invalid authentication string")
		} else {
			user, err = um.Authenticate(authPair[0], authPair[1])
		}
	}
	return
}
