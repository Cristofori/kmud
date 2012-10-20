package database

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"mud/utils"
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
	fId          = "_id"
	fName        = "name"
	fCharacters  = "characters"
	fRoom        = "room"
	fLocation    = "location"
	fTitle       = "title"
	fDescription = "description"
	fNorth       = "exit_north"
	fEast        = "exit_east"
	fSouth       = "exit_south"
	fWest        = "exit_west"
	fUp          = "exit_up"
	fDown        = "exit_down"
	fDefault     = "default"
)

// DB commands
const (
	SET  = "$set"
	PUSH = "$push"
	PULL = "$pull"
)

func FindUser(session *mgo.Session, name string) (bool, error) {
	c := getCollection(session, cUsers)
	q := c.Find(bson.M{fName: name})

	count, err := q.Count()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func FindCharacter(session *mgo.Session, name string) (bool, error) {
	c := getCollection(session, cCharacters)
	q := c.Find(bson.M{fName: name})

	count, err := q.Count()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func NewUser(session *mgo.Session, name string) error {

	found, err := FindUser(session, name)

	if err != nil {
		return err
	}

	if found {
		return newDbError("That user already exists")
	}

	c := getCollection(session, cUsers)
	c.Insert(bson.M{fName: name})

	return nil
}

func NewCharacter(session *mgo.Session, user string, character string) error {

	found, err := FindCharacter(session, character)

	if err != nil {
		return err
	}

	if found {
		return newDbError("That character already exists")
	}

	c := getCollection(session, cUsers)
	c.Update(bson.M{fName: user}, bson.M{PUSH: bson.M{fCharacters: character}})

	c = getCollection(session, cCharacters)
	c.Insert(bson.M{fName: character})

	return nil
}

func GetCharacterRoom(session *mgo.Session, charName string) (Room, error) {
	charCollection := getCollection(session, cCharacters)
	q := charCollection.Find(bson.M{fName: charName})

	var character Character
	err := q.One(&character)

	var room Room
	if err != nil {
		return room, err
	}

	roomCollection := getCollection(session, cRooms)
	q = roomCollection.Find(bson.M{fId: character.Room})

	count, err := q.Count()

	if err != nil {
		return room, err
	}

	if count == 0 {
		room, err = StartingRoom(session)

		if err == nil {
			SetCharacterRoom(session, charName, room.Id)
		}
	} else {
		err = q.One(&room)
	}

	return room, err
}

func GetRoomByLocation(session *mgo.Session, location Coordinate) (Room, error) {
	c := getCollection(session, cRooms)
	q := c.Find(bson.M{fLocation: location})

	count, err := q.Count()

	var room Room
	if err != nil {
		return room, err
	}

	if count == 0 {
		return room, newDbError("Room not found")
	}

	err = q.One(&room)

	return room, err
}

func CreateRoom(session *mgo.Session, room Room) (bson.ObjectId, error) {
	c := getCollection(session, cRooms)
	err := c.Insert(room)

	if err != nil {
		fmt.Printf("Error creating room: %v\n", err)
		return "", err
	}

	room, err = GetRoomByLocation(session, room.Location)

	return room.Id, err
}

func SetCharacterRoom(session *mgo.Session, character string, roomId bson.ObjectId) error {
	c := getCollection(session, cCharacters)
	err := c.Update(bson.M{fName: character}, bson.M{SET: bson.M{fRoom: roomId}})

	if err != nil {
		fmt.Printf("Failed setting character room :%v\n", err)
	}

	return err
}

func GetUserCharacters(session *mgo.Session, name string) ([]string, error) {
	c := getCollection(session, cUsers)
	q := c.Find(bson.M{fName: name})

	result := map[string][]string{}
	err := q.One(&result)

	return result[fCharacters], err
}

func DeleteCharacter(session *mgo.Session, user string, character string) error {
	c := getCollection(session, cUsers)
	c.Update(bson.M{fName: user}, bson.M{PULL: bson.M{fCharacters: utils.Simplify(character)}})

	c = getCollection(session, cCharacters)
	c.Remove(bson.M{fName: character})

	return nil
}

func defaultRoom() Room {
	var room Room
	room.Id = ""
	room.Title = "The Void"
	room.Description = "You are floating in the blackness of space. Complete darkness surrounds " +
		"you in all directions. There is no escape, there is no hope, just the emptiness. " +
		"You are likely to be eaten by a grue."

	room.ExitNorth = false
	room.ExitEast = false
	room.ExitSouth = false
	room.ExitWest = false
	room.ExitUp = false
	room.ExitDown = false

	room.Location = Coordinate{0, 0, 0}

	room.Default = false

	return room
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

	room := defaultRoom()
	room.Location = Coordinate{0, 0, 0}
	room.Default = true

	CreateRoom(session, room)
}

func SetRoomTitle(session *mgo.Session, room Room, title string) error {
	c := getCollection(session, cRooms)
	return c.Update(bson.M{fId: room.Id}, bson.M{SET: bson.M{fTitle: title}})
}

func SetRoomDescription(session *mgo.Session, room Room, description string) error {
	c := getCollection(session, cRooms)
	return c.Update(bson.M{fId: room.Id}, bson.M{SET: bson.M{fDescription: description}})
}

func directionToFieldName(direction ExitDirection) string {
	switch direction {
	case DirectionNorth:
		return fNorth
	case DirectionEast:
		return fEast
	case DirectionSouth:
		return fSouth
	case DirectionWest:
		return fWest
	case DirectionUp:
		return fUp
	case DirectionDown:
		return fDown
	}

	// Wouldn't ever expect DirectionNone to be passed here
	panic("Unexpected code path")
}

func CommitRoom(session *mgo.Session, room Room) error {
	c := getCollection(session, cRooms)
	return c.Update(bson.M{fId: room.Id}, room)
}

func MoveCharacter(session *mgo.Session, character string, direction ExitDirection) (Room, error) {
	room, err := GetCharacterRoom(session, character)

	if err != nil {
		return room, err
	}

	newLocation := room.Location

	switch direction {
	case DirectionNorth:
		newLocation.Y -= 1
	case DirectionEast:
		newLocation.X += 1
	case DirectionSouth:
		newLocation.Y += 1
	case DirectionWest:
		newLocation.X -= 1
	case DirectionUp:
		newLocation.Z -= 1
	case DirectionDown:
		newLocation.Z += 1
	default:
		panic("Unexpected code path")
	}

	newRoom, err := GetRoomByLocation(session, newLocation)
	var newRoomId bson.ObjectId

	if err == nil {
		newRoomId = newRoom.Id
	} else {
		fmt.Printf("No room found at location %v, creating a new one\n", newLocation)
		newRoom = defaultRoom()

		switch direction {
		case DirectionNorth:
			newRoom.ExitSouth = true
		case DirectionEast:
			newRoom.ExitWest = true
		case DirectionSouth:
			newRoom.ExitNorth = true
		case DirectionWest:
			newRoom.ExitEast = true
		case DirectionUp:
			newRoom.ExitDown = true
		case DirectionDown:
			newRoom.ExitUp = true
		default:
			panic("Unexpected code path")
		}

		newRoom.Location = newLocation
		newRoomId, err = CreateRoom(session, newRoom)
	}

	if err == nil {
		err = SetCharacterRoom(session, character, newRoomId)
	}

	return newRoom, err
}

// vim: nocindent
