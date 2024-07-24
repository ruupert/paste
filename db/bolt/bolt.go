package bolt

import (
	"fmt"
	"log"

	pastedb "github.com/ruupert/paste/db"
	bolt "go.etcd.io/bbolt"
)

type Bolton struct {
	bdb *bolt.DB
}

var singleton *Bolton

type Backend interface {
	Get(h string) (string, error)
	Put(p pastedb.PasteRecord) (string, error)
}

func Get(s string) (string, error) {
	bolton := GetBoltInstance()
	var res string
	//fmt.Println(uri)
	berr := bolton.bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("paste"))
		v := b.Get([]byte(s))
		res = string(v)
		return nil
	})
	if berr != nil {
		return "", berr
	}
	return res, nil
}

func Put(s pastedb.PasteRecord) error {
	bolton := GetBoltInstance()
	berr := bolton.bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("paste"))
		err := b.Put([]byte(s.Hash), []byte(s.Body))
		return err
	})
	if berr != nil {
		fmt.Println(berr)
		return berr
	}
	return nil

}

func init() {
	fmt.Println("Initializing")
	mbdb, err := bolt.Open("bolt.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	mbdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("paste"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	singleton = &Bolton{
		bdb: mbdb,
	}
}

func GetBoltInstance() *Bolton {
	return singleton
}
