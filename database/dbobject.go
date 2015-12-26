package database

import (
	"sync"

	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
	"gopkg.in/mgo.v2/bson"
)

type DbObject struct {
	Id types.Id `bson:"_id"`

	mutex     sync.RWMutex
	destroyed bool
}

func (self *DbObject) SetId(id types.Id) {
	self.Id = id
}

func (self *DbObject) GetId() types.Id {
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

func modified(id types.Id) {
	modifiedObjectChannel <- id
}

func idSetToList(set utils.Set) []types.Id {
	ids := make([]types.Id, len(set))

	i := 0
	for id := range set {
		ids[i] = bson.ObjectIdHex(id)
		i++
	}

	return ids
}
