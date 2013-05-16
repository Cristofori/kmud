package database

import (
	"kmud/utils"
	"labix.org/v2/mgo/bson"
	"sync"
)

type objectType int

type Identifiable interface {
	GetId() bson.ObjectId
	GetType() objectType
}

const (
	charType objectType = iota
	roomType objectType = iota
	userType objectType = iota
	zoneType objectType = iota
	itemType objectType = iota
)

type DbObject struct {
	Id bson.ObjectId `bson:"_id"`

	Name string

	objType objectType
	mutex   sync.RWMutex
}

func (self *DbObject) initDbObject(name string, objType objectType) {
	self.Id = bson.NewObjectId()
	self.objType = objType
	self.Name = name
}

func (self *DbObject) GetId() bson.ObjectId {
	// Not mutex-protected since thd ID should never change
	return self.Id
}

func (self *DbObject) GetType() objectType {
	// Not mutex-protected since the object type should never change
	return self.objType
}

func (self *DbObject) PrettyName() string {
	return utils.FormatName(self.GetName())
}

func (self *DbObject) SetName(name string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if name != self.Name {
		self.Name = name
		modified(self)
	}
}

func (self *DbObject) GetName() string {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.Name
}

// vim: nocindent
