package pastedb

import (
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
	Get(h string) (string, error)
	Put(p PasteRecord) (string, error)
	GetName() string
}

// BoltStore
type BoltDatabase struct {
	name string
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
func (b *BoltDatabase) GetName() string {
	return b.name
}

// MemStore quickly, not fast...
type MemoryDatabase struct {
	name string
	sync.RWMutex
	Pastes []PasteRecord
}

var ErrHashNotFound = errors.New("hash not found")

func (b *MemoryDatabase) Get(h string) (string, error) {
	if len(b.Pastes) == 0 {
		return "", ErrHashNotFound
	}
	for _, v := range b.Pastes {
		if h == v.Hash {
			return v.Body, nil
		}
	}
	return "", ErrHashNotFound
}

func (b *MemoryDatabase) Put(p PasteRecord) (string, error) {
	b.RWMutex.Lock()
	defer b.RWMutex.Unlock()
	if len(b.Pastes) == 0 {
		b.Pastes = append(b.Pastes, p)
		return p.Hash, nil
	}
	for _, v := range b.Pastes {
		if v.Hash == p.Hash {
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
		mbdb.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("paste"))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
			return nil
		})
		return &BoltDatabase{name: "Bolt", conn: mbdb}, nil
	case MemoryDatabaseType:
		return &MemoryDatabase{name: "Memory"}, nil
	default:
		return nil, errors.New("unsupported type")
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
