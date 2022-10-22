package main

import (
	"github.com/jakemakesstuff/spherical/hashverifier"
	bolt "go.etcd.io/bbolt"
)

type boltDriver struct {
	db *bolt.DB
}

var pgp = []byte("pgp")

func (b boltDriver) LookupPGP(hostname string) []byte {
	var cpy []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(pgp).Get([]byte(hostname))
		if value != nil {
			cpy = make([]byte, len(value))
			copy(cpy, value)
		}
		return nil
	})
	if err != nil {
		return nil
	}

	return cpy
}

func (b boltDriver) WritePGP(hostname string, value []byte) {
	_ = b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(pgp).Put([]byte(hostname), value)
	})
}

var hashBucket = []byte("hash")

func (b boltDriver) Exists(hash []byte) bool {
	exists := false
	_ = b.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(hashBucket).Get(hash)
		exists = v != nil
		return nil
	})
	return exists
}

func (b boltDriver) Ensure(hash []byte) {
	_ = b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(hashBucket).Put(hash, hashBucket)
	})
}

var _ hashverifier.HashCache = boltDriver{}

func newBolt(path string) (hashverifier.HashCache, error) {
	db, err := bolt.Open(path, 0o666, nil)
	if err != nil {
		return nil, err
	}

	return boltDriver{db: db}, nil
}
