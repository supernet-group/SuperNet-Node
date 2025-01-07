package dbutils

import (
	"SuperNet-Node/utils"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v4"
)

// type DB struct {
// 	db *badger.DB
// }

// func NewDB() (*DB, error) {
// 	opts := badger.DefaultOptions("/tmp/badger").WithLoggingLevel(badger.WARNING)
// 	db, err := badger.Open(opts)
// 	if err != nil {
// 		return nil, fmt.Errorf("> badger.Open: %v", err.Error())
// 	}
// 	return &DB{db: db}, nil
// }

// func Close(db *badger.DB) error {
// 	return db.Close()
// }

var (
	db     *badger.DB
	once   sync.Once
	dbOpen = false
)

func GetDB() *badger.DB {
	once.Do(func() {
		opts := badger.DefaultOptions("/tmp/badger").WithLoggingLevel(badger.WARNING)
		var err error
		db, err = badger.Open(opts)
		if err != nil {
			panic(err)
		}
		dbOpen = true
	})
	return db
}

func CloseDB() {
	if dbOpen {
		db.Close()
		dbOpen = false
	}
}

func Update(db *badger.DB, key, value []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func Get(db *badger.DB, key []byte) ([]byte, error) {
	var valCopy []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		valCopy, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return valCopy, nil
}

func Delete(db *badger.DB, key []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func GenToken(buyer string) (string, error) {
	mlToken, err := utils.GenerateRandomString(16)
	if err != nil {
		return "", fmt.Errorf("> GenerateRandomString: %v", err.Error())
	}

	db := GetDB()
	Update(db, []byte("buyer"), []byte(buyer))
	Update(db, []byte("token"), []byte(mlToken))
	return mlToken, nil
}
