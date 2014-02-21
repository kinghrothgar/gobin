package redis

import (
	"bytes"
	"errors"
	"github.com/kinghrothgar/goblin/storage"
	"github.com/mediocregopher/radix/redis"
	"github.com/grooveshark/golib/gslog"
	realgob "encoding/gob"
	"time"
)

type Gobs map[string]*storage.Gob
type Hordes map[string]storage.Horde
type UIDToHorde map[string]string

type RedisStore struct {
	Client *redis.Client
}

func (redisStore *RedisStore) UIDExist(uid string) (bool, error) {
	gslog.Debug("making redis get call")
	reply := redisStore.Client.Cmd("GET", "gob:" + uid)
	gslog.Debug("made redis get call")
	if reply.Err != nil {
		return false, reply.Err
	}
	if reply.Type == redis.NilReply {
		return false, nil
	}
	return true, nil
}

func (redisStore *RedisStore) PutGob(gob *storage.Gob) error {
	buf := new(bytes.Buffer)
	gobEnc := realgob.NewEncoder(buf)
	if err := gobEnc.Encode(gob); err != nil {
		return err
	}
	reply := redisStore.Client.Cmd("SET", "gob:" + gob.UID, buf.Bytes())
	return reply.Err
}

func (redisStore *RedisStore) GetGob(uid string) (*storage.Gob, error) {
	reply := redisStore.Client.Cmd("GET", "gob:" + uid)
	if reply.Err != nil {
		return nil, reply.Err
	}
	if reply.Type == redis.NilReply {
		return nil, errors.New("uid does not exist")
	}
	b, err := reply.Bytes()
	if err != nil {
		return nil, err
	}
	gob := &storage.Gob{}
	buf := bytes.NewReader(b)
	gobDec := realgob.NewDecoder(buf)
	if err = gobDec.Decode(gob); err != nil {
		return nil, err
	}
	return gob, nil
}

func (redisStore *RedisStore) DelGob(uid string) error {
	// TODO: do I need to check?
	if exist, _ := redisStore.UIDExist(uid); !exist {
		return errors.New("uid does not exist")
	}
	//delete(redisStore.gobs, uid)
	return nil
}

func (redisStore *RedisStore) GetHorde(hordeName string) (storage.Horde, error) {
	reply := redisStore.Client.Cmd("HGETALL", "horde:" + hordeName)
	if reply.Err != nil {
		return nil, reply.Err
	}
	if reply.Type == redis.NilReply {
		return storage.Horde{}, nil
	}
	h, err := reply.Hash()
	if err != nil {
		return nil, err
	}
	horde := storage.Horde{}
	for uid, created := range h {
		t := time.Time{}
		t.GobDecode([]byte(created))
		horde[uid] = t
	}
	return horde, nil
}

func (redisStore *RedisStore) AddUIDHorde(hordeName string, uid string) error {
	createdBytes, _ := time.Now().GobEncode()
	reply := redisStore.Client.Cmd("HMSET", "horde:" + hordeName, uid, string(createdBytes))
	if reply.Err != nil {
		return reply.Err
	}
	reply = redisStore.Client.Cmd("SET", "uid.to.horde:" + uid, hordeName)
	return reply.Err
}

func (redisStore *RedisStore) DelUIDHorde(hordeName string, uid string) error {
	// TODO: should I even be checking if I'm really deleting?
	//horde, ok := redisStore.hordes[hordeName]
	//if !ok {
	//	return errors.New("horde does not exist")
	//}
	//if _, ok = horde[uid]; !ok {
	//	return errors.New("uid does not exist in horde")
	//}
	//delete(horde, uid)
	//delete(redisStore.uidToHorde, uid)
	return nil
}

func (redisStore *RedisStore) Initialize(confStr string) error {
	var err error
	redisStore.Client, err = redis.Dial("tcp", "127.0.0.1:6666")
	return err
}
