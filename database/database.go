package database

import (
	"errors"
	"fmt"
	"kmud/utils"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type UpdateOperation struct {
	id         bson.ObjectId
	collection *mgo.Collection
	field      string
	value      interface{}
}

var session *mgo.Session

var updateChannel chan UpdateOperation

func Init(s *mgo.Session) {
	session = s
	updateChannel = make(chan UpdateOperation)
	go processUpdates()
}

func updateObject(obj Identifiable, field string, value interface{}) {
	var c *mgo.Collection = nil

	switch obj.GetType() {
	case characterType:
		c = getCollection(session, cCharacters)
	case roomType:
		c = getCollection(session, cRooms)
	case userType:
		c = getCollection(session, cUsers)
	default:
		panic("database.updateObject: Unhandled object type")
	}

	updateChannel <- UpdateOperation{id: obj.GetId(), field: field, collection: c, value: value}
}

func processUpdates() {
	for {
		op := <-updateChannel
		// fmt.Println("Processing update:", op.id, op.field, op.value)
		utils.HandleError(updateField(session, op.collection, op.id, op.field, op.value))
	}
}

func getCollection(session *mgo.Session, collection collectionName) *mgo.Collection {
	return session.DB("mud").C(string(collection))
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

func GetCharacterRoom(session *mgo.Session, character Character) (Room, error) {
	return GetRoom(session, character.GetRoomId())
}

func findObject(session *mgo.Session, collection collectionName, query interface{}, object interface{}) error {
	c := getCollection(session, collection)
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

func findObjects(session *mgo.Session, collection collectionName, objects interface{}) error {
	c := getCollection(session, collection)
	iter := c.Find(nil).Iter()
	return iter.All(objects)
}

func GetAllUsers(session *mgo.Session) ([]User, error) {
	var users []User
	err := findObjects(session, cUsers, &users)

	for _, user := range users {
		user.objType = userType

		// TODO: Also sucks
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
	}

	return users, err
}

// TODO: Find a better way to do this, this sucks
//       The bson.ObjectId list contained in character inventory gets deserialized
//       as a []interface{} instead of as a []bson.ObjectId, this converts it to
//       a []bson.ObjectId which makes it much easier to deal with throughout the
//       rest of the code.
func convertFieldToIdSlice(object *DbObject, field string) {
	found := object.hasField(field)
	if !found {
		object.Fields[field] = []bson.ObjectId{}
	} else {
		var idList []bson.ObjectId
		for _, id := range object.getField(field).([]interface{}) {
			idList = append(idList, id.(bson.ObjectId))
		}
		object.Fields[field] = idList
	}
}

func GetAllCharacters(session *mgo.Session) ([]*Character, error) {
	var characters []*Character
	err := findObjects(session, cCharacters, &characters)

	for _, char := range characters {
		char.objType = characterType
		convertFieldToIdSlice(&char.DbObject, characterInventory)
	}

	return characters, err
}

func GetAllRooms(session *mgo.Session) ([]*Room, error) {
	var rooms []*Room
	err := findObjects(session, cRooms, &rooms)

	for _, room := range rooms {
		room.objType = roomType

		// TODO: Also sucks
		location := room.getField(roomLocation).(map[string]interface{})

		var coord Coordinate
		coord.X = location["x"].(int)
		coord.Y = location["y"].(int)
		coord.Z = location["z"].(int)
		room.Fields[roomLocation] = coord

		convertFieldToIdSlice(&room.DbObject, roomItems)
	}

	return rooms, err
}

func GetAllZones(session *mgo.Session) ([]Zone, error) {
	var zones []Zone
	err := findObjects(session, cZones, &zones)
	return zones, err
}

func GetAllItems(session *mgo.Session) ([]Item, error) {
	var items []Item
	err := findObjects(session, cItems, &items)
	return items, err
}

func findRoom(session *mgo.Session, query interface{}) (Room, error) {
	var room Room
	err := findObject(session, cRooms, query, &room)
	return room, err
}

func findCharacter(session *mgo.Session, query interface{}) (Character, error) {
	var character Character
	err := findObject(session, cCharacters, query, &character)
	return character, err
}

func findUser(session *mgo.Session, query interface{}) (User, error) {
	var user User
	err := findObject(session, cUsers, query, &user)
	return user, err
}

func GetUser(session *mgo.Session, id bson.ObjectId) (User, error) {
	return findUser(session, bson.M{fId: id})
}

func GetUserByName(session *mgo.Session, name string) (User, error) {
	return findUser(session, bson.M{fName: name})
}

func GetCharacter(session *mgo.Session, id bson.ObjectId) (Character, error) {
	return findCharacter(session, bson.M{fId: id})
}

func GetCharacterByName(session *mgo.Session, name string) (Character, error) {
	return findCharacter(session, bson.M{fName: name})
}

func GetRoom(session *mgo.Session, id bson.ObjectId) (Room, error) {
	return findRoom(session, bson.M{fId: id})
}

func GetRoomByLocation(session *mgo.Session, location Coordinate) (Room, error) {
	return findRoom(session, bson.M{fLocation: location})
}

func DeleteRoom(session *mgo.Session, id bson.ObjectId) error {
	c := getCollection(session, cRooms)
	return c.RemoveId(id)
}

func DeleteUser(session *mgo.Session, id bson.ObjectId) error {
	c := getCollection(session, cUsers)
	return c.RemoveId(id)
}

func DeleteCharacter(session *mgo.Session, id bson.ObjectId) error {
	c := getCollection(session, cCharacters)
	return c.Remove(bson.M{fId: id})
}

func DeleteItem(session *mgo.Session, id bson.ObjectId) error {
	c := getCollection(session, cItems)
	return c.Remove(bson.M{fId: id})
}

func StartingRoom(session *mgo.Session) (Room, error) {
	c := getCollection(session, cRooms)
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

func DeleteAllRooms(session *mgo.Session) {
	c := getCollection(session, cRooms)
	c.DropCollection()
}

func CreateUser(session *mgo.Session, name string) (User, error) {
	user, err := findUser(session, bson.M{fName: name})

	if err == nil {
		return user, errors.New("That user already exists")
	}

	user = NewUser(name)
	err = CommitUser(session, user)
	return user, err
}

func commitObject(session *mgo.Session, c *mgo.Collection, object Identifiable) error {
	_, err := c.UpsertId(object.GetId(), object)
	printError(err)
	return err
}

func updateField(session *mgo.Session, c *mgo.Collection, id bson.ObjectId, fieldName string, fieldValue interface{}) error {
	err := c.UpdateId(id, bson.M{"$set": bson.M{fieldName: fieldValue}})
	return err
}

func CommitUser(session *mgo.Session, user User) error {
	return commitObject(session, getCollection(session, cUsers), user)
}

func CommitRoom(session *mgo.Session, room Room) error {
	return commitObject(session, getCollection(session, cRooms), room)
}

func CommitCharacter(session *mgo.Session, character Character) error {
	return commitObject(session, getCollection(session, cCharacters), character)
}

func CommitZone(session *mgo.Session, zone Zone) error {
	return commitObject(session, getCollection(session, cZones), zone)
}

func CommitItem(session *mgo.Session, item Item) error {
	return commitObject(session, getCollection(session, cItems), item)
}

// vim: nocindent
