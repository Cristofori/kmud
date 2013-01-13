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

// UpdateUser updates the user in the model with user's Id, replacing it with
// the one that's given. If the given user doesn't exist in the model it will
// be added to. Also takes care of updating the database.
func (self *globalModel) UpdateUser(user database.User) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.users[user.Id] = user

	utils.HandleError(database.CommitUser(self.session, user))
}

// GetCharacter returns the Character object associated the given Id
func (self *globalModel) GetCharacter(id bson.ObjectId) database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return self.characters[id]
}

// GetCharacaterByName searches for a character with the given name. Returns a
// character object along with whether or not it was found in the model
func (self *globalModel) GetCharacterByName(name string) (database.Character, bool) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	name = utils.Simplify(name)

	for _, character := range self.characters {
		if character.Name == name {
			return character, true
		}
	}

	return database.Character{}, false
}

// GetUserCharacters returns all of the Character objects associated with the
// given user id
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

// CharactersIn returns a list of characters that are in the given room,
// excluding the character passed in as the "except" parameter. Returns all
// character type objects, including players, NPCs and MOBs
func (self *globalModel) CharactersIn(roomId bson.ObjectId, except bson.ObjectId) []database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var charList []database.Character

	for _, char := range self.characters {
		if char.RoomId == roomId && char.Id != except && char.Online() {
			charList = append(charList, char)
		}
	}

	return charList
}

// NpcsIn returns all of the NPC characters that are in the given room
func (self *globalModel) NpcsIn(roomId bson.ObjectId) []database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var npcList []database.Character

	for _, char := range self.characters {
		if char.RoomId == roomId && char.IsNpc() {
			npcList = append(npcList, char)
		}
	}

	return npcList
}

// GetOnlineCharacters returns a list of all of the characters who are online
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

// UpdateCharacter updates the character in the model with character's Id,
// replacing it with the one that's given. If the given character doesn't exist
// in the model it will be added to. Also takes care of updating the database.
func (self *globalModel) UpdateCharacter(character database.Character) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.characters[character.Id] = character

	utils.HandleError(database.CommitCharacter(self.session, character))
}

// DeleteCharacter removes the character associated with the given id from the
// model and from the database
func (self *globalModel) DeleteCharacter(id bson.ObjectId) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	delete(self.characters, id)

	utils.HandleError(database.DeleteCharacter(self.session, id))
}

// UpdateRoom updates the room in the model with room's Id, replacing it with
// the one that's given. If the given room doesn't exist in the model it will
// be added to. Also takes care of updating the database.
func (self *globalModel) UpdateRoom(room database.Room) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.rooms[room.Id] = room
	queueEvent(RoomUpdateEvent{room})

	utils.HandleError(database.CommitRoom(self.session, room))
}

// UpdateZone updates the zone in the model with zone's Id, replacing it with
// the one that's given. If the given zone doesn't exist in the model it will
// be added to. Also takes care of updating the database.
func (self *globalModel) UpdateZone(zone database.Zone) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.zones[zone.Id] = zone

	utils.HandleError(database.CommitZone(self.session, zone))
}

// GetRoom returns the room object associated with the given id
func (self *globalModel) GetRoom(id bson.ObjectId) database.Room {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return self.rooms[id]
}

// GetRooms returns a list of all of the rooms in the entire model
func (self *globalModel) GetRooms() []database.Room {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var rooms []database.Room

	for _, room := range self.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

// GetRoomByLocation searches for the room associated with the given coordinate
// in the given zone.  Returns a room object and whether or not it was found. 
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

// GetZone returns the zone object associated with the given id
func (self *globalModel) GetZone(zoneId bson.ObjectId) database.Zone {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return self.zones[zoneId]
}

// GetZones returns all of the zones in the model
func (self *globalModel) GetZones() []database.Zone {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var zones []database.Zone

	for _, zone := range self.zones {
		zones = append(zones, zone)
	}

	return zones
}

// GetZoneByName name searches for a zone with the given name, returns a zone
// object and whether or not it was found
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

// GetUser returns the User object associated with the given id
func (self *globalModel) GetUser(id bson.ObjectId) database.User {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return self.users[id]
}

// GetUsers returns all of the User objects in the model
func (self *globalModel) GetUsers() []database.User {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	var users []database.User

	for _, user := range self.users {
		users = append(users, user)
	}

	return users
}

// GetUserByName searches for the User object with the given name. Returns the
// User and whether or not a match was found.
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

// Removes the User assocaited with the given id from the model. Removes it from the database as well
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

// M is the global model object. All functions are thread-safe and all changes
// made to the model are automatically saved to the database.
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
	newRoom, found := M.GetRoomByLocation(newLocation, room.ZoneId)

	if !found {
		zone := M.GetZone(room.ZoneId)
		fmt.Printf("No room found at location %v%v, creating a new one (%s)\n", zone.Name, newLocation, character.PrettyName())

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
	} else {
		room = newRoom
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
	queueEvent(BroadcastEvent{from, message})
}

func Tell(from database.Character, to database.Character, message string) {
	queueEvent(TellEvent{from, to, message})
}

func Say(from database.Character, message string) {
	queueEvent(SayEvent{from, message})
}

func Emote(from database.Character, message string) {
	queueEvent(EmoteEvent{from, message})
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
