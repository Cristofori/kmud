package database

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type collectionName string

func getCollection(session *mgo.Session, collection collectionName) *mgo.Collection {
	return session.DB("mud").C(string(collection))
}

// Collection names
const (
	cUsers      = collectionName("users")
	cCharacters = collectionName("characters")
	cRooms      = collectionName("rooms")
	cZones      = collectionName("zones")
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

// DB commands
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
	return GetRoom(session, character.RoomId)
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
	return users, err
}

func GetAllCharacters(session *mgo.Session) ([]Character, error) {
	var characters []Character
	err := findObjects(session, cCharacters, &characters)
	return characters, err
}

func GetAllRooms(session *mgo.Session) ([]Room, error) {
	var rooms []Room
	err := findObjects(session, cRooms, &rooms)
	return rooms, err
}

func GetAllZones(session *mgo.Session) ([]Zone, error) {
	var zones []Zone
	err := findObjects(session, cZones, &zones)
	return zones, err
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

func CreateCharacter(session *mgo.Session, user *User, characterName string) (Character, error) {
	character, err := GetCharacterByName(session, characterName)

	if err == nil {
		return character, errors.New("That character already exists")
	}

	startingRoom, err := StartingRoom(session)

	if err != nil {
		fmt.Println("Error getting starting room:", err)
		return character, err
	}

	character = NewCharacter(characterName)
	character.RoomId = startingRoom.Id

	err = CommitCharacter(session, character)

	if err != nil {
		fmt.Println("Error inserting new character object into database:", err)
		return character, err
	}

	if err == nil {
		user.CharacterIds = append(user.CharacterIds, character.Id)
		CommitUser(session, *user)

		if err != nil {
			fmt.Println("Error updating user with new character data:", err)
		}
	}

	return character, err
}

func GetCharacters(session *mgo.Session, user User) []Character {
	var characters []Character
	for _, charId := range user.CharacterIds {
		character, err := GetCharacter(session, charId)

		if err != nil {
			fmt.Printf("Failed to find character with id %s, belonging to user %s: %s\n", charId, user.Name, err)
		} else {
			characters = append(characters, character)
		}
	}

	return characters
}

func DeleteRoom(session *mgo.Session, room Room) error {
	c := getCollection(session, cRooms)
	return c.RemoveId(room.Id)
}

func DeleteUser(session *mgo.Session, user User) error {
	for _, charId := range user.CharacterIds {
		DeleteCharacter(session, &user, charId)
	}

	c := getCollection(session, cUsers)
	return c.RemoveId(user.Id)
}

func removeId(idToRemove bson.ObjectId, ids []bson.ObjectId) []bson.ObjectId {
	length := 0
	result := ids

	for _, id := range ids {
		if id != idToRemove {
			result[length] = id
			length++
		}
	}

	return result[:length]
}

func DeleteCharacter(session *mgo.Session, user *User, charId bson.ObjectId) error {
	user.CharacterIds = removeId(charId, user.CharacterIds)
	err := CommitUser(session, *user)

	if err != nil {
		return err
	}

	c := getCollection(session, cCharacters)
	err = c.Remove(bson.M{fId: charId})

	return err
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

func commitObject(session *mgo.Session, c *mgo.Collection, id bson.ObjectId, object interface{}) error {
	_, err := c.UpsertId(id, object)
	printError(err)
	return err
}

func CommitUser(session *mgo.Session, user User) error {
	return commitObject(session, getCollection(session, cUsers), user.Id, user)
}

func CommitRoom(session *mgo.Session, room Room) error {
	return commitObject(session, getCollection(session, cRooms), room.Id, room)
}

func CommitCharacter(session *mgo.Session, character Character) error {
	return commitObject(session, getCollection(session, cCharacters), character.Id, character)
}

func CommitZone(session *mgo.Session, zone Zone) error {
	return commitObject(session, getCollection(session, cZones), zone.Id, zone)
}

// vim: nocindent
