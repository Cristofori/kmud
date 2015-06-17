package database

import (
	"fmt"
	"time"

	"github.com/Cristofori/kmud/datastore"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
	"gopkg.in/mgo.v2/bson"
)

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

var modifiedObjects map[types.Id]bool
var modifiedObjectChannel chan types.Id

var _session Session
var _dbName string

func Init(session Session, dbName string) {
	_session = session
	_dbName = dbName

	modifiedObjects = make(map[types.Id]bool)
	modifiedObjectChannel = make(chan types.Id, 1)

	watchModifiedObjects()
}

func objectModified(obj types.Identifiable) {
	modifiedObjectChannel <- obj.GetId()
}

func watchModifiedObjects() {
	go func() {
		timeout := make(chan bool)

		startTimeout := func() {
			go func() {
				time.Sleep(1 * time.Second)
				timeout <- true
			}()
		}

		startTimeout()

		for {
			select {
			case id := <-modifiedObjectChannel:
				modifiedObjects[id] = true
			case <-timeout:
				for id := range modifiedObjects {
					go commitObject(id)
				}
				startTimeout()
			}
		}
	}()
}

func getCollection(collection collectionName) Collection {
	return _session.DB(_dbName).C(string(collection))
}

func getCollectionOfObject(obj types.Identifiable) Collection {
	return getCollectionFromType(obj.GetType())
}

func getCollectionFromType(t types.ObjectType) Collection {
	switch t {
	case types.PcType:
		return getCollection(cPlayerChars)
	case types.NpcType:
		return getCollection(cNonPlayerChars)
	case types.UserType:
		return getCollection(cUsers)
	case types.ZoneType:
		return getCollection(cZones)
	case types.AreaType:
		return getCollection(cAreas)
	case types.RoomType:
		return getCollection(cRooms)
	case types.ItemType:
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

func RetrieveObjects(t types.ObjectType, objects interface{}) error {
	c := getCollectionFromType(t)
	return c.Find(nil).Iter().All(objects)
}

func Find(t types.ObjectType, query bson.M) []bson.ObjectId {
	return find(t, query)
}

func FindOne(t types.ObjectType, query bson.M) types.Id {
	var result bson.M
	find_helper(t, query).One(&result)
	id, found := result["_id"]
	if found {
		return id.(bson.ObjectId)
	}
	return nil
}

func FindAll(t types.ObjectType) []bson.ObjectId {
	return find(t, nil)
}

func find(t types.ObjectType, query interface{}) []bson.ObjectId {
	var results []bson.M
	find_helper(t, query).Iter().All(&results)

	var ids []bson.ObjectId
	for _, result := range results {
		ids = append(ids, result["_id"].(bson.ObjectId))
	}

	return ids
}

func find_helper(t types.ObjectType, query interface{}) Query {
	c := getCollectionFromType(t)
	return c.Find(query)
}

func DeleteObject(obj types.Object) error {
	obj.Destroy()
	c := getCollectionOfObject(obj)

	err := c.RemoveId(obj.GetId())

	if err != nil {
		fmt.Println("Delete object failed", obj.GetId())
	}

	return err
}

func commitObject(id types.Id) error {
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
