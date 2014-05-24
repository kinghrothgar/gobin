package storage

import (
	"time"
)

const (
	GOB_INFO_VERSION int = 1
)

type GobInfo struct {
	UID     string
	Token   string
	Type    string
	Created time.Time
	IP      string
	Version int
}

// A data structure to hold a key/value pair.
type UIDCreated struct {
	UID     string
	Created string
}

// A slice of Pairs (UID string and Created time string)
type Horde []*UIDCreated
