package database

import (
	"fmt"
	"labix.org/v2/mgo/bson"
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

var modifiedObjects map[Identifiable]bool
var modifiedObjectChannel chan Identifiable

var session Session

func Init(s Session) {
	session = s

	modifiedObjects = make(map[Identifiable]bool)
	modifiedObjectChannel = make(chan Identifiable)
	go watchModifiedObjects()
}

func modified(obj Identifiable) {
	modifiedObjectChannel <- obj
}

func watchModifiedObjects() {
	for {
		obj := <-modifiedObjectChannel
		commitObject(obj)
	}
}

func getCollection(collection collectionName) Collection {
	return session.DB("mud").C(string(collection))
}

func getCollectionOfObject(obj Identifiable) Collection {
	return getCollectionFromType(obj.GetType())
}

func getCollectionFromType(t objectType) Collection {
	switch t {
	case CharType:
		return getCollection(cCharacters)
	case RoomType:
		return getCollection(cRooms)
	case UserType:
		return getCollection(cUsers)
	case ZoneType:
		return getCollection(cZones)
	case ItemType:
		return getCollection(cItems)
	default:
		panic("database.getCollectionFromType: Unhandled object type")
	}
}

type collectionName string

// Collection names
const (
	cUsers      = collectionName("users")
	cCharacters = collectionName("characters")
	cRooms      = collectionName("rooms")
	cZones      = collectionName("zones")
	cItems      = collectionName("items")
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

func printError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func RetrieveObjects(t objectType, objects interface{}) error {
	c := getCollectionFromType(t)
	iter := c.Find(nil).Iter()
	return iter.All(objects)
}

func DeleteObject(obj Identifiable) error {
	c := getCollectionOfObject(obj)
	return c.RemoveId(obj.GetId())
}

func commitObject(object Identifiable) error {
	c := getCollectionFromType(object.GetType())
	object.ReadLock()
	err := c.UpsertId(object.GetId(), object)
	object.ReadUnlock()
	printError(err)
	return err
}

func updateField(c Collection, id bson.ObjectId, fieldName string, fieldValue interface{}) error {
	err := c.UpdateId(id, bson.M{"$set": bson.M{fieldName: fieldValue}})
	return err
}

// vim: nocindent
