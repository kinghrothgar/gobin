package store

import (
	"crypto/rand"
	"errors"
	"github.com/kinghrothgar/goblin/storage"
	"github.com/kinghrothgar/goblin/storage/memory"
	"github.com/kinghrothgar/goblin/storage/redis"
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
	Configure(string)
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
	if err != nil || gob == nil {
		return []byte{}, "", err
	}
	return gob.Data, gob.Type, err
}

func PutGob(uid string, data []byte, ip string) error {
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

func PutHordeGob(uid string, hordeName string, data []byte, ip string) error {
	if err := PutGob(uid, data, ip); err != nil {
		return err
	}
	return dataStore.AddUIDHorde(hordeName, uid)
}

func Initialize(storeType string, confStr string, uidLength int) error {
	uidLen = uidLength
	switch strings.ToUpper(storeType) {
	case "MEMORY":
		dataStore = memory.New(confStr)
		return nil
	case "REDIS":
		dataStore = redis.New(confStr)
		return nil
	default:
	}
	return errors.New("invalid store type")
}

func Configure(confStr string, uidLength int) {
	dataStore.Configure(confStr)
}

func GetNewUID() string {
	bytes := make([]byte, uidLen)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = ALPHA[b%byte(len(ALPHA))]
	}
	uid := string(bytes)
	if exist, _ := dataStore.UIDExist(uid); exist {
		return GetNewUID()
	}
	return uid
}
