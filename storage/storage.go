package storage

import (
	"time"
)

type Gob struct {
	UID     string
	Type    string
	Data    []byte
	Created time.Time
	IP      string
}

// A data structure to hold a key/value pair.
type UIDCreated struct {
	UID     string
	Created string
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type Horde []*UIDCreated
