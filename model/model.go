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
	users      map[bson.ObjectId]*database.User
	characters map[bson.ObjectId]*database.Character
	rooms      map[bson.ObjectId]*database.Room
	zones      map[bson.ObjectId]*database.Zone
	items      map[bson.ObjectId]*database.Item

	mutex sync.RWMutex
}

// CreateUser creates a new User object in the database and adds it to the model.
// A pointer to the new User object is returned.
func (self *globalModel) CreateUser(name string) *database.User {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	user := database.NewUser(name)
	self.users[user.GetId()] = user

	return user
}

// GetCharacter returns the Character object associated the given Id
func (self *globalModel) GetCharacter(id bson.ObjectId) *database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.characters[id]
}

// GetCharacaterByName searches for a character with the given name. Returns a
// character object, or nil if it wasn't found.
func (self *globalModel) GetCharacterByName(name string) *database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	name = utils.Simplify(name)

	for _, character := range self.characters {
		if character.GetName() == name {
			return character
		}
	}

	return nil
}

// GetUserCharacters returns all of the Character objects associated with the
// given user id
func (self *globalModel) GetUserCharacters(user *database.User) []*database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var characters []*database.Character

	for _, character := range self.characters {
		if character.GetUserId() == user.GetId() {
			characters = append(characters, character)
		}
	}

	return characters
}

// CharactersIn returns a list of characters that are in the given room (NPC or
// player), excluding the character passed in as the "except" parameter.
// Returns all character type objects, including players, NPCs and MOBs
func (self *globalModel) CharactersIn(room *database.Room) []*database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var charList []*database.Character

	for _, char := range self.characters {
		if char.GetRoomId() == room.GetId() && char.IsOnline() {
			charList = append(charList, char)
		}
	}

	return charList
}

// PlayersIn returns all of the player characters that are in the given room
func (self *globalModel) PlayersIn(room *database.Room, except *database.Character) []*database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var playerList []*database.Character

	for _, char := range self.characters {
		if char.GetRoomId() == room.GetId() && char.IsPlayer() && char.IsOnline() && char != except {
			playerList = append(playerList, char)
		}
	}

	return playerList
}

// NpcsIn returns all of the NPC characters that are in the given room
func (self *globalModel) NpcsIn(room *database.Room) []*database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var npcList []*database.Character

	for _, char := range self.characters {
		if char.GetRoomId() == room.GetId() && char.IsNpc() {
			npcList = append(npcList, char)
		}
	}

	return npcList
}

// GetOnlineCharacters returns a list of all of the characters who are online
func (self *globalModel) GetOnlineCharacters() []*database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var characters []*database.Character

	for _, char := range self.characters {
		if char.IsOnline() {
			characters = append(characters, char)
		}
	}

	return characters
}

// CreateCharacter creates a new Character object in the database and adds it to the model.
// A pointer to the new character object is returned.
func (self *globalModel) CreateCharacter(name string, parentUser *database.User, startingRoom *database.Room) *database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	character := database.NewCharacter(name, parentUser.GetId(), startingRoom.GetId())
	self.characters[character.GetId()] = character

	return character
}

func (self *globalModel) CreateNpc(name string, room *database.Room) *database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	npc := database.NewNpc(name, room.GetId())
	self.characters[npc.GetId()] = npc

	return npc
}

// DeleteCharacter removes the character associated with the given id from the
// model and from the database
func (self *globalModel) DeleteCharacter(id bson.ObjectId) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	delete(self.characters, id)

	utils.HandleError(database.DeleteCharacter(id))
}

// CreateRoom creates a new Room object in the database and adds it to the model.
// A pointer to the new room object is returned.
func (self *globalModel) CreateRoom(zone *database.Zone) *database.Room {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	room := database.NewRoom(zone.GetId())
	self.rooms[room.GetId()] = room

	return room
}

func (self *globalModel) CreateZone(name string) *database.Zone {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	zone := database.NewZone(name)
	self.zones[zone.GetId()] = zone

	return zone
}

// GetRoom returns the room object associated with the given id
func (self *globalModel) GetRoom(id bson.ObjectId) *database.Room {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.rooms[id]
}

// GetRooms returns a list of all of the rooms in the entire model
func (self *globalModel) GetRooms() []*database.Room {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var rooms []*database.Room

	for _, room := range self.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

func (self *globalModel) GetRoomsInZone(zone *database.Zone) []*database.Room {
	allRooms := self.GetRooms()

	var rooms []*database.Room

	for _, room := range allRooms {
		if room.GetZoneId() == zone.GetId() {
			rooms = append(rooms, room)
		}
	}

	return rooms
}

// GetRoomByLocation searches for the room associated with the given coordinate
// in the given zone. Returns a nil room object if it was not found.
func (self *globalModel) GetRoomByLocation(coordinate database.Coordinate, zone *database.Zone) *database.Room {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	for _, room := range self.rooms {
		if room.GetLocation() == coordinate && room.GetZoneId() == zone.GetId() {
			return room
		}
	}

	return nil
}

// GetZone returns the zone object associated with the given id
func (self *globalModel) GetZone(zoneId bson.ObjectId) *database.Zone {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.zones[zoneId]
}

// GetZones returns all of the zones in the model
func (self *globalModel) GetZones() []*database.Zone {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var zones []*database.Zone

	for _, zone := range self.zones {
		zones = append(zones, zone)
	}

	return zones
}

// GetZoneByName name searches for a zone with the given name, returns a zone
// object and whether or not it was found
func (self *globalModel) GetZoneByName(name string) *database.Zone {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	for _, zone := range self.zones {
		if zone.GetName() == name {
			return zone
		}
	}

	return nil
}

func (self *globalModel) deleteRoom(id bson.ObjectId) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	delete(self.rooms, id)

	utils.HandleError(database.DeleteRoom(id))
}

// GetUser returns the User object associated with the given id
func (self *globalModel) GetUser(id bson.ObjectId) *database.User {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.users[id]
}

// GetUsers returns all of the User objects in the model
func (self *globalModel) GetUsers() []*database.User {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var users []*database.User

	for _, user := range self.users {
		users = append(users, user)
	}

	return users
}

// GetUserByName searches for the User object with the given name. Returns a
// nil User if one was not found.
func (self *globalModel) GetUserByName(username string) *database.User {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	for _, user := range self.users {
		if user.GetName() == username {
			return user
		}
	}

	return nil
}

// Removes the User assocaited with the given id from the model. Removes it from the database as well
func (self *globalModel) DeleteUser(id bson.ObjectId) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for _, character := range self.characters {
		if character.GetUserId() == id {
			delete(self.characters, character.GetId())
			utils.HandleError(database.DeleteCharacter(character.GetId()))
		}
	}

	delete(self.users, id)

	utils.HandleError(database.DeleteUser(id))
}

func (self *globalModel) CreateItem(name string) *database.Item {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	item := database.NewItem(name)
	self.items[item.GetId()] = item

	return item
}

// GetItem returns the Item object associated the given id
func (self *globalModel) GetItem(id bson.ObjectId) *database.Item {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.items[id]
}

// GetItems returns the Items object associated the given ids
func (self *globalModel) GetItems(itemIds []bson.ObjectId) []*database.Item {
	var items []*database.Item

	for _, itemId := range itemIds {
		items = append(items, self.GetItem(itemId))
	}

	return items
}

func (self *globalModel) ItemsIn(room *database.Room) []*database.Item {
	return self.GetItems(room.GetItemIds())
}

// DeleteItem removes the item associated with the given id from the
// model and from the database
func (self *globalModel) DeleteItem(id bson.ObjectId) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	delete(self.items, id)

	utils.HandleError(database.DeleteItem(id))
}

// M is the global model object. All functions are thread-safe and all changes
// made to the model are automatically saved to the database.
var M globalModel

var eventQueueChannel chan Event

/**
 * Initializes the global model object and starts up the main event loop
 */
func Init(session *mgo.Session) error {
	database.Init(session)

	M = globalModel{}

	M.users = map[bson.ObjectId]*database.User{}
	M.characters = map[bson.ObjectId]*database.Character{}
	M.rooms = map[bson.ObjectId]*database.Room{}
	M.zones = map[bson.ObjectId]*database.Zone{}
	M.items = map[bson.ObjectId]*database.Item{}

	users, err := database.GetAllUsers()
	utils.HandleError(err)

	for _, user := range users {
		M.users[user.GetId()] = user
	}

	characters, err := database.GetAllCharacters()
	utils.HandleError(err)

	for _, character := range characters {
		M.characters[character.GetId()] = character
	}

	rooms, err := database.GetAllRooms()
	utils.HandleError(err)

	for _, room := range rooms {
		M.rooms[room.GetId()] = room
	}

	zones, err := database.GetAllZones()
	utils.HandleError(err)

	for _, zone := range zones {
		M.zones[zone.GetId()] = zone
	}

	items, err := database.GetAllItems()
	utils.HandleError(err)

	for _, item := range items {
		M.items[item.GetId()] = item
	}

	// Start the event loop
	eventQueueChannel = make(chan Event, 100)
	go eventLoop()

	return err
}

func MoveCharacterToLocation(character *database.Character, zone *database.Zone, location database.Coordinate) (*database.Room, error) {
	newRoom := M.GetRoomByLocation(location, zone)

	if newRoom == nil {
		return newRoom, errors.New("Invalid location")
	}

	oldRoom := M.GetRoom(character.GetRoomId())

	character.SetRoom(newRoom.GetId())

	queueEvent(EnterEvent{Character: character, RoomId: newRoom.GetId()})
	queueEvent(LeaveEvent{Character: character, RoomId: oldRoom.GetId()})

	return newRoom, nil
}

func MoveCharacterToRoom(character *database.Character, newRoom *database.Room) {
	oldRoomId := character.GetRoomId()
	character.SetRoom(newRoom.GetId())

	queueEvent(EnterEvent{Character: character, RoomId: newRoom.GetId()})
	queueEvent(LeaveEvent{Character: character, RoomId: oldRoomId})
}

func MoveCharacter(character *database.Character, direction database.ExitDirection) (*database.Room, error) {
	room := M.GetRoom(character.GetRoomId())

	if room == nil {
		return room, errors.New("Character doesn't appear to be in any room")
	}

	if !room.HasExit(direction) {
		return room, errors.New("Attempted to move through an exit that the room does not contain")
	}

	newLocation := room.NextLocation(direction)
	newRoom := M.GetRoomByLocation(newLocation, M.GetZone(room.GetZoneId()))

	if newRoom == nil {
		zone := M.GetZone(room.GetZoneId())
		fmt.Printf("No room found at location %v %v, creating a new one (%s)\n", zone.GetName(), newLocation, character.PrettyName())

		room = M.CreateRoom(M.GetZone(room.GetZoneId()))

		switch direction {
		case database.DirectionNorth:
			room.SetExitEnabled(database.DirectionSouth, true)
		case database.DirectionNorthEast:
			room.SetExitEnabled(database.DirectionSouthWest, true)
		case database.DirectionEast:
			room.SetExitEnabled(database.DirectionWest, true)
		case database.DirectionSouthEast:
			room.SetExitEnabled(database.DirectionNorthWest, true)
		case database.DirectionSouth:
			room.SetExitEnabled(database.DirectionNorth, true)
		case database.DirectionSouthWest:
			room.SetExitEnabled(database.DirectionNorthEast, true)
		case database.DirectionWest:
			room.SetExitEnabled(database.DirectionEast, true)
		case database.DirectionNorthWest:
			room.SetExitEnabled(database.DirectionSouthEast, true)
		case database.DirectionUp:
			room.SetExitEnabled(database.DirectionDown, true)
		case database.DirectionDown:
			room.SetExitEnabled(database.DirectionUp, true)
		default:
			panic("Unexpected code path")
		}

		room.SetLocation(newLocation)
	} else {
		room = newRoom
	}

	return MoveCharacterToLocation(character, M.GetZone(room.GetZoneId()), room.GetLocation())
}

func DeleteRoom(room *database.Room) {
	M.deleteRoom(room.GetId())

	// Disconnect all exits leading to this room
	loc := room.GetLocation()

	updateRoom := func(dir database.ExitDirection) {
		next := loc.Next(dir)
		room := M.GetRoomByLocation(next, M.GetZone(room.GetZoneId()))

		if room != nil {
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

func BroadcastMessage(from *database.Character, message string) {
	queueEvent(BroadcastEvent{from, message})
}

func Tell(from *database.Character, to *database.Character, message string) {
	queueEvent(TellEvent{from, to, message})
}

func Say(from *database.Character, message string) {
	queueEvent(SayEvent{from, message})
}

func Emote(from *database.Character, message string) {
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
		if room.GetZoneId() == zoneId {
			top = room.GetLocation().Y
			bottom = room.GetLocation().Y
			left = room.GetLocation().X
			right = room.GetLocation().X
			high = room.GetLocation().Z
			low = room.GetLocation().Z
			break
		}
	}

	for _, room := range rooms {
		if room.GetZoneId() != zoneId {
			continue
		}

		if room.GetLocation().Z < high {
			high = room.GetLocation().Z
		}

		if room.GetLocation().Z > low {
			low = room.GetLocation().Z
		}

		if room.GetLocation().Y < top {
			top = room.GetLocation().Y
		}

		if room.GetLocation().Y > bottom {
			bottom = room.GetLocation().Y
		}

		if room.GetLocation().X < left {
			left = room.GetLocation().X
		}

		if room.GetLocation().X > right {
			right = room.GetLocation().X
		}
	}

	return database.Coordinate{X: left, Y: top, Z: high},
		database.Coordinate{X: right, Y: bottom, Z: low}
}

func MoveRoomsToZone(fromZoneId bson.ObjectId, toZoneId bson.ObjectId) {
	for _, room := range M.GetRooms() {
		if room.GetZoneId() == fromZoneId {
			room.SetZoneId(toZoneId)
		}
	}
}

// vim: nocindent
