package database

import (
	"kmud/utils"
	"fmt"
	// "labix.org/v2/mgo/bson"
	"sync"
)

var modifiedObjectsMutex sync.Mutex

type Session interface {
	DB(string) Database
}

type Database interface {
	C(string) Collection
}

type Collection interface {
	Find(interface{}) Query
	RemoveId(interface{}) error
	Remove(interface{}) error
	DropCollection() error
	UpdateId(interface{}, interface{}) error
	UpsertId(interface{}, interface{}) error
}

type Query interface {
	Count() (int, error)
	One(interface{}) error
	Iter() Iterator
}

type Iterator interface {
	All(interface{}) error
}

var modifiedObjects map[Identifiable]bool
var modifiedObjectChannel chan Identifiable

var session Session

func Init(s Session) {
	session = s

	modifiedObjects = make(map[Identifiable]bool)
	modifiedObjectChannel = make(chan Identifiable, 10)

	go watchModifiedObjects()
}

func objectModified(obj Identifiable) {
	modifiedObjectChannel <- obj
}

func watchModifiedObjects() {
	for {
		id := <-modifiedObjectChannel
		modifiedObjectsMutex.Lock()
		modifiedObjects[id] = true
		modifiedObjectsMutex.Unlock()

		// TODO FIXME - Periodically save in separate routine
		saveModifiedObjects()
	}
}

func saveModifiedObjects() {
	modifiedObjectsMutex.Lock()
	for obj := range modifiedObjects {
		commitObject(obj)
	}
	modifiedObjectsMutex.Unlock()
}

func getCollection(collection collectionName) Collection {
	return session.DB("mud").C(string(collection))
}

func getCollectionOfObject(obj Identifiable) Collection {
	return getCollectionFromType(obj.GetType())
}

func getCollectionFromType(t objectType) Collection {
	switch t {
	case PcType:
		return getCollection(cPlayerChars)
	case NpcType:
		return getCollection(cNonPlayerChars)
	case UserType:
		return getCollection(cUsers)
	case ZoneType:
		return getCollection(cZones)
	case AreaType:
		return getCollection(cAreas)
	case RoomType:
		return getCollection(cRooms)
	case ItemType:
		return getCollection(cItems)
	default:
		panic("database.getCollectionFromType: Unhandled object type")
	}
}

type collectionName string

// Collection names
const (
	cUsers          = collectionName("users")
	cPlayerChars    = collectionName("player_characters")
	cNonPlayerChars = collectionName("npcs")
	cRooms          = collectionName("rooms")
	cZones          = collectionName("zones")
	cItems          = collectionName("items")
	cAreas          = collectionName("areas")
)

// Field names
const (
	fId           = "_id"
	fName         = "name"
	fCharacterIds = "characterids"
	fRoom         = "room"
	fLocation     = "location"
	fDefault      = "default"
)

// MongDB operations
const (
	SET  = "$set"
	PUSH = "$push"
	PULL = "$pull"
)

func RetrieveObjects(t objectType, objects interface{}) error {
	c := getCollectionFromType(t)
	return c.Find(nil).Iter().All(objects)
}

func DeleteObject(obj Identifiable) error {
	obj.Destroy()
	c := getCollectionOfObject(obj)
	return c.RemoveId(obj.GetId())
}

func commitObject(object Identifiable) error {
	if object.IsDestroyed() {
		return nil
	}

	c := getCollectionFromType(object.GetType())

	object.ReadLock()
	err := c.UpsertId(object.GetId(), object)
	object.ReadUnlock()

	if err != nil {
		fmt.Println("Update failed", object.GetId())
	}

	utils.HandleError(err)
	return err
}

// vim: nocindent
