package store

import (
	"crypto/rand"
	"errors"
	"github.com/kinghrothgar/goblin/storage"
	"github.com/grooveshark/golib/gslog"
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
	DelUIDExist(string) (bool, error)
	PutGob([]byte, *storage.GobInfo) error
	AppendGob(string, []byte) error
	GetGob(string) ([]byte, *storage.GobInfo, error)
	GetGobLen(string) (int, error)
	DelGob(string) error
	DelUIDToUID(string) (string, error)
	GetHorde(string) (storage.Horde, error)
	AddUIDHorde(string, string) error
	DelUIDHorde(string) error
	AddUIDFIFO(string, string) error
	CurrentUIDFIFO(string) (string, error)
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
	gob, gobInfo, err := dataStore.GetGob(uid)
	if err != nil || gob == nil {
		return []byte{}, "", err
	}
	return gob.Data, gobInfo.Type, err
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

func AppendGob(uid string, data []byte) error {
	return nil
}

// TODO: looking up length of 
func PutFIFOGob(fifoName string, data []byte, ip string) (error) {
	newUIDFlag := false
	uid, err := dataStore.CurrentUIDFIFO(fifoName)
	if err != nil {
		return err
	}
	if uid == "" {
		uid = GetNewUID()
		newUIDFlag = true
	}
	// TODO: Doing an extra redis hit when not need if already a new UID
	if length, err := dataStore.GetGobLen(uid); err != nil {
		return err
	// If length + data is over 10MB, store in new gob
	} else if (length + len(data)) > 10485760 {
		uid = GetNewUID()
		newUIDFlag = true
	}

	if err := AppendGob(uid, data); err != nil {
		return err
	}
	if newUIDFlag {
		return dataStore.AddUIDFIFO(fifoName, uid)
	}
	return nil
}

func DelUIDToUID(delUID string) (string, error) {
	return dataStore.DelUIDToUID(delUID)
}

func DelGob(uid string) error {
	if err := dataStore.DelGob(uid); err != nil {
		return err
	}
	return dataStore.DelUIDHorde(uid)
}

func Initialize(storeType string, confStr string, uidLength int, delUIDLength int) error {
	uidLen = uidLength
	delUIDLen = delUIDLength
	switch strings.ToUpper(storeType) {
	//case "MEMORY":
	//	dataStore = memory.New(confStr)
	//	return nil
	case "REDIS":
		dataStore = redis.New(confStr)
		return nil
	default:
		gslog.Debug("STORE: initialized with store type: %s, conf string: %s, uid length: %d, del uid length: %d", storeType, confStr, uidLen, delUIDLen)
	}
	return errors.New("invalid store type")
}

func Configure(confStr string, uidLength int, delUIDLength int) {
	// TODO: Are these thread safe?
	uidLen = uidLength
	delUIDLen = delUIDLength
	dataStore.Configure(confStr)
	gslog.Debug("STORE: configured with conf string: %s, uid length: %d, del uid length: %d", confStr, uidLen, delUIDLen)
}

func randomString(length int) string {
	bytes := make([]byte, length)
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
// TODO: Not thread safe
func GetNewUID() string {
	uid := randomString(uidLen)
	if exist, _ := dataStore.UIDExist(uid); exist {
		return GetNewUID()
	}
	return uid
}
