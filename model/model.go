package model

import (
	"errors"
	"fmt"
	"kmud/database"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
	"sync"
)

type UpdateType int

const (
	NewNpcUpdate          UpdateType = iota
	CharacterDeleteUpdate UpdateType = iota
)

type globalModel struct {
	users map[bson.ObjectId]*database.User
	chars map[bson.ObjectId]*database.Character
	rooms map[bson.ObjectId]*database.Room
	zones map[bson.ObjectId]*database.Zone
	items map[bson.ObjectId]*database.Item

	mutex sync.RWMutex
}

// M is the global model object. All functions are thread-safe and all changes
// made to the model are automatically saved to the database.
var M globalModel

var eventListeners map[UpdateType][]chan interface{}

// CreateUser creates a new User object in the database and adds it to the model.
// A pointer to the new User object is returned.
func (self *globalModel) CreateUser(name string, password string) *database.User {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	user := database.NewUser(name, password)
	self.users[user.GetId()] = user

	return user
}

// GetOrCreateUser attempts to retrieve the existing user from the model by the given name.
// if none exists, then a new one is created with the given credentials.
func (self *globalModel) GetOrCreateUser(name string, password string) *database.User {
	user := self.GetUserByName(name)

	if user == nil {
		user = self.CreateUser(name, password)
	}

	return user
}

// GetUsers returns all of the User objects in the model
func (self *globalModel) GetUsers() database.Users {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var users database.Users

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
		if utils.Compare(user.GetName(), username) {
			return user
		}
	}

	return nil
}

func (self *globalModel) DeleteUserId(userId bson.ObjectId) {
	self.DeleteUser(self.GetUser(userId))
}

// Removes the given User from the model. Removes it from the database as well.
func (self *globalModel) DeleteUser(user *database.User) {
	for _, character := range M.GetUserCharacters(user) {
		self.DeleteCharacter(character)
	}

	self.mutex.Lock()
	defer self.mutex.Unlock()

	delete(self.users, user.GetId())
	utils.HandleError(database.DeleteObject(user))
}

// GetCharacter returns the Character object associated the given Id
func (self *globalModel) GetCharacter(id bson.ObjectId) *database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.chars[id]
}

// GetCharacaterByName searches for a character with the given name. Returns a
// character object, or nil if it wasn't found.
func (self *globalModel) GetCharacterByName(name string) *database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	name = utils.Simplify(name)

	for _, character := range self.chars {
		if character.GetName() == name {
			return character
		}
	}

	return nil
}

func (self *globalModel) GetAllNpcs() []*database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	npcs := []*database.Character{}

	for _, character := range self.chars {
		if character.IsNpc() {
			npcs = append(npcs, character)
		}
	}

	return npcs
}

// GetUserCharacters returns all of the Character objects associated with the
// given user id
func (self *globalModel) GetUserCharacters(user *database.User) []*database.Character {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var characters []*database.Character

	for _, character := range self.chars {
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

	for _, char := range self.chars {
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

	for _, char := range self.chars {
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

	for _, char := range self.chars {
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

	for _, char := range self.chars {
		if char.IsPlayer() && char.IsOnline() {
			characters = append(characters, char)
		}
	}

	return characters
}

// CreatePlayer creates a new player-controlled Character object in the
// database and adds it to the model.  A pointer to the new character object is
// returned.
func (self *globalModel) CreatePlayer(name string, parentUser *database.User, startingRoom *database.Room) *database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	character := database.NewCharacter(name, parentUser.GetId(), startingRoom.GetId())
	self.chars[character.GetId()] = character

	return character
}

// GetOrCreatePlayer attempts to retrieve the existing user from the model by the given name.
// if none exists, then a new one is created. If the name matches an NPC (rather than a player)
// then nil will be returned.
func (self *globalModel) GetOrCreatePlayer(name string, parentUser *database.User, startingRoom *database.Room) *database.Character {
	player := self.GetCharacterByName(name)

	if player == nil {
		player = self.CreatePlayer(name, parentUser, startingRoom)
	} else if player.IsNpc() {
		return nil
	}

	return player
}

// CreateNpc is a convenience function for creating a new character object that
// is an NPC (as opposed to an actual player-controlled character)
func (self *globalModel) CreateNpc(name string, room *database.Room) *database.Character {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	npc := database.NewNpc(name, room.GetId())
	self.chars[npc.GetId()] = npc

	emit(NewNpcUpdate, npc)

	return npc
}

func (self *globalModel) DeleteCharacterId(id bson.ObjectId) {
	self.DeleteCharacter(self.GetCharacter(id))
}

// DeleteCharacter removes the character (either NPC or player-controlled)
// associated with the given id from the model and from the database
func (self *globalModel) DeleteCharacter(character *database.Character) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	emit(CharacterDeleteUpdate, character)

	delete(self.chars, character.GetId())
	utils.HandleError(database.DeleteObject(character))
}

// CreateRoom creates a new Room object in the database and adds it to the model.
// A pointer to the new Room object is returned.
func (self *globalModel) CreateRoom(zone *database.Zone, location database.Coordinate) (*database.Room, error) {
	existingRoom := M.GetRoomByLocation(location, zone)
	if existingRoom != nil {
		return nil, errors.New("A room already exists at that location")
	}

	self.mutex.Lock()
	defer self.mutex.Unlock()

	room := database.NewRoom(zone.GetId(), location)
	self.rooms[room.GetId()] = room

	return room, nil
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

// GetRoomsInZone returns a slice containing all of the rooms that belong to
// the given zone
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
func (self *globalModel) GetZones() database.Zones {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	var zones []*database.Zone

	for _, zone := range self.zones {
		zones = append(zones, zone)
	}

	return zones
}

// CreateZone creates a new Zone object in the database and adds it to the model.
// A pointer to the new Zone object is returned.
func (self *globalModel) CreateZone(name string) (*database.Zone, error) {
	if M.GetZoneByName(name) != nil {
		return nil, errors.New("A zone with that name already exists")
	}

	self.mutex.Lock()
	defer self.mutex.Unlock()

	zone := database.NewZone(name)
	self.zones[zone.GetId()] = zone

	return zone, nil
}

// Removes the given Zone from the model. Removes it from the database as well.
func (self *globalModel) DeleteZone(zone *database.Zone) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	delete(self.zones, zone.GetId())
	utils.HandleError(database.DeleteObject(zone))
}

// GetZoneByName name searches for a zone with the given name, returns a zone
// object and whether or not it was found
func (self *globalModel) GetZoneByName(name string) *database.Zone {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	for _, zone := range self.zones {
		if utils.Compare(zone.GetName(), name) {
			return zone
		}
	}

	return nil
}

// DeleteRoom removes the given room object from the model and the database. It
// also disables all exits in neighboring rooms that lead to the given room.
func (self *globalModel) DeleteRoom(room *database.Room) {
	self.mutex.Lock()
	delete(self.rooms, room.GetId())

	utils.HandleError(database.DeleteObject(room))
	self.mutex.Unlock()

	// Disconnect all exits leading to this room
	loc := room.GetLocation()

	updateRoom := func(dir database.Direction) {
		next := loc.Next(dir)
		room := M.GetRoomByLocation(next, M.GetZone(room.GetZoneId()))

		if room != nil {
			room.SetExitEnabled(dir.Opposite(), false)
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

// GetUser returns the User object associated with the given id
func (self *globalModel) GetUser(id bson.ObjectId) *database.User {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.users[id]
}

// CreateItem creates an item object in the database with the given name and
// adds it to the model. It's up to the caller to ensure that the item actually
// gets put somewhere meaningful.
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

// ItemsIn returns a slice containing all of the items in the given room
func (self *globalModel) ItemsIn(room *database.Room) []*database.Item {
	return self.GetItems(room.GetItemIds())
}

func (self *globalModel) DeleteItemId(itemId bson.ObjectId) {
	self.DeleteItem(self.GetItem(itemId))
}

// DeleteItem removes the item associated with the given id from the
// model and from the database
func (self *globalModel) DeleteItem(item *database.Item) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	delete(self.items, item.GetId())

	utils.HandleError(database.DeleteObject(item))
}

// Initializes the global model object and starts up the main event loop
func Init(session database.Session) error {
	database.Init(session)

	M = globalModel{}

	eventListeners = map[UpdateType][]chan interface{}{}

	M.users = map[bson.ObjectId]*database.User{}
	M.chars = map[bson.ObjectId]*database.Character{}
	M.rooms = map[bson.ObjectId]*database.Room{}
	M.zones = map[bson.ObjectId]*database.Zone{}
	M.items = map[bson.ObjectId]*database.Item{}

	users := []*database.User{}
	err := database.RetrieveObjects(database.UserType, &users)
	utils.HandleError(err)

	for _, user := range users {
		M.users[user.GetId()] = user
	}

	characters := []*database.Character{}
	err = database.RetrieveObjects(database.CharType, &characters)
	utils.HandleError(err)

	for _, character := range characters {
		M.chars[character.GetId()] = character
	}

	rooms := []*database.Room{}
	err = database.RetrieveObjects(database.RoomType, &rooms)
	utils.HandleError(err)

	for _, room := range rooms {
		M.rooms[room.GetId()] = room
	}

	zones := []*database.Zone{}
	err = database.RetrieveObjects(database.ZoneType, &zones)
	utils.HandleError(err)

	for _, zone := range zones {
		M.zones[zone.GetId()] = zone
	}

	items := []*database.Item{}
	err = database.RetrieveObjects(database.ItemType, &items)
	utils.HandleError(err)

	for _, item := range items {
		M.items[item.GetId()] = item
	}

	// Start the event loop
	eventQueueChannel = make(chan Event, 100)
	go eventLoop()

	fights = map[*database.Character]*database.Character{}
	go combatLoop()

	return err
}

// MoveCharacter attempts to move the character to the given coordinates
// specific by location. Returns an error if there is no room to move to.
func MoveCharacterToLocation(character *database.Character, zone *database.Zone, location database.Coordinate) (*database.Room, error) {
	newRoom := M.GetRoomByLocation(location, zone)

	if newRoom == nil {
		return nil, errors.New("Invalid location")
	}

	MoveCharacterToRoom(character, newRoom)
	return newRoom, nil
}

// MoveCharacterTo room moves the character to the given room
func MoveCharacterToRoom(character *database.Character, newRoom *database.Room) {
	oldRoomId := character.GetRoomId()
	character.SetRoomId(newRoom.GetId())

	oldRoom := M.GetRoom(oldRoomId)

	queueEvent(EnterEvent{Character: character, Room: newRoom, SourceRoom: oldRoom})
	queueEvent(LeaveEvent{Character: character, Room: oldRoom, DestRoom: newRoom})
}

// MoveCharacter moves the given character in the given direction. If there is
// no exit in that direction, and error is returned. If there is an exit, but no
// room connected to it, then a room is automatically created for the character
// to move in to.
func MoveCharacter(character *database.Character, direction database.Direction) (*database.Room, error) {
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
		fmt.Printf("No room found at location %v %v, creating a new one (%s)\n", zone.GetName(), newLocation, character.GetName())

		var err error
		room, err = M.CreateRoom(M.GetZone(room.GetZoneId()), newLocation)

		if err != nil {
			return nil, err
		}

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
	} else {
		room = newRoom
	}

	return MoveCharacterToLocation(character, M.GetZone(room.GetZoneId()), room.GetLocation())
}

// BroadcastMessage sends a message to all users that are logged in
func BroadcastMessage(from *database.Character, message string) {
	queueEvent(BroadcastEvent{from, message})
}

// Tell sends a message to the specified character
func Tell(from *database.Character, to *database.Character, message string) {
	queueEvent(TellEvent{from, to, message})
}

// Say sends a message to all characters in the given character's room
func Say(from *database.Character, message string) {
	queueEvent(SayEvent{from, message})
}

// Emote sends an emote message to all characters in the given character's room
func Emote(from *database.Character, message string) {
	queueEvent(EmoteEvent{from, message})
}

// ZoneCorners returns cordinates that indiate the highest and lowest points of
// the map in 3 dimensions
func ZoneCorners(zone *database.Zone) (database.Coordinate, database.Coordinate) {
	var top int
	var bottom int
	var left int
	var right int
	var high int
	var low int

	rooms := M.GetRoomsInZone(zone)

	for _, room := range M.rooms {
		top = room.GetLocation().Y
		bottom = room.GetLocation().Y
		left = room.GetLocation().X
		right = room.GetLocation().X
		high = room.GetLocation().Z
		low = room.GetLocation().Z
		break
	}

	for _, room := range rooms {
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

// Returns the exit direction of the given room if it is adjacent, otherwise DirectionNone
func DirectionBetween(from, to *database.Room) database.Direction {
	zone := M.GetZone(from.GetZoneId())
	for _, exit := range from.GetExits() {
		nextLocation := from.NextLocation(exit)
		nextRoom := M.GetRoomByLocation(nextLocation, zone)

		if nextRoom == to {
			return exit
		}
	}

	return database.DirectionNone
}

func emit(updateType UpdateType, data interface{}) {
	for _, channel := range eventListeners[updateType] {
		channel <- data
	}
}

func Watch(updateType UpdateType) chan interface{} {
	channel := make(chan interface{})
	eventListeners[updateType] = append(eventListeners[updateType], channel)
	return channel
}

func Unwatch(updateType UpdateType, channel chan interface{}) bool {
	for i, myChannel := range eventListeners[updateType] {
		if myChannel == channel {
			// TODO: Potential memory leak. See http://code.google.com/p/go-wiki/wiki/SliceTricks
			eventListeners[updateType] = append(eventListeners[updateType][:i], eventListeners[updateType][i+1:]...)
			return true
		}
	}

	return false
}

// vim: nocindent
