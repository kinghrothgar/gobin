package redis

import (
	"bytes"
	realgob "encoding/gob"
	"github.com/kinghrothgar/goblin/storage"
	"github.com/mediocregopher/radix/redis"
	"github.com/kinghrothgar/goblin/storage/redis/pool"
	"time"
)

type RedisStore struct {
	*pool.Pool
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

// New returns a new RedisStore
func New(confStr string) *RedisStore {
	return &RedisStore{pool.New("tcp", confStr, 50)}
}


func (redisStore *RedisStore) UIDExist(uid string) (bool, error) {
	client, err := redisStore.Get()
	if err != nil {
		return false, err
	}
	reply := client.Cmd("GET", gobKey(uid))
	if reply.Err != nil {
		return false, reply.Err
	}
	redisStore.Put(client)
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
	client, err := redisStore.Get()
	if err != nil {
		return err
	}
	if reply := client.Cmd("SET", gobKey(gob.UID), gobBytes); reply.Err != nil {
		return reply.Err
	}
	redisStore.Put(client)
	return nil
}

func (redisStore *RedisStore) GetGob(uid string) (*storage.Gob, error) {
	client, err := redisStore.Get()
	if err != nil {
		return nil, err
	}
	reply := client.Cmd("GET", gobKey(uid))
	if reply.Err != nil {
		return nil, reply.Err
	}
	redisStore.Put(client)
	if reply.Type == redis.NilReply {
		return nil, nil
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
	client, err := redisStore.Get()
	if err != nil {
		return nil, err
	}
	reply := client.Cmd("ZRANGE", "horde:"+hordeName, 0, -1)
	if reply.Err != nil {
		return nil, reply.Err
	}
	redisStore.Put(client)
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
	client, err := redisStore.Get()
	if err != nil {
		return err
	}
	if reply := client.Cmd("ZADD", hordeKey(hordeName), now.UnixNano(), buf.Bytes()); reply.Err != nil {
		return reply.Err
	}
	if reply := client.Cmd("SET", "uid.to.horde:"+uid, hordeName); reply.Err != nil {
		return reply.Err
	}
	redisStore.Put(client)
	return nil
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

func (redisStore *RedisStore) Configure(confStr string) {
	redisStore.SetConnection("tcp", confStr)
}
