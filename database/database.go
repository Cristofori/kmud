package database

import (
	"errors"
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
		// fmt.Println("Processing update:", op.id, op.field, op.value)
		commitObject(obj)
	}
}

func getCollection(collection collectionName) Collection {
	return session.DB("mud").C(string(collection))
}

func getCollectionFromType(t objectType) Collection {
	switch t {
	case charType:
		return getCollection(cCharacters)
	case roomType:
		return getCollection(cRooms)
	case userType:
		return getCollection(cUsers)
	case zoneType:
		return getCollection(cZones)
	case itemType:
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

func findObject(collection collectionName, query interface{}, object interface{}) error {
	c := getCollection(collection)
	q := c.Find(query)

	count, err := q.Count()

	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New(fmt.Sprintf("Query return no results: %v", query))
	}

	err = q.One(object)

	return err
}

func findObjects(collection collectionName, objects interface{}) error {
	c := getCollection(collection)
	iter := c.Find(nil).Iter()
	return iter.All(objects)
}

func GetAllUsers() ([]*User, error) {
	var users []*User
	err := findObjects(cUsers, &users)

	for _, user := range users {
		user.objType = userType
	}

	return users, err
}

func GetAllCharacters() ([]*Character, error) {
	var characters []*Character
	err := findObjects(cCharacters, &characters)

	for _, char := range characters {
		char.objType = charType
	}

	return characters, err
}

func GetAllRooms() ([]*Room, error) {
	var rooms []*Room
	err := findObjects(cRooms, &rooms)

	for _, room := range rooms {
		room.objType = roomType
	}

	return rooms, err
}

func GetAllZones() ([]*Zone, error) {
	var zones []*Zone
	err := findObjects(cZones, &zones)

	for _, zone := range zones {
		zone.objType = zoneType
	}

	return zones, err
}

func GetAllItems() ([]*Item, error) {
	var items []*Item
	err := findObjects(cItems, &items)

	for _, item := range items {
		item.objType = itemType
	}

	return items, err
}

func DeleteRoom(id bson.ObjectId) error {
	c := getCollection(cRooms)
	return c.RemoveId(id)
}

func DeleteUser(id bson.ObjectId) error {
	c := getCollection(cUsers)
	return c.RemoveId(id)
}

func DeleteCharacter(id bson.ObjectId) error {
	c := getCollection(cCharacters)
	return c.Remove(bson.M{fId: id})
}

func DeleteItem(id bson.ObjectId) error {
	c := getCollection(cItems)
	return c.Remove(bson.M{fId: id})
}

func DeleteAllRooms() {
	c := getCollection(cRooms)
	c.DropCollection()
}

func commitObject(object Identifiable) error {
	c := getCollectionFromType(object.GetType())
	err := c.UpsertId(object.GetId(), object)
	printError(err)
	return err
}

func updateField(c Collection, id bson.ObjectId, fieldName string, fieldValue interface{}) error {
	err := c.UpdateId(id, bson.M{"$set": bson.M{fieldName: fieldValue}})
	return err
}

// vim: nocindent
