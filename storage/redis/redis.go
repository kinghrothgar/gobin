package redis

import (
	"bytes"
	realgob "encoding/gob"
	"errors"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/storage"
	"github.com/mediocregopher/radix/redis"
	"time"
)

type RedisStore struct {
	Client *redis.Client
}

// gobKey forms the redis key from the uid
func gobKey(uid string) string {
	return "gob:" + uid
}

// hordekey forms the redis key from the horde name
func hordeKey(hordeName string) string {
	return "horde:" + hordeName
}

// gobEncode encodes a storage gob into a byte array
func gobEncode(gob *storage.Gob) ([]byte, error) {
	buf := new(bytes.Buffer)
	gobEnc := realgob.NewEncoder(buf)
	err := gobEnc.Encode(gob)
	return buf.Bytes(), err
}

// gobDecode decodes a byte array into a storage gob
func gobDecode(gobBytes []byte) (*storage.Gob, error) {
	gob := &storage.Gob{}
	buf := bytes.NewReader(gobBytes)
	gobDec := realgob.NewDecoder(buf)
	err := gobDec.Decode(gob)
	return gob, err
}

func (redisStore *RedisStore) UIDExist(uid string) (bool, error) {
	gslog.Debug("making redis get call")
	reply := redisStore.Client.Cmd("GET", gobKey(uid))
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
	gobBytes, err := gobEncode(gob)
	if err != nil {
		return err
	}
	reply := redisStore.Client.Cmd("SET", gobKey(gob.UID), gobBytes)
	return reply.Err
}

func (redisStore *RedisStore) GetGob(uid string) (*storage.Gob, error) {
	reply := redisStore.Client.Cmd("GET", gobKey(uid))
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
	return gobDecode(b)
}

func (redisStore *RedisStore) DelGob(uid string) error {
	//delete(redisStore.gobs, uid)
	return nil
}

func (redisStore *RedisStore) GetHorde(hordeName string) (storage.Horde, error) {
	reply := redisStore.Client.Cmd("ZRANGE", "horde:"+hordeName, 0, -1)
	if reply.Err != nil {
		return nil, reply.Err
	}
	h, err := reply.ListBytes()
	if err != nil {
		return nil, err
	}
	horde := make(storage.Horde, len(h))
	for i, bs := range h {
		uidCreated := &storage.UIDCreated{}
		buf := bytes.NewReader(bs)
		gobDec := realgob.NewDecoder(buf)
		if err = gobDec.Decode(uidCreated); err != nil {
			return nil, err
		}
		horde[i] = uidCreated
	}
	return horde, nil
}

func (redisStore *RedisStore) AddUIDHorde(hordeName string, uid string) error {
	now := time.Now()
	uidCreated := storage.UIDCreated{UID: uid, Created: now.String()}
	buf := new(bytes.Buffer)
	gobEnc := realgob.NewEncoder(buf)
	if err := gobEnc.Encode(uidCreated); err != nil {
		return err
	}
	reply := redisStore.Client.Cmd("ZADD", hordeKey(hordeName), now.UnixNano(), buf.Bytes())
	if reply.Err != nil {
		return reply.Err
	}
	reply = redisStore.Client.Cmd("SET", "uid.to.horde:"+uid, hordeName)
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
