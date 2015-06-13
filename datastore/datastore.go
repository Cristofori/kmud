package datastore

import (
	"sync"

	"github.com/Cristofori/kmud/types"

	"gopkg.in/mgo.v2/bson"
)

var _data map[bson.ObjectId]types.Object
var _mutex sync.RWMutex

func Init() {
	_data = map[bson.ObjectId]types.Object{}
}

func Get(id bson.ObjectId) types.Object {
	_mutex.RLock()
	defer _mutex.RUnlock()

	return _data[id]
}

func Contains(obj types.Identifiable) bool {
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

func Set(obj types.Object) {
	_mutex.Lock()
	defer _mutex.Unlock()

	_data[obj.GetId()] = obj
}

func Remove(obj types.Identifiable) {
	_mutex.Lock()
	defer _mutex.Unlock()

	delete(_data, obj.GetId())
}

func ClearAll() {
	_mutex.Lock()
	defer _mutex.Unlock()

	Init()
}
