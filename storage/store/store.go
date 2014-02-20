package store

import (
	"crypto/rand"
	"errors"
	"github.com/kinghrothgar/goblin/storage"
	"github.com/kinghrothgar/goblin/storage/memory"
	"net"
	"net/http"
	"regexp"
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
	AddToHorde(string, string) error
	DelFromHorde(string, string) error
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
	alphaReg  = regexp.MustCompile("^[A-Za-z]+$")
)

func GetGob(uid string) ([]byte, string, error) {
	if len(uid) > uidLen || !alphaReg.MatchString(uid) {
		return nil, "", errors.New("invalid uid")
	}
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
		Type:    http.DetectContentType(data),
		IP:      ip,
		Created: time.Now(),
	}
	return dataStore.PutGob(gob)
}

func Initialize(storeType string, confStr string, uidLength int) error {
	uidLen = uidLength
	switch strings.ToUpper(storeType) {
	case "MEMORY":
		dataStore = new(memory.MemoryStore)
	default:
		return errors.New("invalid store type")
	}
	return dataStore.Initialize(confStr)
}

func GetRandUID() string {
	bytes := make([]byte, uidLen)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = ALPHA[b%byte(len(ALPHA))]
	}
	uid := string(bytes)
	if exist, _ := dataStore.UIDExist(uid); exist {
		return GetRandUID()
	}
	return uid
}
