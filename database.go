package main

import (
	"github.com/satori/go.uuid"
)

type ID uuid.UUID

func generateID() ID {
	return ID(uuid.NewV4())
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

type Database interface {
	Find(ID) Asset
	Save(Asset) error
}

type mapDatabase struct {
	backend map[string]Asset
}

func NewMapDatabase() Database {
	return &mapDatabase{
		backend: make(map[string]Asset),
	}
}

func (db *mapDatabase) Find(id ID) Asset {
	return db.backend[id.String()]
}

func (db *mapDatabase) Save(asset Asset) error {
	db.backend[asset.ID().String()] = asset
	return nil
}
