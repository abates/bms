package database

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

type boltDb struct {
	bolt *bolt.DB
}

func OpenBoltDb(filename string) (Database, error) {
	var err error
	db := &boltDb{}
	db.bolt, err = bolt.Open(filename, 0600, nil)
	if err == nil {
		err = db.bolt.Update(func(tx *bolt.Tx) (err error) {
			_, err = tx.CreateBucketIfNotExists([]byte("bms"))
			return
		})
	}
	return db, err
}

func (db *boltDb) Find(id ID, receiver Storable) error {
	return db.bolt.View(func(tx *bolt.Tx) error {
		err := ErrNotFound
		bucket := tx.Bucket([]byte("bms"))
		data := bucket.Get(id[:])
		if data != nil {
			err = receiver.UnmarshalBinary(data)
		}
		return err
	})
}

func (db *boltDb) Save(id ID, object Storable) error {
	logger.Infof("Saving %v %v", id, object)
	gob, err := object.MarshalBinary()
	if err == nil {
		err = db.bolt.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("bms"))
			return bucket.Put(id[:], gob)
		})
	}

	return err
}

func (db *boltDb) Delete(id ID) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("bms"))
		return bucket.Delete(id[:])
	})
}
