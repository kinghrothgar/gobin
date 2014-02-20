package storage

import (
	"net"
	"time"
)

type Gob struct {
	UID     string
	Type    string
	Data    []byte
	Created time.Time
	IP      net.IP
}

type Horde map[string]time.Time
