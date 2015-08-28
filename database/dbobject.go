package database

import (
	"sync"

	"github.com/Cristofori/kmud/datastore"
	"github.com/Cristofori/kmud/types"
	"gopkg.in/mgo.v2/bson"
)

type DbObject struct {
	Id types.Id `bson:"_id"`

	mutex     sync.RWMutex
	destroyed bool
}

func (self *DbObject) init(obj types.Object) {
	self.Id = bson.NewObjectId()
	datastore.Set(obj)
	commitObject(self.Id)
}

func (self *DbObject) GetId() types.Id {
	// Not mutex-protected since the ID should never change
	return self.Id
}

func (self *DbObject) ReadLock() {
	self.mutex.RLock()
}

func (self *DbObject) ReadUnlock() {
	self.mutex.RUnlock()
}

func (self *DbObject) WriteLock() {
	self.mutex.Lock()
}

func (self *DbObject) WriteUnlock() {
	self.mutex.Unlock()
}

func (self *DbObject) Destroy() {
	self.WriteLock()
	defer self.WriteUnlock()

	self.destroyed = true
}

func (self *DbObject) IsDestroyed() bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.destroyed
}

func (self *DbObject) modified() {
	modifiedObjectChannel <- self.Id
}

func idMapToList(m map[string]bool) []types.Id {
	ids := make([]types.Id, len(m))

	i := 0
	for id := range m {
		ids[i] = bson.ObjectIdHex(id)
		i++
	}

	return ids
}

// vim: nocindent
