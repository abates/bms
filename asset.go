package main

import (
	"io"
	"os"
)

type Asset interface {
	io.ReadWriteCloser
	io.Seeker
	Name() string
	Readdir(count int) ([]os.FileInfo, error)
	Stat() (os.FileInfo, error)
}
