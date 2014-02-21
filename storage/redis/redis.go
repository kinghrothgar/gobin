package redis

import (
	"errors"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/storage"
	"github.com/mediocregopher/radix/redis"
	"time"
)

type Gobs map[string]*storage.Gob
type Hordes map[string]storage.Horde
type UIDToHorde map[string]string

type RedisStore struct {
	Client *redis.Client
}

func (redisStore *RedisStore) UIDExist(uid string) (bool, error) {
	if _, ok := redisStore.gobs[uid]; ok {
		return true, nil
	}
	return false, nil
}

func (redisStore *RedisStore) PutGob(gob *storage.Gob) error {
	redisStore.gobs[gob.UID] = gob
	return nil
}

func (redisStore *RedisStore) GetGob(uid string) (*storage.Gob, error) {
	if gob, ok := redisStore.gobs[uid]; ok {
		return gob, nil
	}
	return nil, errors.New("uid does not exist")
}

func (redisStore *RedisStore) DelGob(uid string) error {
	// TODO: do I need to check?
	if exist, _ := redisStore.UIDExist(uid); !exist {
		return errors.New("uid does not exist")
	}
	delete(redisStore.gobs, uid)
	return nil
}

func (redisStore *RedisStore) GetHorde(hordeName string) (storage.Horde, error) {
	if horde, ok := redisStore.hordes[hordeName]; ok {
		gslog.Debug("%+v", horde)
		return horde, nil
	}
	return storage.Horde{}, nil
}

func (redisStore *RedisStore) AddUIDHorde(hordeName string, uid string) error {
	// TODO: do I need to do this
	if horde, ok := redisStore.hordes[hordeName]; ok {
		horde[uid] = time.Now()
	} else {
		redisStore.hordes[hordeName] = storage.Horde{uid: time.Now()}
	}
	redisStore.uidToHorde[uid] = hordeName
	return nil
}

func (redisStore *RedisStore) DelUIDHorde(hordeName string, uid string) error {
	// TODO: should I even be checking if I'm really deleting?
	horde, ok := redisStore.hordes[hordeName]
	if !ok {
		return errors.New("horde does not exist")
	}
	if _, ok = horde[uid]; !ok {
		return errors.New("uid does not exist in horde")
	}
	delete(horde, uid)
	delete(redisStore.uidToHorde, uid)
	return nil
}

func (redisStore *RedisStore) Initialize(confStr string) error {
	var err error
	if r, err = redis.Dial("tcp", "127.0.0.1:6379"); err != nil {
		return err
	}
	return nil
}
