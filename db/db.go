package pastedb

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	hash "hash/fnv"
	"log"
	"strings"

	bolt "go.etcd.io/bbolt"
)

const Size = 5

type DatabaseType int

const (
	BoltDatabaseType DatabaseType = iota
)

type DatabaseInterface interface {
	Get(h string) (string, error)
	Put(p PasteRecord) (string, error)
}

type BoltDatabase struct {
	conn *bolt.DB
}

func (b *BoltDatabase) Get(h string) (string, error) {
	var res string
	berr := b.conn.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("paste"))
		v := b.Get([]byte(h))
		res = string(v)
		return nil
	})
	if berr != nil {
		return "", berr
	}
	return res, nil
}
func (b *BoltDatabase) Put(p PasteRecord) (string, error) {
	berr := b.conn.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("paste"))
		err := b.Put([]byte(p.Hash), []byte(p.Body))
		return err
	})
	if berr != nil {
		fmt.Println(berr)
		return "", berr
	}
	return p.Hash, nil

}

func NewDatabaseType(dbType DatabaseType) (DatabaseInterface, error) {
	switch dbType {
	case BoltDatabaseType:
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
		return &BoltDatabase{conn: mbdb}, nil
	default:
		return nil, errors.New("unsupported payment gateway type")
	}
}

type PasteRecord struct {
	Hash string `db:"hash"`
	Body string `db:"body"`
}

func (c *PasteRecord) New(body string) {
	c.Body = body
	c.Hash = c.digest(body)
}

func (c *PasteRecord) digest(body string) string {
	hh := hash.New32a()
	hh.Write([]byte(body))
	enc := b64.URLEncoding.Strict().EncodeToString(hh.Sum(nil))
	return strings.TrimSuffix(enc, "==")
}
