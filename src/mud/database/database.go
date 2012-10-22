package database

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type dbError struct {
	message string
}

func (self dbError) Error() string {
	return self.message
}

func newDbError(message string) dbError {
	var err dbError
	err.message = message
	return err
}

type collectionName string

func getCollection(session *mgo.Session, collection collectionName) *mgo.Collection {
	return session.DB("mud").C(string(collection))
}

// Collection names
const (
	cUsers      = collectionName("users")
	cCharacters = collectionName("characters")
	cRooms      = collectionName("rooms")
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
		fmt.Printf("Error: %s\n", err)
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
		return newDbError(fmt.Sprintf("Query return no results: %v", query))
	}

	err = q.One(object)

	return err
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
		return character, newDbError("That character already exists")
	}

	startingRoom, err := StartingRoom(session)

	if err != nil {
		fmt.Printf("Error getting starting room: %v\n", err)
		return character, err
	}

	character = newCharacter(characterName)
	character.RoomId = startingRoom.Id

	err = CommitCharacter(session, character)

	if err != nil {
		fmt.Printf("Error inserting new character object into database: %v\n", err)
		return character, err
	}

	if err == nil {
		user.CharacterIds = append(user.CharacterIds, character.Id)
		CommitUser(session, *user)

		if err != nil {
			fmt.Printf("Error updating user with new character data: %v\n", err)
		}
	}

	return character, err
}

func GetUserCharacters(session *mgo.Session, user User) []Character {
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

func DeleteCharacter(session *mgo.Session, user *User, charId bson.ObjectId) error {
	// TODO - Figure out how to remove an element from the middle of a slice,
	//        and then just modify the user object and use CommitUser
	c := getCollection(session, cUsers)
	err := c.Update(bson.M{fId: user.Id}, bson.M{PULL: bson.M{fCharacterIds: charId}})

	if err != nil {
		fmt.Printf("Failed 1: %v %v\n", user.Id, charId)
		return err
	}

	modifiedUser, err := GetUser(session, user.Id)

	if err != nil {
		fmt.Printf("Failed 2\n")
		return err
	}

	user.CharacterIds = modifiedUser.CharacterIds

	c = getCollection(session, cCharacters)
	err = c.Remove(bson.M{fId: charId})

	if err != nil {
		fmt.Printf("Failed 3\n")
	}

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
		return room, newDbError("No default room found")
	}

	if count > 1 {
		fmt.Printf("Warning: More than one default room found\n")
	}

	err = q.One(&room)

	return room, err
}

func GenerateDefaultMap(session *mgo.Session) {
	c := getCollection(session, cRooms)
	c.DropCollection()

	room := newRoom()
	room.Location = Coordinate{0, 0, 0}
	room.Default = true

	CommitRoom(session, room)
}

func CreateUser(session *mgo.Session, name string) (User, error) {
	user, err := findUser(session, bson.M{fName: name})

	if err == nil {
		return user, newDbError("That user already exists")
	}

	user = newUser(name)
	err = CommitUser(session, user)
	return user, err
}

func CommitUser(session *mgo.Session, user User) error {
	c := getCollection(session, cUsers)
	_, err := c.UpsertId(user.Id, user)
	printError(err)
	return err
}

func CommitRoom(session *mgo.Session, room Room) error {
	c := getCollection(session, cRooms)
	_, err := c.UpsertId(room.Id, room)
	printError(err)
	return err
}

func CommitCharacter(session *mgo.Session, character Character) error {
	c := getCollection(session, cCharacters)
	_, err := c.UpsertId(character.Id, character)
	printError(err)
	return err
}

func MoveCharacter(session *mgo.Session, character *Character, direction ExitDirection) (Room, error) {
	room, err := GetRoom(session, character.RoomId)

	if err != nil {
		return room, err
	}

	newLocation := room.Location.Next(direction)
	room, err = GetRoomByLocation(session, newLocation)

	if err != nil {
		fmt.Printf("No room found at location %v, creating a new one\n", newLocation)
		room = newRoom()

		switch direction {
		case DirectionNorth:
			room.ExitSouth = true
		case DirectionEast:
			room.ExitWest = true
		case DirectionSouth:
			room.ExitNorth = true
		case DirectionWest:
			room.ExitEast = true
		case DirectionUp:
			room.ExitDown = true
		case DirectionDown:
			room.ExitUp = true
		default:
			panic("Unexpected code path")
		}

		room.Location = newLocation
		err = CommitRoom(session, room)
	}

	character.RoomId = room.Id
	err = CommitCharacter(session, *character)

	return room, err
}

// vim: nocindent
