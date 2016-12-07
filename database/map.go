package database

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

func (db *mapDatabase) Save(id ID, object Storable) error {
	gob, err := object.MarshalBinary()
	if err == nil {
		db.backend[id.String()] = gob
	}
	return err
}

func (db *mapDatabase) Delete(id ID) error {
	delete(db.backend, id.String())
	return nil
}
