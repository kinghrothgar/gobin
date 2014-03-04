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
	DelUIDExist(string) (bool, error)
	PutGob(*storage.Gob) error
	GetGob(string) (*storage.Gob, error)
	DelGob(string) error
	DelUIDToUID(string) (string, error)
	UIDToHorde(string) (string, error)
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
	delUIDLen int
)

func GetGob(uid string) ([]byte, string, error) {
	gob, err := dataStore.GetGob(uid)
	if err != nil || gob == nil {
		return []byte{}, "", err
	}
	return gob.Data, gob.Type, err
}

func PutGob(data []byte, ip string) (string, string, error) {
	uid := GetNewUID()
	delUID := GetNewDelUID()
	gob := &storage.Gob{
		UID:     uid,
		DelUID:  delUID,
		Data:    data,
		IP:      ip,
		Created: time.Now(),
	}
	err := dataStore.PutGob(gob)
	return uid, delUID, err
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

func PutHordeGob(hordeName string, data []byte, ip string) (string, string, error) {
	uid, delUID, err := PutGob(data, ip)
	if err != nil {
		return uid, delUID, err
	}
	return uid, delUID, dataStore.AddUIDHorde(hordeName, uid)
}

func DelUIDToUID(delUID string) (string, error) {
	return dataStore.DelUIDToUID(delUID)
}

func DelGob(uid string) error {
}

func Initialize(storeType string, confStr string, uidLength int, delUIDLength int) error {
	uidLen = uidLength
	delUIDLen = delUIDLength
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

func Configure(confStr string, uidLength int, delUIDLength int) {
	// TODO: Are these thread safe?
	uidLen = uidLength
	delUIDLen = delUIDLength
	dataStore.Configure(confStr)
}

func randomString(length int) string {
	bytes := make([]byte, uidLen)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = ALPHA[b%byte(len(ALPHA))]
	}
	return string(bytes)
}

func GetNewDelUID() string {
	delUID := randomString(delUIDLen)
	if exist, _ := dataStore.DelUIDExist(delUID); exist {
		return GetNewDelUID()
	}
	return delUID
}

// TODO: Need to error
func GetNewUID() string {
	uid := randomString(uidLen)
	if exist, _ := dataStore.UIDExist(uid); exist {
		return GetNewUID()
	}
	return uid
}
