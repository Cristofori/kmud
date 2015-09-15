package datastore

import (
	"sync"

	"github.com/Cristofori/kmud/types"
)

var _data map[types.Id]types.Object
var _mutex sync.RWMutex

func init() {
	ClearAll()
}

func Get(id types.Id) types.Object {
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

func ContainsId(id types.Id) bool {
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
	RemoveId(obj.GetId())
}

func RemoveId(id types.Id) {
	_mutex.Lock()
	defer _mutex.Unlock()

	delete(_data, id)
}

func ClearAll() {
	_mutex.Lock()
	defer _mutex.Unlock()

	_data = map[types.Id]types.Object{}
}
