package engine

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"mud/database"
	"mud/utils"
	"sync"
	"time"
)

type engineError struct {
	message string
}

func (self engineError) Error() string {
	return self.message
}

func newEngineError(message string) error {
	return engineError{message: message}
}

type globalModel struct {
	Characters map[bson.ObjectId]database.Character
	Rooms      map[bson.ObjectId]database.Room
}

var _model globalModel
var _session *mgo.Session

// TODO Use a read/write mutex
var _mutex sync.Mutex

func spewEvents() {
	for {
		fmt.Printf("Broadcasting...\n")
		broadcast("Here's an event")
		time.Sleep(5 * time.Second)
	}
}

func StartUp(session *mgo.Session) error {
	_session = session
	_model = globalModel{}

	_model.Characters = map[bson.ObjectId]database.Character{}
	_model.Rooms = map[bson.ObjectId]database.Room{}

	rooms, err := database.GetAllRooms(session)
	utils.HandleError(err)

	for _, room := range rooms {
		_model.Rooms[room.Id] = room
	}

	characters, err := database.GetAllCharacters(session)
	utils.HandleError(err)

	for _, character := range characters {
		_model.Characters[character.Id] = character
	}

	go spewEvents()

	return err
}

func UpdateCharacter(character database.Character) error {
	_mutex.Lock()
	defer _mutex.Unlock()

	_model.Characters[character.Id] = character
	return database.CommitCharacter(_session, character)
}

func UpdateRoom(room database.Room) error {
	_mutex.Lock()
	defer _mutex.Unlock()

	_model.Rooms[room.Id] = room
	return database.CommitRoom(_session, room)
}

func AddRoom(room database.Room) error {
	_mutex.Lock()
	_model.Rooms[room.Id] = room
	_mutex.Unlock()

	return UpdateRoom(room)
}

func MoveCharacter(character database.Character, direction database.ExitDirection) (database.Character, database.Room, error) {
	_mutex.Lock()
	room := _model.Rooms[character.RoomId]
	_mutex.Unlock()

	if room.Id == "" {
		return character, room, newEngineError("Character doesn't appear to be in any room")
	}

	if !room.HasExit(direction) {
		return character, room, newEngineError("Attempted to move through an exit that the room does not contain")
	}

	newLocation := room.Location.Next(direction)
	room, found := GetRoomByLocation(newLocation)

	if !found {
		fmt.Printf("No room found at location %v, creating a new one\n", newLocation)
		room = database.NewRoom()

		switch direction {
		case database.DirectionNorth:
			room.ExitSouth = true
		case database.DirectionEast:
			room.ExitWest = true
		case database.DirectionSouth:
			room.ExitNorth = true
		case database.DirectionWest:
			room.ExitEast = true
		case database.DirectionUp:
			room.ExitDown = true
		case database.DirectionDown:
			room.ExitUp = true
		default:
			panic("Unexpected code path")
		}

		room.Location = newLocation
		err := AddRoom(room)
		utils.HandleError(err)
	}

	character.RoomId = room.Id
	err := UpdateCharacter(character)

	utils.HandleError(err)

	return character, room, err
}

func GetCharacterRoom(character database.Character) database.Room {
	_mutex.Lock()
	defer _mutex.Unlock()

	return _model.Rooms[character.RoomId]
}

func GetRoomByLocation(coordinate database.Coordinate) (database.Room, bool) {
	_mutex.Lock()
	defer _mutex.Unlock()

	ret := database.Room{}
	found := false

	for _, room := range _model.Rooms {
		if room.Location == coordinate {
			found = true
			ret = room
			break
		}
	}

	return ret, found
}

func GenerateDefaultMap() {
	_mutex.Lock()
	{
		_model.Rooms = map[bson.ObjectId]database.Room{}
		database.DeleteAllRooms(_session)
	}
	_mutex.Unlock()

	room := database.NewRoom()
	room.Location = database.Coordinate{0, 0, 0}
	room.Default = true

	AddRoom(room)
}

// vim: nocindent
