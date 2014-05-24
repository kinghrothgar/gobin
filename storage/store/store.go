package store

import (
	"crypto/rand"
	"errors"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/storage"
	//"github.com/kinghrothgar/goblin/storage/memory"
	"github.com/kinghrothgar/goblin/storage/redis"
	"strings"
	"time"
)

const (
	ALPHA = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type DataStore interface {
	UIDExist(string) (bool, error)
	TokenExist(string) (bool, error)
	PutGob([]byte, *storage.GobInfo) error
	AppendGob(string, []byte) error
	GetGob(string) ([]byte, *storage.GobInfo, error)
	GetGobLen(string) (int, error)
	DelGob(string) error
	TokenToUID(string) (string, error)
	GetHorde(string) (storage.Horde, error)
	AddUIDHorde(string, string) error
	DelUIDHorde(string) error
	Configure(string)
}

type FileStore interface {
	UIDExist(string) (bool, error)
	PutGob(*storage.GobInfo) error
	GetGob(string) (*storage.GobInfo, error)
	DelGob(string) error
	GetHorde(string) (storage.Horde, error)
	AddToHorde(string, string) error
	DelFromHorde(string, string) error
	Initialize(string) error
}

var (
	dataStore DataStore
	uidLen    int
	tokenLen  int
)

func GetGob(uid string) ([]byte, string, error) {
	gob, gobInfo, err := dataStore.GetGob(uid)
	if err != nil || gob == nil {
		return []byte{}, "", err
	}
	return gob, gobInfo.Type, err
}

func PutGob(data []byte, ip string) (string, string, error) {
	uid := GetNewUID()
	token := GetNewToken()
	t := time.Now()
	gobInfo := &storage.GobInfo{
		UID:     uid,
		Token:   token,
		IP:      ip,
		Created: t,
		Version: storage.GOB_INFO_VERSION,
	}
	err := dataStore.PutGob(data, gobInfo)
	return uid, token, err
}

func AppendGob(uid string, data []byte) error {
	return dataStore.AppendGob(uid, data)
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

func TokenToUID(token string) (string, error) {
	return dataStore.TokenToUID(token)
}

func DelGob(uid string) error {
	if err := dataStore.DelGob(uid); err != nil {
		return err
	}
	return dataStore.DelUIDHorde(uid)
}

func Initialize(storeType string, confStr string, uidLength int, tokenLength int) error {
	uidLen = uidLength
	tokenLen = tokenLength
	switch strings.ToUpper(storeType) {
	//case "MEMORY":
	//	dataStore = memory.New(confStr)
	//	return nil
	case "REDIS":
		dataStore = redis.New(confStr)
		return nil
	default:
		gslog.Debug("STORE: initialized with store type: %s, conf string: %s, uid length: %d, token length: %d", storeType, confStr, uidLen, tokenLen)
	}
	return errors.New("invalid store type")
}

func Configure(confStr string, uidLength int, tokenLength int) {
	// TODO: Are these thread safe?
	uidLen = uidLength
	tokenLen = tokenLength
	dataStore.Configure(confStr)
	gslog.Debug("STORE: configured with conf string: %s, uid length: %d, token length: %d", confStr, uidLen, tokenLen)
}

func randomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = ALPHA[b%byte(len(ALPHA))]
	}
	return string(bytes)
}

func GetNewToken() string {
	token := randomString(tokenLen)
	if exist, _ := dataStore.TokenExist(token); exist {
		return GetNewToken()
	}
	return token
}

// TODO: Need to error
// TODO: Not thread safe
func GetNewUID() string {
	uid := randomString(uidLen)
	if exist, _ := dataStore.UIDExist(uid); exist {
		return GetNewUID()
	}
	return uid
}
