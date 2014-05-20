package storage

import (
	"time"
)

type GobInfo struct {
	UID     string
	DelUID  string
	Type    string
	Created time.Time
	IP      string
}

// A data structure to hold a key/value pair.
type UIDCreated struct {
	UID     string
	Created string
}

// A slice of Pairs (UID string and Created time string)
type Horde []*UIDCreated
