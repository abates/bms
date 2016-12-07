package database

import (
	"encoding"
	"fmt"
	"github.com/satori/go.uuid"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type ID []byte

func NewID() ID {
	id := uuid.NewV4()
	return ID(id[:])
}

func (id ID) String() string {
	i, _ := uuid.FromBytes(id)
	return i.String()
}

type Storable interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type Database interface {
	Find(id ID, receiver Storable) error
	Save(id ID, storable Storable) error
	Delete(ID) error
}
