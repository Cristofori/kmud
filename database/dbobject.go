package database

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
)

type DbObject struct {
	Id bson.ObjectId `bson:"_id"`

	mutex     sync.RWMutex
	destroyed bool
}

func (self *DbObject) initDbObject() {
	self.Id = bson.NewObjectId()
}

func (self *DbObject) GetId() bson.ObjectId {
	// Not mutex-protected since thd ID should never change
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

// vim: nocindent
