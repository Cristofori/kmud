package engine

import (
	"container/list"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"mud/database"
	"mud/utils"
	"sync"
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
	Users      map[bson.ObjectId]database.User
	Characters map[bson.ObjectId]database.Character
	Rooms      map[bson.ObjectId]database.Room
}

var _model globalModel
var _session *mgo.Session

var _mutex sync.Mutex

var eventQueueChannel chan Event

func StartUp(session *mgo.Session) error {
	_session = session
	_model = globalModel{}

	_model.Users = map[bson.ObjectId]database.User{}
	_model.Characters = map[bson.ObjectId]database.Character{}
	_model.Rooms = map[bson.ObjectId]database.Room{}

	users, err := database.GetAllUsers(session)
	utils.HandleError(err)

	for _, user := range users {
		_model.Users[user.Id] = user
	}

	characters, err := database.GetAllCharacters(session)
	utils.HandleError(err)

	for _, character := range characters {
		_model.Characters[character.Id] = character
	}

	rooms, err := database.GetAllRooms(session)
	utils.HandleError(err)

	for _, room := range rooms {
		_model.Rooms[room.Id] = room
	}

	// Start the event loop
	eventQueueChannel = make(chan Event, 100)
	go eventLoop()

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

	queueEvent(RoomUpdateEvent{room})

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

	oldRoomId := character.RoomId

	newLocation := room.Location.Next(direction)
	room, found := GetRoomByLocation(newLocation)

	if !found {
		fmt.Printf("No room found at location %v, creating a new one (%s)\n", newLocation, character.PrettyName())
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

	if err == nil {
		queueEvent(EnterEvent{Character: character, RoomId: room.Id})
		queueEvent(LeaveEvent{Character: character, RoomId: oldRoomId})
	}

	return character, room, err
}

func GetUserByName(username string) (database.User, error) {
	_mutex.Lock()
	defer _mutex.Unlock()

	for _, user := range _model.Users {
		if user.Name == username {
			return user, nil
		}
	}

	return database.User{}, newEngineError("User not found")
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

func BroadcastMessage(from database.Character, message string) {
	msg := from.PrettyName() + ": " + message
	queueEvent(MessageEvent{msg})
}

func queueEvent(event Event) {
	eventQueueChannel <- event
}

func eventLoop() {
	var m sync.Mutex
	cond := sync.NewCond(&m)

	eventQueue := list.New()

	go func() {
		for {
			event := <-eventQueueChannel

			cond.L.Lock()
			eventQueue.PushBack(event)
			cond.L.Unlock()
			cond.Signal()
		}
	}()

	for {
		cond.L.Lock()
		for eventQueue.Len() == 0 {
			cond.Wait()
		}

		event := eventQueue.Remove(eventQueue.Front())
		cond.L.Unlock()

		broadcast(event.(Event))
	}
}

func CharactersIn(room database.Room, except database.Character) *list.List {
	_mutex.Lock()
	defer _mutex.Unlock()

	charList := list.New()

	for _, char := range _model.Characters {
		if char.RoomId == room.Id && char.Id != except.Id && char.Online() {
			charList.PushBack(char)
		}
	}

	return charList
}

func Login(user database.User) error {
	_mutex.Lock()
	defer _mutex.Unlock()

	if user.Online() {
		return newEngineError("That user is already online")
	}

	user.SetOnline(true)
	_model.Users[user.Id] = user

	return nil
}

func Logout(user database.User) {
	_mutex.Lock()
	defer _mutex.Unlock()

	user.SetOnline(false)
	_model.Users[user.Id] = user
}

func CreateUser(username string) (database.User, error) {
	_mutex.Lock()
	defer _mutex.Unlock()

	user, err := database.CreateUser(_session, username)

	if err == nil {
		user.SetOnline(true)
		_model.Users[user.Id] = user
	}

	return user, err
}

// vim: nocindent
