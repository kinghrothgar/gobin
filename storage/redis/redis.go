package redis

import (
	"bytes"
	realgob "encoding/gob"
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/storage"
	"github.com/kinghrothgar/redis/pool"
	"github.com/mediocregopher/radix/redis"
	"time"
)

const (
	MB      int = 1048576
	DAY     int = 86400
	WEEK        = 7 * DAY
	YEAR        = 365 * DAY
	DEL_TTL     = WEEK
)

type RedisStore struct {
	*pool.Pool
}

// gobKey forms the redis key from the uid
func gobKey(uid string) string {
	return "gob:" + uid
}

// gobKey forms the redis key from the uid
func gobInfoKey(uid string) string {
	return "gobInfo:" + uid
}

// hordekey forms the redis key from the horde name for the list
func hordeListKey(hordeName string) string {
	return "hordeList:" + hordeName
}

// hordekey forms the redis key from the horde name for the list
func fifoListKey(fifoName string) string {
	return "fifoList:" + fifoName
}

// hordekey forms the redis key from the horde name for the hash
func hordeHashKey(hordeName string) string {
	return "hordeHash:" + hordeName
}

// delKey forms the redis key from the delete uid
func delKey(delUID string) string {
	return "del:" + delUID
}

func uidToHordeKey(uid string) string {
	return "uid.to.horde:" + uid
}

// deletedKey forms a key that will be inaccessble to users
func deletedKey(key string) string {
	return key + ":deleted"
}


// gobEncode encodes a storage gob into a byte array
func gobInfoEncode(gob *storage.GobInfo) ([]byte, error) {
	buf := new(bytes.Buffer)
	gobEnc := realgob.NewEncoder(buf)
	err := gobEnc.Encode(gob)
	return buf.Bytes(), err
}

func calculateTTL(gobBytes []byte) int {
	sizeMB := len(gobBytes) / MB
	if sizeMB < 1 {
		return YEAR
	}
	return (-20*sizeMB + 207) * DAY
}

// gobDecode decodes a byte array into a storage gob
func gobInfoDecode(gobInfoBytes []byte) (*storage.GobInfo, error) {
	gobInfo := &storage.GobInfo{}
	buf := bytes.NewReader(gobInfoBytes)
	gobDec := realgob.NewDecoder(buf)
	err := gobDec.Decode(gobInfo)
	return gobInfo, err
}

// New returns a new RedisStore
func New(confStr string) *RedisStore {
	return &RedisStore{pool.New("tcp", confStr, 100)}
}

// setTTL sets the expire time for the uid based on the size.  Expects to be
// run in a goroutine, so it does not return an error.  It instead logs it.
// TODO: should I be passing the client in?
func (redisStore *RedisStore) setTTLRoutine(client *pool.Client, uid string, data []byte) {
	defer redisStore.Put(client)
	ttl := calculateTTL(data)
	// If both replay with 0 int, both were successful
	i1, _ := client.Cmd("EXPIRE", gobKey(uid), ttl).Int()
	i2, _ := client.Cmd("EXPIRE", gobInfoKey(uid), ttl).Int()
	if (i1 + i2) == 0 {
		gslog.Error("REDIS: could not set expire time for uid '%s' to %d seconds", uid, ttl)
	}
}

// deleteExpire modifies the key so that it is inaccessble via normal methods
// and sets the TTL to a week
func (redisStore *RedisStore) deleteExpire(client *pool.Client, key string) error {
	// Make gob inaccessble using normal key
	reply := client.Cmd("RENAME", key, deletedKey(key))
	if reply.Err != nil {
		return reply.Err
	}
	reply = client.Cmd("EXPIRE", deletedKey(key), DEL_TTL)
	if reply.Err != nil {
		return reply.Err
	}
	if i, _ := reply.Int(); i == 0 {
		gslog.Error("REDIS: could not set expire time for deleted key '%s'", key)
	}
	return nil
}

func (redisStore *RedisStore) deleteExpireRoutine(client *pool.Client, key string) {
	err := redisStore.deleteExpire(client, key)
	if err != nil {
		gslog.Error("REDIS: could not set expire time for deleted key '%s' with error: %s", key)
		return
	}
	redisStore.Put(client)
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

func (redisStore *RedisStore) DelUIDExist(delUID string) (bool, error) {
	client, err := redisStore.Get()
	if err != nil {
		return false, err
	}
	reply := client.Cmd("GET", delKey(delUID))
	if reply.Err != nil {
		return false, reply.Err
	}
	redisStore.Put(client)
	if reply.Type == redis.NilReply {
		return false, nil
	}
	return true, nil
}

func (redisStore *RedisStore) PutGob(data []byte, gobInfo *storage.GobInfo) error {
	gobInfoBytes, err := gobInfoEncode(gobInfo)
	if err != nil {
		return err
	}
	client, err := redisStore.Get()
	if err != nil {
		return err
	}
	// Set the gob data and info to their respective keys
	if reply := client.Cmd("MSET", gobKey(gobInfo.UID), data, gobInfoKey(gobInfo.UID), gobInfoBytes); reply.Err != nil {
		return reply.Err
	}
	// TODO: should I just run this in a goroutine and not worry about it?
	if reply := client.Cmd("SET", delKey(gobInfo.DelUID), gobInfo.UID); reply.Err != nil {
		return reply.Err
	}
	go redisStore.setTTLRoutine(client, gobInfo.UID, data)
	return nil
}

func (redisStore *RedisStore) AppendGob(uid string, data []byte) error {
	client, err := redisStore.Get()
	if err != nil {
		return err
	}
	if reply := client.Cmd("APPEND", gobKey(gob.UID), data); reply.Err != nil {
		return reply.Err
	}
	// TODO: should I just run this in a goroutine and not worry about it?
	if reply := client.Cmd("SET", delKey(gob.DelUID), gob.UID); reply.Err != nil {
		return reply.Err
	}
	go redisStore.setTTLRoutine(client, gob.UID, gobBytes)
	return nil
}

func (redisStore *RedisStore) GetGob(uid string) ([]byte, *storage.GobInfo, error) {
	client, err := redisStore.Get()
	if err != nil {
		return nil, nil, err
	}
	reply := client.Cmd("MGET", gobKey(uid), gobInfoKey(uid))
	if reply.Err != nil {
		return nil, nil, reply.Err
	}
	list, err := reply.ListBytes()
	if err != nil {
		return nil, nil, err
	}
	if list[0] == nil || list[1] == nil {
		return []byte{}, nil, nil
	}
	data := list[0]
	gobInfoBytes := list[1]
	gobInfo, err := gobInfoDecode(gobInfoBytes)
	// Set TTL if no error
	// TODO: Should I do this?
	if err == nil {
		go redisStore.setTTLRoutine(client, uid, data)
	} else {
		redisStore.Put(client)
	}
	return data, gobInfo, err
}

func (redisStore *RedisStore) GetGobLen(uid string) (int, error) {
	client, err := redisStore.Get()
	if err != nil {
		return -1, err
	}
	defer redisStore.Put(client)
	reply := client.Cmd("STRLEN", gobKey(uid))
	if reply.Err != nil {
		return -1, reply.Err
	}
	return reply.Int()

}


func (redisStore *RedisStore) DelGob(uid string) error {
	key := gobKey(uid)
	// Make gob inaccessble using normal methods
	client, err := redisStore.Get()
	if err != nil {
		return err
	}
	err = redisStore.deleteExpire(client, key)
	if err != nil {
		return err
	}
	redisStore.Put(client)
	return nil
}

func (redisStore *RedisStore) DelUIDToUID(delUID string) (string, error) {
	client, err := redisStore.Get()
	if err != nil {
		return "", err
	}
	reply := client.Cmd("GET", delKey(delUID))
	if reply.Err != nil {
		return "", reply.Err
	}
	redisStore.Put(client)
	uid, err := reply.Str()
	if err != nil {
		return "", err
	}
	return uid, nil
}

// TODO: Verify uid's still exist?
func (redisStore *RedisStore) GetHorde(hordeName string) (storage.Horde, error) {
	client, err := redisStore.Get()
	if err != nil {
		return nil, err
	}
	reply := client.Cmd("LRANGE", hordeListKey(hordeName), 0, -1)
	if reply.Err != nil {
		return nil, reply.Err
	}
	hordeList, _ := reply.List()
	reply = client.Cmd("HGETALL", hordeHashKey(hordeName))
	if reply.Err != nil {
		return nil, reply.Err
	}
	hordeHash, _ := reply.Hash()
	redisStore.Put(client)
	length := len(hordeList)
	horde := make(storage.Horde, length)
	for i, uid := range hordeList {
		uidCreated := &storage.UIDCreated{
			UID: hordeList[i],
			Created: hordeHash[uid],
		}
		horde[i] = uidCreated
	}
	return horde, nil
}

func (redisStore *RedisStore) AddUIDHorde(hordeName string, uid string) error {
	now := time.Now()
	client, err := redisStore.Get()
	if err != nil {
		return err
	}
	if reply := client.Cmd("LPUSH", hordeListKey(hordeName), uid); reply.Err != nil {
		return reply.Err
	}
	if reply := client.Cmd("HSET", hordeHashKey(hordeName), uid, now.String()); reply.Err != nil {
		return reply.Err
	}
	if reply := client.Cmd("SET", uidToHordeKey(uid), hordeName); reply.Err != nil {
		return reply.Err
	}
	redisStore.Put(client)
	return nil
}

func (redisStore *RedisStore) uidToHorde(client *pool.Client, uid string) (string, error) {
	reply := client.Cmd("GET", uidToHordeKey(uid))
	if reply.Err != nil {
		return "", reply.Err
	}
	if reply.Type == redis.NilReply {
		return "", nil
	}
	hordeName, err := reply.Str()
	if err != nil {
		return "", err
	}
	return hordeName, nil
}

// DelUIDHorde deletes uid from a horde if it is in one.  Returns an error if
// it fails to connect to Redis or it fails to remove it from a horde
func (redisStore *RedisStore) DelUIDHorde(uid string) error {
	client, err := redisStore.Get()
	if err != nil {
		return err
	}
	hordeName, err := redisStore.uidToHorde(client, uid)
	if err != nil {
		return err
	}
	if hordeName == "" {
		redisStore.Put(client)
		return nil
	}
	if reply := client.Cmd("LREM", hordeListKey(hordeName), 1, uid); reply.Err != nil {
		return reply.Err
	}
	if reply := client.Cmd("HDEL", hordeHashKey(hordeName), uid); reply.Err != nil {
		return reply.Err
	}
	go redisStore.deleteExpireRoutine(client, uidToHordeKey(uid))
	return nil
}

func (redisStore *RedisStore) CurrentUIDFIFO(fifoName string) (string, error) {
	client, err := redisStore.Get()
	if err != nil {
		return "", err
	}
	reply := client.Cmd("LRANGE", fifoListKey(fifoName), -1, -1)
	list, err := reply.List()
	if err != nil {
		return "", err
	}
	if len(list) > 0 {
		return list[0], nil
	}
	return "", nil
}

func (redisStore *RedisStore) Configure(confStr string) {
	redisStore.SetConnection("tcp", confStr)
}
