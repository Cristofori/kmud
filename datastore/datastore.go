package datastore

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
)

type ObjectType int

type Identifiable interface {
	GetId() bson.ObjectId
	GetType() ObjectType
	ReadLock()
	ReadUnlock()
	Destroy()
	IsDestroyed() bool
}

var _data map[bson.ObjectId]Identifiable
var _mutex sync.RWMutex

func Init() {
	_data = map[bson.ObjectId]Identifiable{}
}

func Get(id bson.ObjectId) Identifiable {
	_mutex.RLock()
	defer _mutex.RUnlock()

	return _data[id]
}

func Contains(obj Identifiable) bool {
	_mutex.RLock()
	defer _mutex.RUnlock()

	_, found := _data[obj.GetId()]
	return found
}

func ContainsId(id bson.ObjectId) bool {
	_mutex.RLock()
	defer _mutex.RUnlock()

	_, found := _data[id]
	return found
}

func Set(obj Identifiable) {
	_mutex.Lock()
	defer _mutex.Unlock()

	_data[obj.GetId()] = obj
}

func Remove(obj Identifiable) {
	_mutex.Lock()
	defer _mutex.Unlock()

	delete(_data, obj.GetId())
}

func ClearAll() {
	_mutex.Lock()
	defer _mutex.Unlock()

	Init()
}
