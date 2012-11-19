package engine

import (
	"container/list"
	"errors"
	"fmt"
	"kmud/database"
	"kmud/utils"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"sync"
)

type globalModel struct {
	Users      map[bson.ObjectId]database.User
	Characters map[bson.ObjectId]database.Character
	Rooms      map[bson.ObjectId]database.Room
	Zones      map[bson.ObjectId]database.Zone
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
	_model.Zones = map[bson.ObjectId]database.Zone{}

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

	zones, err := database.GetAllZones(session)
	utils.HandleError(err)

	for _, zone := range zones {
		_model.Zones[zone.Id] = zone
	}

	// Start the event loop
	eventQueueChannel = make(chan Event, 100)
	go eventLoop()

	return err
}

func UpdateUser(user database.User) error {
	_mutex.Lock()
	defer _mutex.Unlock()

	_model.Users[user.Id] = user
	return database.CommitUser(_session, user)
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

func UpdateZone(zone database.Zone) error {
	_mutex.Lock()
	defer _mutex.Unlock()

	_model.Zones[zone.Id] = zone

	return database.CommitZone(_session, zone)
}

func MoveCharacterToLocation(character *database.Character, location database.Coordinate) (database.Room, error) {
	oldRoomId := character.RoomId

	_mutex.Lock()
	oldRoom := _model.Rooms[oldRoomId]
	_mutex.Unlock()

	newRoom, found := GetRoomByLocation(location, oldRoom.ZoneId)

	if !found {
		return newRoom, errors.New("Invalid location")
	}

	character.RoomId = newRoom.Id

	err := UpdateCharacter(*character)

	utils.HandleError(err)

	if err == nil {
		queueEvent(EnterEvent{Character: *character, RoomId: newRoom.Id})
		queueEvent(LeaveEvent{Character: *character, RoomId: oldRoomId})
	}

	return newRoom, err
}

func MoveCharacterToRoom(character *database.Character, newRoom database.Room) error {
	oldRoomId := character.RoomId
	character.RoomId = newRoom.Id

	err := UpdateCharacter(*character)

	utils.HandleError(err)

	if err == nil {
		queueEvent(EnterEvent{Character: *character, RoomId: newRoom.Id})
		queueEvent(LeaveEvent{Character: *character, RoomId: oldRoomId})
	}

	return err
}

func MoveCharacter(character *database.Character, direction database.ExitDirection) (database.Room, error) {
	_mutex.Lock()
	room := _model.Rooms[character.RoomId]
	_mutex.Unlock()

	if room.Id == "" {
		return room, errors.New("Character doesn't appear to be in any room")
	}

	if !room.HasExit(direction) {
		return room, errors.New("Attempted to move through an exit that the room does not contain")
	}

	newLocation := room.Location.Next(direction)
	room, found := GetRoomByLocation(newLocation, room.ZoneId)

	if !found {
		fmt.Printf("No room found at location %v, creating a new one (%s)\n", newLocation, character.PrettyName())

		_mutex.Lock()
		room = database.NewRoom(room.ZoneId)
		_mutex.Unlock()

		switch direction {
		case database.DirectionNorth:
			room.ExitSouth = true
		case database.DirectionNorthEast:
			room.ExitSouthWest = true
		case database.DirectionEast:
			room.ExitWest = true
		case database.DirectionSouthEast:
			room.ExitNorthWest = true
		case database.DirectionSouth:
			room.ExitNorth = true
		case database.DirectionSouthWest:
			room.ExitNorthEast = true
		case database.DirectionWest:
			room.ExitEast = true
		case database.DirectionNorthWest:
			room.ExitSouthEast = true
		case database.DirectionUp:
			room.ExitDown = true
		case database.DirectionDown:
			room.ExitUp = true
		default:
			panic("Unexpected code path")
		}

		room.Location = newLocation

		_mutex.Lock()
		err := UpdateRoom(room)
		_mutex.Unlock()

		utils.HandleError(err)
	}

	return MoveCharacterToLocation(character, room.Location)
}

func DeleteRoom(room database.Room) error {
	_mutex.Lock()

	err := database.DeleteRoom(_session, room)

	if err == nil {
		delete(_model.Rooms, room.Id)

		// Disconnect all exits leading to this room
		loc := room.Location

		updateRoom := func(dir database.ExitDirection) {
			next := loc.Next(dir)
			room, found := GetRoomByLocation(next, room.ZoneId)

			if found {
				var exitToDisable database.ExitDirection
				switch dir {
				case database.DirectionNorth:
					exitToDisable = database.DirectionSouth
				case database.DirectionNorthEast:
					exitToDisable = database.DirectionSouthWest
				case database.DirectionEast:
					exitToDisable = database.DirectionWest
				case database.DirectionSouthEast:
					exitToDisable = database.DirectionNorthWest
				case database.DirectionSouth:
					exitToDisable = database.DirectionNorth
				case database.DirectionSouthWest:
					exitToDisable = database.DirectionNorthEast
				case database.DirectionWest:
					exitToDisable = database.DirectionEast
				case database.DirectionNorthWest:
					exitToDisable = database.DirectionSouthEast
				case database.DirectionUp:
					exitToDisable = database.DirectionDown
				case database.DirectionDown:
					exitToDisable = database.DirectionUp
				}
				room.SetExitEnabled(exitToDisable, false)
				UpdateRoom(room)
			}
		}

		_mutex.Unlock()
		updateRoom(database.DirectionNorth)
		updateRoom(database.DirectionNorthEast)
		updateRoom(database.DirectionEast)
		updateRoom(database.DirectionSouthEast)
		updateRoom(database.DirectionSouth)
		updateRoom(database.DirectionSouthWest)
		updateRoom(database.DirectionWest)
		updateRoom(database.DirectionNorthWest)
		updateRoom(database.DirectionUp)
		updateRoom(database.DirectionDown)
	} else {
		_mutex.Unlock()
	}

	return err
}

func GetUser(id bson.ObjectId) database.User {
	_mutex.Lock()
	defer _mutex.Unlock()

	return _model.Users[id]
}

func GetUsers() []database.User {
	_mutex.Lock()
	defer _mutex.Unlock()

	var users []database.User

	for _, user := range _model.Users {
		users = append(users, user)
	}

	return users
}

func GetZones() []database.Zone {
	_mutex.Lock()
	defer _mutex.Unlock()

	var zones []database.Zone

	for _, zone := range _model.Zones {
		zones = append(zones, zone)
	}

	return zones
}

func GetUserByName(username string) (database.User, error) {
	_mutex.Lock()
	defer _mutex.Unlock()

	for _, user := range _model.Users {
		if user.Name == username {
			return user, nil
		}
	}

	return database.User{}, errors.New("User not found")
}

func GetCharacterRoom(character database.Character) database.Room {
	_mutex.Lock()
	defer _mutex.Unlock()

	return _model.Rooms[character.RoomId]
}

func GetCharacterUser(character database.Character) database.User {
	_mutex.Lock()
	defer _mutex.Unlock()

	for _, user := range _model.Users {
		for _, charId := range user.CharacterIds {
			if charId == character.Id {
				return user
			}
		}
	}

	return database.User{}

}

func GetRoomByLocation(coordinate database.Coordinate, zoneId bson.ObjectId) (database.Room, bool) {
	_mutex.Lock()
	defer _mutex.Unlock()

	ret := database.Room{}
	found := false

	for _, room := range _model.Rooms {
		if room.Location == coordinate && room.ZoneId == zoneId {
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

	room := database.NewRoom("")
	room.Location = database.Coordinate{0, 0, 0}
	room.Default = true

	UpdateRoom(room)
}

func BroadcastMessage(from database.Character, message string) {
	queueEvent(MessageEvent{from, message})
}

func Say(from database.Character, message string) {
	queueEvent(SayEvent{from, message})
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

func CharactersIn(room database.Room, except database.Character) []database.Character {
	_mutex.Lock()
	defer _mutex.Unlock()

	var charList []database.Character

	for _, char := range _model.Characters {
		if char.RoomId == room.Id && char.Id != except.Id && char.Online() {
			charList = append(charList, char)
		}
	}

	return charList
}

func Login(user database.User) error {
	_mutex.Lock()
	defer _mutex.Unlock()

	if user.Online() {
		return errors.New("That user is already online")
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
		_model.Users[user.Id] = user
	}

	return user, err
}

func CreateCharacter(user *database.User, charname string) (database.Character, error) {
	_mutex.Lock()
	defer _mutex.Unlock()

	character, err := database.CreateCharacter(_session, user, charname)

	if err == nil {
		_model.Users[user.Id] = *user
		_model.Characters[character.Id] = character
	}

	return character, err
}

func GetCharacter(id bson.ObjectId) database.Character {
	_mutex.Lock()
	defer _mutex.Unlock()

	return _model.Characters[id]
}

func GetCharacters(user database.User) []database.Character {
	_mutex.Lock()
	defer _mutex.Unlock()

	var characters []database.Character

	for _, charId := range user.CharacterIds {
		characters = append(characters, _model.Characters[charId])
	}

	return characters
}

func GetOnlineCharacters() []database.Character {
	_mutex.Lock()
	defer _mutex.Unlock()

	var characters []database.Character

	for _, char := range _model.Characters {
		if char.Online() {
			characters = append(characters, char)
		}
	}

	return characters
}

func DeleteCharacter(user *database.User, charId bson.ObjectId) error {
	_mutex.Lock()
	defer _mutex.Unlock()

	err := database.DeleteCharacter(_session, user, charId)

	if err == nil {
		delete(_model.Characters, charId)
		_model.Users[user.Id] = *user
	}

	return err
}

func DeleteUser(userId bson.ObjectId) error {
	_mutex.Lock()
	defer _mutex.Unlock()

	err := database.DeleteUser(_session, _model.Users[userId])

	if err == nil {
		delete(_model.Users, userId)
	}

	return err
}

/**
 * Returns cordinates that indiate the highest and lowest points of
 * the map in 3 dimensions
 */
func ZoneCorners(zoneId bson.ObjectId) (database.Coordinate, database.Coordinate) {
	_mutex.Lock()
	defer _mutex.Unlock()

	var top int
	var bottom int
	var left int
	var right int
	var high int
	var low int

	for _, room := range _model.Rooms {
		if room.ZoneId == zoneId {
			top = room.Location.Y
			bottom = room.Location.Y
			left = room.Location.X
			right = room.Location.X
			high = room.Location.Z
			low = room.Location.Z
			break
		}
	}

	for _, room := range _model.Rooms {
		if room.ZoneId != zoneId {
			continue
		}

		if room.Location.Z < high {
			high = room.Location.Z
		}

		if room.Location.Z > low {
			low = room.Location.Z
		}

		if room.Location.Y < top {
			top = room.Location.Y
		}

		if room.Location.Y > bottom {
			bottom = room.Location.Y
		}

		if room.Location.X < left {
			left = room.Location.X
		}

		if room.Location.X > right {
			right = room.Location.X
		}
	}

	return database.Coordinate{X: left, Y: top, Z: high},
		database.Coordinate{X: right, Y: bottom, Z: low}
}

func GetZone(zoneId bson.ObjectId) database.Zone {
	_mutex.Lock()
	defer _mutex.Unlock()
	return _model.Zones[zoneId]
}

func GetRoom(roomId bson.ObjectId) database.Room {
	_mutex.Lock()
	defer _mutex.Unlock()
	return _model.Rooms[roomId]
}

func GetZoneByName(name string) (database.Zone, bool) {
	_mutex.Lock()
	defer _mutex.Unlock()

	for _, zone := range _model.Zones {
		if zone.Name == name {
			return zone, true
		}
	}

	return database.Zone{}, false
}

func MoveRoomsToZone(fromZoneId bson.ObjectId, toZoneId bson.ObjectId) {
	_mutex.Lock()

	for _, room := range _model.Rooms {
		if room.ZoneId == fromZoneId {
			room.ZoneId = toZoneId

			_mutex.Unlock()
			UpdateRoom(room)
			_mutex.Lock()
		}
	}

	_mutex.Unlock()
}

// vim: nocindent
