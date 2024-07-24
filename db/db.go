package pastedb

import (
	b64 "encoding/base64"
	hash "hash/fnv"
	"strings"
)

const Size = 5

type PasteRecord struct {
	Hash string `db:"hash"`
	Body string `db:"body"`
}

func (c *PasteRecord) New(body string) {
	c.Body = body
	c.Hash = c.Digest(body)
}

func (c *PasteRecord) Digest(body string) string {
	hh := hash.New32a()
	hh.Write([]byte(body))
	enc := b64.URLEncoding.Strict().EncodeToString(hh.Sum(nil))
	return strings.TrimSuffix(enc, "==")
}
