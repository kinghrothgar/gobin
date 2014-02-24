package store

import (
	"crypto/rand"
	"errors"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/storage"
	"github.com/kinghrothgar/goblin/storage/memory"
	"github.com/kinghrothgar/goblin/storage/redis"
	"net"
	"strings"
	"time"
)

const (
	ALPHA = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type DataStore interface {
	UIDExist(string) (bool, error)
	PutGob(*storage.Gob) error
	GetGob(string) (*storage.Gob, error)
	DelGob(string) error
	GetHorde(string) (storage.Horde, error)
	AddUIDHorde(string, string) error
	DelUIDHorde(string, string) error
	Initialize(string) error
}

type FileStore interface {
	UIDExist(string) (bool, error)
	PutGob(*storage.Gob) error
	GetGob(string) (*storage.Gob, error)
	DelGob(string) error
	GetHorde(string) (storage.Horde, error)
	AddToHorde(string, string) error
	DelFromHorde(string, string) error
	Initialize(string) error
}

var (
	dataStore DataStore
	uidLen    int
)

func GetGob(uid string) ([]byte, string, error) {
	gob, err := dataStore.GetGob(uid)
	if err != nil {
		return nil, "", err
	}
	return gob.Data, gob.Type, err
}

func PutGob(uid string, data []byte, ip net.IP) error {
	gob := &storage.Gob{
		UID:     uid,
		Data:    data,
		IP:      ip,
		Created: time.Now(),
	}
	return dataStore.PutGob(gob)
}

// Should I return a Horde or just a map?
// TODO: Have dataStore store the and return the hordes in a sorted list
func GetHorde(hordeName string) (storage.Horde, error) {
	horde, err := dataStore.GetHorde(hordeName)
	if err != nil {
		return nil, err
	}
	return horde, nil
}

func PutHordeGob(uid string, hordeName string, data []byte, ip net.IP) error {
	if err := PutGob(uid, data, ip); err != nil {
		return err
	}
	return dataStore.AddUIDHorde(hordeName, uid)
}

func Initialize(storeType string, confStr string, uidLength int) error {
	uidLen = uidLength
	switch strings.ToUpper(storeType) {
	case "MEMORY":
		dataStore = new(memory.MemoryStore)
	case "REDIS":
		dataStore = new(redis.RedisStore)
	default:
		return errors.New("invalid store type")
	}
	return dataStore.Initialize(confStr)
}

func GetNewUID() string {
	bytes := make([]byte, uidLen)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = ALPHA[b%byte(len(ALPHA))]
	}
	uid := string(bytes)
	gslog.Debug("checking if " + uid + " exists")
	if exist, _ := dataStore.UIDExist(uid); exist {
		gslog.Debug(uid + " exists")
		return GetNewUID()
	}
	return uid
}
