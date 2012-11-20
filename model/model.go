package model

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
	users      map[bson.ObjectId]database.User
	characters map[bson.ObjectId]database.Character
	rooms      map[bson.ObjectId]database.Room
	zones      map[bson.ObjectId]database.Zone

	mutex   sync.Mutex
	session *mgo.Session
}

func (self *globalModel) UpdateUser(user database.User) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.users[user.Id] = user

	utils.HandleError(database.CommitUser(self.session, user))
}

func (self *globalModel) GetCharacter(id bson.ObjectId) database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return self.characters[id]
}

func (self *globalModel) GetCharacterByName(name string) (database.Character, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for _, character := range self.characters {
		if character.Name == name {
			return character, true
		}
	}

	return database.Character{}, false
}

func (self *globalModel) GetUserCharacters(userId bson.ObjectId) []database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var characters []database.Character

	for _, character := range self.characters {
		if character.UserId == userId {
			characters = append(characters, character)
		}
	}

	return characters
}

func (self *globalModel) CharactersIn(room database.Room, except database.Character) []database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var charList []database.Character

	for _, char := range self.characters {
		if char.RoomId == room.Id && char.Id != except.Id && char.Online() {
			charList = append(charList, char)
		}
	}

	return charList
}

func (self *globalModel) GetOnlineCharacters() []database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var characters []database.Character

	for _, char := range self.characters {
		if char.Online() {
			characters = append(characters, char)
		}
	}

	return characters
}

func (self *globalModel) UpdateCharacter(character database.Character) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.characters[character.Id] = character

	utils.HandleError(database.CommitCharacter(self.session, character))
}

func (self *globalModel) DeleteCharacter(id bson.ObjectId) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	delete(self.characters, id)

	utils.HandleError(database.DeleteCharacter(self.session, id))
}

func (self *globalModel) UpdateRoom(room database.Room) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.rooms[room.Id] = room
	queueEvent(RoomUpdateEvent{room})

	utils.HandleError(database.CommitRoom(self.session, room))
}

func (self *globalModel) UpdateZone(zone database.Zone) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.zones[zone.Id] = zone

	utils.HandleError(database.CommitZone(self.session, zone))
}

func (self *globalModel) GetRoom(id bson.ObjectId) database.Room {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return self.rooms[id]
}

func (self *globalModel) GetRooms() []database.Room {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var rooms []database.Room

	for _, room := range self.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

func (self *globalModel) GetRoomByLocation(coordinate database.Coordinate, zoneId bson.ObjectId) (database.Room, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	ret := database.Room{}
	found := false

	for _, room := range self.rooms {
		if room.Location == coordinate && room.ZoneId == zoneId {
			found = true
			ret = room
			break
		}
	}

	return ret, found
}

func (self *globalModel) GetZone(zoneId bson.ObjectId) database.Zone {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return self.zones[zoneId]
}

func (self *globalModel) GetZones() []database.Zone {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var zones []database.Zone

	for _, zone := range self.zones {
		zones = append(zones, zone)
	}

	return zones
}

func (self *globalModel) GetZoneByName(name string) (database.Zone, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for _, zone := range self.zones {
		if zone.Name == name {
			return zone, true
		}
	}

	return database.Zone{}, false
}

func (self *globalModel) deleteRoom(id bson.ObjectId) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	delete(self.rooms, id)

	utils.HandleError(database.DeleteRoom(self.session, id))
}

func (self *globalModel) GetUser(id bson.ObjectId) database.User {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return self.users[id]
}

func (self *globalModel) GetUsers() []database.User {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var users []database.User

	for _, user := range self.users {
		users = append(users, user)
	}

	return users
}

func (self *globalModel) GetUserByName(username string) (database.User, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for _, user := range self.users {
		if user.Name == username {
			return user, true
		}
	}

	return database.User{}, false
}

func (self *globalModel) DeleteUser(id bson.ObjectId) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for _, character := range self.characters {
		if character.UserId == id {
			delete(self.characters, character.Id)
			utils.HandleError(database.DeleteCharacter(self.session, character.Id))
		}
	}

	delete(self.users, id)

	utils.HandleError(database.DeleteUser(self.session, id))
}

var M globalModel
var eventQueueChannel chan Event

/**
 * Initializes the global model object and starts up the main event loop
 */
func Init(session *mgo.Session) error {
	M = globalModel{}

	M.session = session

	M.users = map[bson.ObjectId]database.User{}
	M.characters = map[bson.ObjectId]database.Character{}
	M.rooms = map[bson.ObjectId]database.Room{}
	M.zones = map[bson.ObjectId]database.Zone{}

	users, err := database.GetAllUsers(session)
	utils.HandleError(err)

	for _, user := range users {
		M.users[user.Id] = user
	}

	characters, err := database.GetAllCharacters(session)
	utils.HandleError(err)

	for _, character := range characters {
		M.characters[character.Id] = character
	}

	rooms, err := database.GetAllRooms(session)
	utils.HandleError(err)

	for _, room := range rooms {
		M.rooms[room.Id] = room
	}

	zones, err := database.GetAllZones(session)
	utils.HandleError(err)

	for _, zone := range zones {
		M.zones[zone.Id] = zone
	}

	// Start the event loop
	eventQueueChannel = make(chan Event, 100)
	go eventLoop()

	return err
}

func MoveCharacterToLocation(character *database.Character, location database.Coordinate) (database.Room, error) {
	oldRoom := M.GetRoom(character.RoomId)

	newRoom, found := M.GetRoomByLocation(location, oldRoom.ZoneId)

	if !found {
		return newRoom, errors.New("Invalid location")
	}

	character.RoomId = newRoom.Id

	M.UpdateCharacter(*character)

	queueEvent(EnterEvent{Character: *character, RoomId: newRoom.Id})
	queueEvent(LeaveEvent{Character: *character, RoomId: oldRoom.Id})

	return newRoom, nil
}

func MoveCharacterToRoom(character *database.Character, newRoom database.Room) {
	oldRoomId := character.RoomId
	character.RoomId = newRoom.Id

	M.UpdateCharacter(*character)

	queueEvent(EnterEvent{Character: *character, RoomId: newRoom.Id})
	queueEvent(LeaveEvent{Character: *character, RoomId: oldRoomId})
}

func MoveCharacter(character *database.Character, direction database.ExitDirection) (database.Room, error) {
	room := M.GetRoom(character.RoomId)

	if room.Id == "" {
		return room, errors.New("Character doesn't appear to be in any room")
	}

	if !room.HasExit(direction) {
		return room, errors.New("Attempted to move through an exit that the room does not contain")
	}

	newLocation := room.Location.Next(direction)
	room, found := M.GetRoomByLocation(newLocation, room.ZoneId)

	if !found {
		fmt.Printf("No room found at location %v, creating a new one (%s)\n", newLocation, character.PrettyName())

		room = database.NewRoom(room.ZoneId)

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
		M.UpdateRoom(room)
	}

	return MoveCharacterToLocation(character, room.Location)
}

func DeleteRoom(room database.Room) {
	M.deleteRoom(room.Id)

	// Disconnect all exits leading to this room
	loc := room.Location

	updateRoom := func(dir database.ExitDirection) {
		next := loc.Next(dir)
		room, found := M.GetRoomByLocation(next, room.ZoneId)

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
			M.UpdateRoom(room)
		}
	}

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
}

func BroadcastMessage(from database.Character, message string) {
	queueEvent(MessageEvent{from, message})
}

func Say(from database.Character, message string) {
	queueEvent(SayEvent{from, message})
}

func queueEvent(event Event) {
	eventQueueChannel <- event // TODO: Function not likely thread-safe
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

func Login(user database.User) error {
	if user.Online() {
		return errors.New("That user is already online")
	}

	user.SetOnline(true)
	M.UpdateUser(user) // TODO: Avoid unnecessary database call

	return nil
}

func Logout(user database.User) {
	user.SetOnline(false)
	M.UpdateUser(user) // TODO: Avoid unnecessary database call
}

/**
 * Returns cordinates that indiate the highest and lowest points of
 * the map in 3 dimensions
 */
func ZoneCorners(zoneId bson.ObjectId) (database.Coordinate, database.Coordinate) {
	var top int
	var bottom int
	var left int
	var right int
	var high int
	var low int

	rooms := M.GetRooms()

	for _, room := range rooms {
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

	for _, room := range rooms {
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

func MoveRoomsToZone(fromZoneId bson.ObjectId, toZoneId bson.ObjectId) {
	for _, room := range M.GetRooms() {
		if room.ZoneId == fromZoneId {
			room.ZoneId = toZoneId

			M.UpdateRoom(room)
		}
	}
}

// vim: nocindent
