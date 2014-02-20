package memory

import (
	"errors"
	"github.com/kinghrothgar/goblin/storage"
	"sync"
)

type Gobs map[string]*storage.Gob
type Hordes map[string]storage.Horde

type MemoryStore struct {
	gobs   Gobs
	hordes Hordes
	lock   sync.RWMutex
}

func (memoryStore *MemoryStore) UIDExist(uid string) (bool, error) {
	memoryStore.lock.RLock()
	defer memoryStore.lock.RUnlock()
	if _, ok := memoryStore.gobs[uid]; ok {
		return true, nil
	}
	return false, nil
}

func (memoryStore *MemoryStore) PutGob(gob *storage.Gob) error {
	memoryStore.lock.Lock()
	defer memoryStore.lock.Unlock()
	memoryStore.gobs[gob.UID] = gob
	return nil
}

func (memoryStore *MemoryStore) GetGob(uid string) (*storage.Gob, error) {
	memoryStore.lock.RLock()
	defer memoryStore.lock.RUnlock()
	if gob, ok := memoryStore.gobs[uid]; ok {
		return gob, nil
	}
	return nil, errors.New("uid does not exist")
}

func (memoryStore *MemoryStore) DelGob(uid string) error {
	memoryStore.lock.Lock()
	defer memoryStore.lock.Unlock()
	// TODO: do I need to check?
	if exist, _ := memoryStore.UIDExist(uid); !exist {
		return errors.New("uid does not exist")
	}
	delete(memoryStore.gobs, uid)
	return nil
}

func (memoryStore *MemoryStore) GetHorde(hordeName string) (storage.Horde, error) {
	memoryStore.lock.RLock()
	defer memoryStore.lock.RUnlock()
	if horde, ok := memoryStore.hordes[hordeName]; ok {
		return horde, nil
	}
	return storage.Horde{}, nil
}

func (memoryStore *MemoryStore) AddToHorde(hordeName string, uid string) error {
	memoryStore.lock.Lock()
	defer memoryStore.lock.Unlock()
	// TODO: do I need to do this
	if horde, ok := memoryStore.hordes[hordeName]; ok {
		horde[uid] = true
	} else {
		memoryStore.hordes[hordeName] = storage.Horde{uid: true}
	}
	return nil
}

func (memoryStore *MemoryStore) DelFromHorde(hordeName string, uid string) error {
	memoryStore.lock.Lock()
	defer memoryStore.lock.Unlock()
	// TODO: should I even be checking if I'm really deleting?
	horde, ok := memoryStore.hordes[hordeName]
	if !ok {
		return errors.New("horde does not exist")
	}
	if _, ok = horde[uid]; !ok {
		return errors.New("uid does not exist in horde")
	}
	delete(horde, uid)
	return nil
}

func (memoryStore *MemoryStore) Initialize(confStr string) error {
	memoryStore.lock.Lock()
	defer memoryStore.lock.Unlock()
	memoryStore.gobs = Gobs{}
	memoryStore.hordes = Hordes{}
	return nil
}
