package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/abates/bms/database"
	"github.com/lunixbochs/struc"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type User struct {
	ID           database.ID `struc:"[16]byte"`
	UsernameLen  int         `struc:"uint16,sizeof=Username"`
	Username     string      `struc:string`
	PasswordLen  int         `struc:"uint16,sizeof=Password"`
	Password     string      `struc:string`
	RootFolderID database.ID `struc:"[16]byte"`
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

func (u *User) MarshalBinary() ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := struc.Pack(buffer, u)
	logger.Infof("Marshal:\n%s", hex.Dump(buffer.Bytes()))
	return buffer.Bytes(), err
}

func (u *User) SetPassword(password string) (err error) {
	var p []byte
	p, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err == nil {
		u.Password = string(p)
	}
	return
}

func (u *User) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewBuffer(data)
	err := struc.Unpack(buffer, u)
	return err
}

type UserManager struct {
}

func NewUserManager() *UserManager {
	return &UserManager{}
}

func (um *UserManager) Add(username, password string) (*User, error) {
	user := &User{}
	err := db.Find(database.ID(username), user)
	if err == nil {
		err = fmt.Errorf("%s already exists", username)
	} else if err == database.ErrNotFound {
		user.ID = database.NewID()
		user.Username = username
		user.SetPassword(password)
		rootFolder := NewFolder("", 0700)
		user.RootFolderID = rootFolder.ID()
		err = db.Save(rootFolder.ID(), rootFolder)
		if err == nil {
			err = db.Save(database.ID(username), user)
		}
	}

	return user, err
}

func (um *UserManager) Find(username string) (*User, error) {
	user := &User{}
	err := db.Find(database.ID(username), user)
	if err == database.ErrNotFound {
		err = fmt.Errorf("Could not find %s in the database", username)
	}
	return user, err
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
