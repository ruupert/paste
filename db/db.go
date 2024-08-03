//go:build !wasm
// +build !wasm

package pastedb

import (
	"bytes"
	b64 "encoding/base64"
	"errors"
	"fmt"
	hash "hash/fnv"
	"log"
	"strings"
	"sync"

	bolt "go.etcd.io/bbolt"
)

const Size = 5

type DatabaseType int

const (
	BoltDatabaseType DatabaseType = iota
	MemoryDatabaseType
)

type DatabaseInterface interface {
	Get(h []byte) ([]byte, error)
	Put(p PasteRecord) ([]byte, error)
	GetName() string
}

// BoltStore
type BoltDatabase struct {
	name string
	conn *bolt.DB
}

var ErrHashNotFound = errors.New("hash not found")

func (b *BoltDatabase) Get(h []byte) ([]byte, error) {
	var res []byte
	err := b.conn.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("paste"))
		v := b.Get(h)
		res = v
		return nil
	})
	if err != nil {
		return []byte(""), err
	}
	if bytes.Equal(res, []byte("")) {
		return []byte(""), ErrHashNotFound
	}
	return []byte(res), nil
}

func (b *BoltDatabase) Put(p PasteRecord) ([]byte, error) {
	err := b.conn.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("paste"))
		err := b.Put([]byte(p.Hash), []byte(p.Body))
		return err
	})
	if err != nil {
		return []byte(""), err
	}
	return p.Hash, nil

}
func (b *BoltDatabase) GetName() string {
	return b.name
}

// MemStore quickly, not fast...
type MemoryDatabase struct {
	name string
	sync.RWMutex
	Pastes []PasteRecord
}

func (b *MemoryDatabase) Get(h []byte) ([]byte, error) {
	if len(b.Pastes) == 0 {
		return []byte(""), ErrHashNotFound
	}
	for _, v := range b.Pastes {
		if bytes.Equal(h, v.Hash) {
			return v.Body, nil
		}
	}
	return []byte(""), ErrHashNotFound
}

func (b *MemoryDatabase) Put(p PasteRecord) ([]byte, error) {
	b.RWMutex.Lock()
	defer b.RWMutex.Unlock()
	if len(b.Pastes) == 0 {
		b.Pastes = append(b.Pastes, p)
		return p.Hash, nil
	}
	for _, v := range b.Pastes {
		if bytes.Equal(v.Hash, p.Hash) {
			return v.Hash, nil
		}
	}
	b.Pastes = append(b.Pastes, p)
	return p.Hash, nil
}
func (b *MemoryDatabase) GetName() string {
	return b.name
}

func NewDatabaseType(dbType DatabaseType) (DatabaseInterface, error) {
	switch dbType {
	case BoltDatabaseType:
		mbdb, err := bolt.Open("bolt.db", 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		err = mbdb.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("paste"))
			if err != nil {
				if err != bolt.ErrBucketExists {
					log.Fatal(err)
					return fmt.Errorf("create bucket: %s", err)
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
		return &BoltDatabase{name: "Bolt", conn: mbdb}, nil
	case MemoryDatabaseType:
		return &MemoryDatabase{name: "Memory"}, nil
	default:
		return nil, errors.New("unsupported type")
	}
}

type PasteRecord struct {
	Hash []byte `db:"hash"`
	Body []byte `db:"body"`
}

func (c *PasteRecord) New(body []byte) {
	c.Body = body
	c.Hash = c.digest(body)
}

func (c *PasteRecord) digest(body []byte) []byte {
	hh := hash.New32a()
	_, err := hh.Write(body)
	if err != nil {
		fmt.Println(err)
	}
	enc := b64.URLEncoding.Strict().EncodeToString(hh.Sum(nil))
	return []byte(strings.TrimSuffix(enc, "=="))
}
