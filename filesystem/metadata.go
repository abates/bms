package filesystem

import (
	"bytes"
	"github.com/abates/bms/database"
	"github.com/lunixbochs/struc"
	"io"
	"os"
	"time"
)

type Metadata struct {
	ID          database.ID `struc:"[16]byte"`
	Owner       database.ID `struc:"[16]byte"`
	Mode        os.FileMode `struc:"uint32"`
	ModTime     int64       `struc:"int64"`
	IsDirectory bool
	NameLen     int `struc:"int16,sizeof=Name"`
	Name        string
}

func NewMetadata(name string, mode os.FileMode) *Metadata {
	return &Metadata{
		ID:      database.NewID(),
		Name:    name,
		Mode:    mode,
		ModTime: time.Now().Unix(),
	}
}

func (ah *Metadata) SetName(name string) {
	ah.NameLen = len(name)
	ah.Name = name
}

func (ah *Metadata) UnmarshalBinary(data []byte) error {
	return ah.Unpack(bytes.NewBuffer(data))
}

func (ah *Metadata) Unpack(reader io.Reader) error {
	return struc.Unpack(reader, ah)
}

func (ah *Metadata) MarshalBinary() ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := ah.Pack(buffer)
	return buffer.Bytes(), err
}

func (ah *Metadata) Pack(writer io.Writer) error {
	return struc.Pack(writer, ah)
}
