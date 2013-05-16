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

func GetCharacterRoom(character Character) (Room, error) {
	return GetRoom(character.GetRoomId())
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

		// TODO: Also sucks
		/*
			colorMode := user.getField(userColorMode).(int)

			switch utils.ColorMode(colorMode) {
			case utils.ColorModeLight:
				user.Fields[userColorMode] = utils.ColorModeLight
			case utils.ColorModeDark:
				user.Fields[userColorMode] = utils.ColorModeDark
			case utils.ColorModeNone:
				user.Fields[userColorMode] = utils.ColorModeNone
			default:
				panic("database.GetAllUsers(): Unhandled case in switch statement")
			}
		*/
	}

	return users, err
}

func GetAllCharacters() ([]*Character, error) {
	var characters []*Character
	err := findObjects(cCharacters, &characters)
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
	return zones, err
}

func GetAllItems() ([]*Item, error) {
	var items []*Item
	err := findObjects(cItems, &items)
	return items, err
}

func findRoom(query interface{}) (Room, error) {
	var room Room
	err := findObject(cRooms, query, &room)
	return room, err
}

func findCharacter(query interface{}) (Character, error) {
	var character Character
	err := findObject(cCharacters, query, &character)
	return character, err
}

func findUser(query interface{}) (User, error) {
	var user User
	err := findObject(cUsers, query, &user)
	return user, err
}

func GetUser(id bson.ObjectId) (User, error) {
	return findUser(bson.M{fId: id})
}

func GetUserByName(name string) (User, error) {
	return findUser(bson.M{fName: name})
}

func GetCharacter(id bson.ObjectId) (Character, error) {
	return findCharacter(bson.M{fId: id})
}

func GetCharacterByName(name string) (Character, error) {
	return findCharacter(bson.M{fName: name})
}

func GetRoom(id bson.ObjectId) (Room, error) {
	return findRoom(bson.M{fId: id})
}

func GetRoomByLocation(location Coordinate) (Room, error) {
	return findRoom(bson.M{fLocation: location})
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

func StartingRoom() (Room, error) {
	c := getCollection(cRooms)
	q := c.Find(bson.M{fDefault: true})

	count, err := q.Count()

	var room Room
	if err != nil {
		return room, err
	}

	if count == 0 {
		return room, errors.New("No default room found")
	}

	if count > 1 {
		fmt.Println("Warning: More than one default room found")
	}

	err = q.One(&room)

	return room, err
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
