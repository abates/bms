package database

import (
	"encoding"
	"fmt"
	"github.com/satori/go.uuid"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type ID uuid.UUID

func NewID() ID {
	return ID(uuid.NewV4())
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

type Storable interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	ID() ID
}

type Database interface {
	Find(id ID, receiver Storable) error
	Save(Storable) error
	Delete(ID) error
}

type mapDatabase struct {
	backend map[string][]byte
}

func NewMapDatabase() Database {
	return &mapDatabase{
		backend: make(map[string][]byte),
	}
}

func (db *mapDatabase) Find(id ID, receiver Storable) error {
	err := ErrNotFound
	if data, found := db.backend[id.String()]; found {
		err = receiver.UnmarshalBinary(data)
	}
	return err
}

func (db *mapDatabase) Save(object Storable) error {
	gob, err := object.MarshalBinary()
	if err == nil {
		db.backend[object.ID().String()] = gob
	}
	return err
}

func (db *mapDatabase) Delete(id ID) error {
	delete(db.backend, id.String())
	return nil
}
