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

var modifiedObjects = map[types.Id]bool{}
var modifiedObjectChannel chan types.Id

var _session Session
var _dbName string

func Init(session Session, dbName string) {
	_session = session
	_dbName = dbName

	modifiedObjectChannel = make(chan types.Id, 1)

	watchModifiedObjects()
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
				modifiedObjects = map[types.Id]bool{}
				startTimeout()
			}
		}
	}()
}

func getCollection(collection collectionName) Collection {
	return _session.DB(_dbName).C(string(collection))
}

func getCollectionOfObject(obj types.Object) Collection {
	switch obj.(type) {
	case types.PC:
		return getCollection(cPlayerChars)
	case types.NPC:
		return getCollection(cNonPlayerChars)
	case types.Spawner:
		return getCollection(cSpawners)
	case types.User:
		return getCollection(cUsers)
	case types.Zone:
		return getCollection(cZones)
	case types.Area:
		return getCollection(cAreas)
	case types.Room:
		return getCollection(cRooms)
	case types.Item:
		return getCollection(cItems)
	case types.Skill:
		return getCollection(cSkills)
	case types.Shop:
		return getCollection(cShops)
	default:
		panic(fmt.Sprintf("unrecognized object in getCollectionOfObject: %v", obj))
	}
}

func getCollectionFromType(t types.ObjectType) Collection {
	switch t {
	case types.PcType:
		return getCollection(cPlayerChars)
	case types.NpcType:
		return getCollection(cNonPlayerChars)
	case types.SpawnerType:
		return getCollection(cSpawners)
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
	case types.SkillType:
		return getCollection(cSkills)
	case types.ShopType:
		return getCollection(cShops)
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
	cSpawners       = collectionName("spawners")
	cRooms          = collectionName("rooms")
	cZones          = collectionName("zones")
	cItems          = collectionName("items")
	cAreas          = collectionName("areas")
	cSkills         = collectionName("skills")
	cShops          = collectionName("shops")
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

func RetrieveObjects(t types.ObjectType, objects interface{}) {
	c := getCollectionFromType(t)
	err := c.Find(nil).Iter().All(objects)
	utils.HandleError(err)
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

func DeleteObject(id types.Id) {
	object := datastore.Get(id)
	datastore.Remove(object)

	object.Destroy()

	c := getCollectionOfObject(object)
	utils.HandleError(c.RemoveId(object.GetId()))
}

func commitObject(id types.Id) {
	object := datastore.Get(id)

	if object == nil || object.IsDestroyed() {
		return
	}

	c := getCollectionOfObject(object)

	object.ReadLock()
	err := c.UpsertId(object.GetId(), object)
	object.ReadUnlock()

	if err != nil {
		fmt.Println("Update failed", object.GetId())
	}

	utils.HandleError(err)
}

// vim: nocindent
