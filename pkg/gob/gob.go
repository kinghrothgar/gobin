package gob

import (
	"math/rand"
	"net/http"
	"time"
)

// Metadata for a gob
// TODO I don't really want ID, AuthKey, and ContentType to be exported, but then I can't use them in sqlx
type Metadata struct {
	ID          string    `db:"id"`
	AuthKey     string    `db:"auth_key"`
	Encrypted   bool      `db:"encrypted"`
	CreateDate  time.Time `db:"create_date"`
	ExpireDate  time.Time `db:"expire_date"`
	Size        int64     `db:"size"`
	OwnerID     int       `db:"owner_id"`
	ContentType string    `db:"content_type"`
}

const (
	// IDLen length of gob id string
	IDLen = 6
	// AuthKeyLen length of gob authKey string
	AuthKeyLen = 16
	// LegibleAlphanumeric is a string containing all alphanumeric characters except for ones that fonts can make indescernable: O, 0, l, 1
	LegibleAlphanumeric = "ABCDEFGHIJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
)

var legibleRunes = []rune(LegibleAlphanumeric)

// TODO does this create a uniform distrobution?
// TODO set seed?
// TODO where should this go?
func randomReadableString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = legibleRunes[rand.Intn(len(legibleRunes))]
	}
	return string(b)
}

// NewMetadata returns new *Metadata instance
func NewMetadata() *Metadata {
	id := randomReadableString(IDLen)
	authKey := randomReadableString(AuthKeyLen)
	return &Metadata{
		ID:         id,
		AuthKey:    authKey,
		CreateDate: time.Now(),
	}
}

// SetContentType uses http.DetectContentType to set the contentType
// It considers at most the first 512 bytes of data.
func (g *Metadata) SetContentType(data []byte) string {
	g.ContentType = http.DetectContentType(data)
	return g.ContentType
}
