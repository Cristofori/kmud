package database

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"mud/utils"
)

func useFmt() { fmt.Printf("") }

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
	fTitle       = "title"
	fDescription = "description"
	fNorth       = "exit_north"
	fEast        = "exit_east"
	fSouth       = "exit_south"
	fWest        = "exit_west"
	fUp          = "exit_up"
	fDown        = "exit_down"
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

func GetCharacterRoom(session *mgo.Session, character string) (Room, error) {
	c := getCollection(session, cCharacters)
	q := c.Find(bson.M{fName: character})

	charResult := map[string]string{}
	err := q.One(&charResult)

	var room Room
	if err != nil {
		return room, err
	}

	c = getCollection(session, cRooms)
	q = c.Find(bson.M{fId: charResult[fRoom]})

	count, err := q.Count()

	if err != nil {
		return room, err
	}

	if count == 0 {
		SetCharacterRoom(session, character, "1")
		q = c.Find(bson.M{fId: "1"})
	}

	roomResult := map[string]string{}
	err = q.One(&roomResult)

	if err != nil {
		return room, err
	}

	room.Id = roomResult[fId]
	room.Title = roomResult[fTitle]
	room.Description = roomResult[fDescription]

	room.Exits = map[ExitDirection]bool{}
	room.Exits[DirectionNorth] = roomResult[fNorth] == "true"
	room.Exits[DirectionEast] = roomResult[fEast] == "true"
	room.Exits[DirectionSouth] = roomResult[fSouth] == "true"
	room.Exits[DirectionWest] = roomResult[fWest] == "true"
	room.Exits[DirectionUp] = roomResult[fUp] == "true"
	room.Exits[DirectionDown] = roomResult[fDown] == "true"

	return room, nil
}

func SetCharacterRoom(session *mgo.Session, character string, roomId string) error {
	c := getCollection(session, cCharacters)
	return c.Update(bson.M{fName: character}, bson.M{SET: bson.M{fRoom: roomId}})
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

func GenerateDefaultMap(session *mgo.Session) {
	c := getCollection(session, cRooms)
	c.DropCollection()

	c.Insert(bson.M{fId: "1",
		fTitle: "The Void",
		fDescription: "You are floating in the blackness of space. Complete darkness surrounds " +
			"you in all directions. There is no escape, there is no hope, just the emptiness. " +
			"You are likely to be eaten by a grue."})
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

func SetRoomExitEnabled(session *mgo.Session, room Room, direction ExitDirection, enabled bool) error {
	c := getCollection(session, cRooms)
	directionField := directionToFieldName(direction)

	value := "false"
	if enabled {
		value = "true"
	}

	return c.Update(bson.M{fId: room.Id}, bson.M{SET: bson.M{directionField: value}})
}

// vim: nocindent
