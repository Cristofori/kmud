package database

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"kmud/datastore"
	"kmud/utils"
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

var modifiedObjects map[bson.ObjectId]bool
var modifiedObjectChannel chan bson.ObjectId

var _session Session
var _dbName string

func Init(session Session, dbName string) {
	_session = session
	_dbName = dbName

	modifiedObjects = make(map[bson.ObjectId]bool)
	modifiedObjectChannel = make(chan bson.ObjectId, 10)

	go watchModifiedObjects()
}

func objectModified(obj datastore.Identifiable) {
	modifiedObjectChannel <- obj.GetId()
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
	for id := range modifiedObjects {
		commitObject(id)
	}
	modifiedObjectsMutex.Unlock()
}

func getCollection(collection collectionName) Collection {
	return _session.DB(_dbName).C(string(collection))
}

func getCollectionOfObject(obj datastore.Identifiable) Collection {
	return getCollectionFromType(obj.GetType())
}

func getCollectionFromType(t datastore.ObjectType) Collection {
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

func RetrieveObjects(t datastore.ObjectType, objects interface{}) error {
	c := getCollectionFromType(t)
	return c.Find(nil).Iter().All(objects)
}

func Find(t datastore.ObjectType, key string, value interface{}) []bson.ObjectId {
	return find(t, bson.M{key: value})
}

func FindAll(t datastore.ObjectType) []bson.ObjectId {
	return find(t, nil)
}

func find(t datastore.ObjectType, query interface{}) []bson.ObjectId {
	c := getCollectionFromType(t)

	var results []interface{}
	c.Find(query).Iter().All(&results)

	var ids []bson.ObjectId

	for _, result := range results {
		ids = append(ids, result.(bson.M)["_id"].(bson.ObjectId))
	}

	return ids
}

func DeleteObject(obj datastore.Identifiable) error {
	obj.Destroy()
	c := getCollectionOfObject(obj)

	err := c.RemoveId(obj.GetId())

	if err != nil {
		fmt.Println("Delete object failed", obj.GetId())
	}

	return err
}

func commitObject(id bson.ObjectId) error {
	object := datastore.Get(id)

	if object == nil || object.IsDestroyed() {
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
