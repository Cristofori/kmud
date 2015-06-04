package model

import (
	"errors"
	"fmt"
	"kmud/database"
	"kmud/utils"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

var _users map[bson.ObjectId]*database.User
var _chars map[bson.ObjectId]*database.Character
var _zones map[bson.ObjectId]*database.Zone
var _areas map[bson.ObjectId]*database.Area
var _rooms map[bson.ObjectId]*database.Room
var _items map[bson.ObjectId]*database.Item

var mutex sync.RWMutex

// CreateUser creates a new User object in the database and adds it to the model.
// A pointer to the new User object is returned.
func CreateUser(name string, password string) *database.User {
	mutex.Lock()
	defer mutex.Unlock()

	user := database.NewUser(name, password)
	_users[user.GetId()] = user

	return user
}

// GetOrCreateUser attempts to retrieve the existing user from the model by the given name.
// if none exists, then a new one is created with the given credentials.
func GetOrCreateUser(name string, password string) *database.User {
	user := GetUserByName(name)

	if user == nil {
		user = CreateUser(name, password)
	}

	return user
}

// GetUsers returns all of the User objects in the model
func GetUsers() database.Users {
	mutex.RLock()
	defer mutex.RUnlock()

	var users database.Users

	for _, user := range _users {
		users = append(users, user)
	}

	return users
}

// GetUserByName searches for the User object with the given name. Returns a
// nil User if one was not found.
func GetUserByName(username string) *database.User {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, user := range _users {
		if utils.Compare(user.GetName(), username) {
			return user
		}
	}

	return nil
}

func DeleteUserId(userId bson.ObjectId) {
	DeleteUser(GetUser(userId))
}

// Removes the given User from the model. Removes it from the database as well.
func DeleteUser(user *database.User) {
	for _, character := range GetUserCharacters(user) {
		DeleteCharacter(character)
	}

	mutex.Lock()
	defer mutex.Unlock()

	delete(_users, user.GetId())
	utils.HandleError(database.DeleteObject(user))
}

// GetCharacter returns the Character object associated the given Id
func GetCharacter(id bson.ObjectId) *database.Character {
	mutex.RLock()
	defer mutex.RUnlock()

	return _chars[id]
}

// GetCharacaterByName searches for a character with the given name. Returns a
// character object, or nil if it wasn't found.
func GetCharacterByName(name string) *database.Character {
	mutex.RLock()
	defer mutex.RUnlock()

	name = utils.Simplify(name)

	for _, character := range _chars {
		if character.GetName() == name {
			return character
		}
	}

	return nil
}

func GetAllNpcs() []*database.Character {
	mutex.RLock()
	defer mutex.RUnlock()

	npcs := []*database.Character{}

	for _, character := range _chars {
		if character.IsNpc() {
			npcs = append(npcs, character)
		}
	}

	return npcs
}

func GetAllNpcTemplates() []*database.Character {
	mutex.RLock()
	defer mutex.RUnlock()

	templates := []*database.Character{}

	for _, character := range _chars {
		if character.IsNpcTemplate() {
			templates = append(templates, character)
		}
	}

	return templates
}

// GetUserCharacters returns all of the Character objects associated with the
// given user id
func GetUserCharacters(user *database.User) []*database.Character {
	mutex.RLock()
	defer mutex.RUnlock()

	var characters []*database.Character

	for _, character := range _chars {
		if character.GetUserId() == user.GetId() {
			characters = append(characters, character)
		}
	}

	return characters
}

// CharactersIn returns a list of characters that are in the given room (NPC or
// player), excluding the character passed in as the "except" parameter.
// Returns all character type objects, including players, NPCs and MOBs
func CharactersIn(room *database.Room) []*database.Character {
	mutex.RLock()
	defer mutex.RUnlock()

	var charList []*database.Character

	for _, char := range _chars {
		if char.GetRoomId() == room.GetId() && char.IsOnline() {
			charList = append(charList, char)
		}
	}

	return charList
}

// PlayersIn returns all of the player characters that are in the given room
func PlayersIn(room *database.Room, except *database.Character) []*database.Character {
	mutex.RLock()
	defer mutex.RUnlock()

	var playerList []*database.Character

	for _, char := range _chars {
		if char.GetRoomId() == room.GetId() && char.IsPlayer() && char.IsOnline() && char != except {
			playerList = append(playerList, char)
		}
	}

	return playerList
}

// NpcsIn returns all of the NPC characters that are in the given room
func NpcsIn(room *database.Room) []*database.Character {
	mutex.RLock()
	defer mutex.RUnlock()

	var npcList []*database.Character

	for _, char := range _chars {
		if char.GetRoomId() == room.GetId() && char.IsNpc() {
			npcList = append(npcList, char)
		}
	}

	return npcList
}

// GetOnlineCharacters returns a list of all of the characters who are online
func GetOnlineCharacters() []*database.Character {
	mutex.RLock()
	defer mutex.RUnlock()

	var characters []*database.Character

	for _, char := range _chars {
		if char.IsPlayer() && char.IsOnline() {
			characters = append(characters, char)
		}
	}

	return characters
}

// CreatePlayer creates a new player-controlled Character object in the
// database and adds it to the model.  A pointer to the new character object is
// returned.
func CreatePlayer(name string, parentUser *database.User, startingRoom *database.Room) *database.Character {
	mutex.Lock()
	defer mutex.Unlock()

	character := database.NewCharacter(name, parentUser.GetId(), startingRoom.GetId())
	_chars[character.GetId()] = character

	return character
}

// GetOrCreatePlayer attempts to retrieve the existing user from the model by the given name.
// if none exists, then a new one is created. If the name matches an NPC (rather than a player)
// then nil will be returned.
func GetOrCreatePlayer(name string, parentUser *database.User, startingRoom *database.Room) *database.Character {
	player := GetCharacterByName(name)

	if player == nil {
		player = CreatePlayer(name, parentUser, startingRoom)
	} else if player.IsNpc() {
		return nil
	}

	return player
}

// CreateNpc is a convenience function for creating a new character object that
// is an NPC (as opposed to an actual player-controlled character)
func CreateNpc(name string, room *database.Room) *database.Character {
	mutex.Lock()
	defer mutex.Unlock()

	npc := database.NewNpc(name, room.GetId())
	_chars[npc.GetId()] = npc

	return npc
}

func CreateNpcTemplate(name string) *database.Character {
	mutex.Lock()
	defer mutex.Unlock()

	template := database.NewNpcTemplate(name)
	_chars[template.GetId()] = template

	return template
}

func DeleteCharacterId(id bson.ObjectId) {
	DeleteCharacter(GetCharacter(id))
}

// DeleteCharacter removes the character (either NPC or player-controlled)
// associated with the given id from the model and from the database
func DeleteCharacter(character *database.Character) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(_chars, character.GetId())
	utils.HandleError(database.DeleteObject(character))
}

// CreateRoom creates a new Room object in the database and adds it to the model.
// A pointer to the new Room object is returned.
func CreateRoom(zone *database.Zone, location database.Coordinate) (*database.Room, error) {
	existingRoom := GetRoomByLocation(location, zone)
	if existingRoom != nil {
		return nil, errors.New("A room already exists at that location")
	}

	mutex.Lock()
	defer mutex.Unlock()

	room := database.NewRoom(zone.GetId(), location)
	_rooms[room.GetId()] = room

	return room, nil
}

// GetRoom returns the room object associated with the given id
func GetRoom(id bson.ObjectId) *database.Room {
	mutex.RLock()
	defer mutex.RUnlock()

	return _rooms[id]
}

// GetRooms returns a list of all of the rooms in the entire model
func GetRooms() []*database.Room {
	mutex.RLock()
	defer mutex.RUnlock()

	var rooms []*database.Room

	for _, room := range _rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

// GetRoomsInZone returns a slice containing all of the rooms that belong to
// the given zone
func GetRoomsInZone(zone *database.Zone) []*database.Room {
	allRooms := GetRooms()

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
func GetRoomByLocation(coordinate database.Coordinate, zone *database.Zone) *database.Room {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, room := range _rooms {
		if room.GetLocation() == coordinate && room.GetZoneId() == zone.GetId() {
			return room
		}
	}

	return nil
}

// GetZone returns the zone object associated with the given id
func GetZone(zoneId bson.ObjectId) *database.Zone {
	mutex.RLock()
	defer mutex.RUnlock()

	return _zones[zoneId]
}

// GetZones returns all of the zones in the model
func GetZones() database.Zones {
	mutex.RLock()
	defer mutex.RUnlock()

	var zones database.Zones

	for _, zone := range _zones {
		zones = append(zones, zone)
	}

	return zones
}

// CreateZone creates a new Zone object in the database and adds it to the model.
// A pointer to the new Zone object is returned.
func CreateZone(name string) (*database.Zone, error) {
	if GetZoneByName(name) != nil {
		return nil, errors.New("A zone with that name already exists")
	}

	mutex.Lock()
	defer mutex.Unlock()

	zone := database.NewZone(name)
	_zones[zone.GetId()] = zone

	return zone, nil
}

// Removes the given Zone from the model. Removes it from the database as well.
func DeleteZone(zone *database.Zone) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(_zones, zone.GetId())
	utils.HandleError(database.DeleteObject(zone))
}

// GetZoneByName name searches for a zone with the given name, returns a zone
// object and whether or not it was found
func GetZoneByName(name string) *database.Zone {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, zone := range _zones {
		if utils.Compare(zone.GetName(), name) {
			return zone
		}
	}

	return nil
}

func GetAreas(zone *database.Zone) database.Areas {
	mutex.RLock()
	defer mutex.RUnlock()

	var areas database.Areas

	for _, area := range _areas {
		if area.GetZoneId() == zone.GetId() {
			areas = append(areas, area)
		}
	}

	return areas
}

func GetArea(areaId bson.ObjectId) *database.Area {
	mutex.RLock()
	defer mutex.RUnlock()

	return _areas[areaId]
}

func CreateArea(name string, zone *database.Zone) (*database.Area, error) {
	if GetAreaByName(name) != nil {
		return nil, errors.New("An area with that name already exists")
	}

	mutex.Lock()
	defer mutex.Unlock()

	area := database.NewArea(name, zone.GetId())
	_areas[area.GetId()] = area

	return area, nil
}

func GetAreaByName(name string) *database.Area {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, area := range _areas {
		if utils.Compare(area.GetName(), name) {
			return area
		}
	}

	return nil
}

func DeleteArea(area *database.Area) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(_areas, area.GetId())
	utils.HandleError(database.DeleteObject(area))
}

// DeleteRoom removes the given room object from the model and the database. It
// also disables all exits in neighboring rooms that lead to the given room.
func DeleteRoom(room *database.Room) {
	mutex.Lock()
	delete(_rooms, room.GetId())

	utils.HandleError(database.DeleteObject(room))
	mutex.Unlock()

	// Disconnect all exits leading to this room
	loc := room.GetLocation()

	updateRoom := func(dir database.Direction) {
		next := loc.Next(dir)
		room := GetRoomByLocation(next, GetZone(room.GetZoneId()))

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
func GetUser(id bson.ObjectId) *database.User {
	mutex.RLock()
	defer mutex.RUnlock()

	return _users[id]
}

// CreateItem creates an item object in the database with the given name and
// adds it to the model. It's up to the caller to ensure that the item actually
// gets put somewhere meaningful.
func CreateItem(name string) *database.Item {
	mutex.Lock()
	defer mutex.Unlock()

	item := database.NewItem(name)
	_items[item.GetId()] = item

	return item
}

// GetItem returns the Item object associated the given id
func GetItem(id bson.ObjectId) *database.Item {
	mutex.RLock()
	defer mutex.RUnlock()

	return _items[id]
}

// GetItems returns the Items object associated the given ids
func GetItems(itemIds []bson.ObjectId) []*database.Item {
	var items []*database.Item

	for _, itemId := range itemIds {
		items = append(items, GetItem(itemId))
	}

	return items
}

// ItemsIn returns a slice containing all of the items in the given room
func ItemsIn(room *database.Room) []*database.Item {
	return GetItems(room.GetItemIds())
}

func DeleteItemId(itemId bson.ObjectId) {
	DeleteItem(GetItem(itemId))
}

// DeleteItem removes the item associated with the given id from the
// model and from the database
func DeleteItem(item *database.Item) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(_items, item.GetId())

	utils.HandleError(database.DeleteObject(item))
}

// Initializes the global model object and starts up the main event loop
func Init(session database.Session) error {
	database.Init(session)

	_users = map[bson.ObjectId]*database.User{}
	_chars = map[bson.ObjectId]*database.Character{}
	_zones = map[bson.ObjectId]*database.Zone{}
	_areas = map[bson.ObjectId]*database.Area{}
	_rooms = map[bson.ObjectId]*database.Room{}
	_items = map[bson.ObjectId]*database.Item{}

	users := []*database.User{}
	err := database.RetrieveObjects(database.UserType, &users)
	utils.HandleError(err)

	for _, user := range users {
		_users[user.GetId()] = user
	}

	characters := []*database.Character{}
	err = database.RetrieveObjects(database.CharType, &characters)
	utils.HandleError(err)

	for _, character := range characters {
		_chars[character.GetId()] = character
	}

	zones := []*database.Zone{}
	err = database.RetrieveObjects(database.ZoneType, &zones)
	utils.HandleError(err)

	for _, zone := range zones {
		_zones[zone.GetId()] = zone
	}

	areas := []*database.Area{}
	err = database.RetrieveObjects(database.AreaType, &areas)
	utils.HandleError(err)

	for _, area := range areas {
		_areas[area.GetId()] = area
	}

	rooms := []*database.Room{}
	err = database.RetrieveObjects(database.RoomType, &rooms)
	utils.HandleError(err)

	for _, room := range rooms {
		_rooms[room.GetId()] = room
	}

	items := []*database.Item{}
	err = database.RetrieveObjects(database.ItemType, &items)
	utils.HandleError(err)

	for _, item := range items {
		_items[item.GetId()] = item
	}

	// Start the event loop
	go eventLoop()

	fights = map[*database.Character]*database.Character{}
	go combatLoop()

	return err
}

// MoveCharacter attempts to move the character to the given coordinates
// specific by location. Returns an error if there is no room to move to.
func MoveCharacterToLocation(character *database.Character, zone *database.Zone, location database.Coordinate) (*database.Room, error) {
	newRoom := GetRoomByLocation(location, zone)

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

	oldRoom := GetRoom(oldRoomId)

	queueEvent(EnterEvent{Character: character, Room: newRoom, SourceRoom: oldRoom})
	queueEvent(LeaveEvent{Character: character, Room: oldRoom, DestRoom: newRoom})
}

// MoveCharacter moves the given character in the given direction. If there is
// no exit in that direction, and error is returned. If there is an exit, but no
// room connected to it, then a room is automatically created for the character
// to move in to.
func MoveCharacter(character *database.Character, direction database.Direction) (*database.Room, error) {
	room := GetRoom(character.GetRoomId())

	if room == nil {
		return room, errors.New("Character doesn't appear to be in any room")
	}

	if !room.HasExit(direction) {
		return room, errors.New("Attempted to move through an exit that the room does not contain")
	}

	newLocation := room.NextLocation(direction)
	newRoom := GetRoomByLocation(newLocation, GetZone(room.GetZoneId()))

	if newRoom == nil {
		zone := GetZone(room.GetZoneId())
		fmt.Printf("No room found at location %v %v, creating a new one (%s)\n", zone.GetName(), newLocation, character.GetName())

		var err error
		room, err = CreateRoom(GetZone(room.GetZoneId()), newLocation)

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

	return MoveCharacterToLocation(character, GetZone(room.GetZoneId()), room.GetLocation())
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

	rooms := GetRoomsInZone(zone)

	for _, room := range rooms {
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
	zone := GetZone(from.GetZoneId())
	for _, exit := range from.GetExits() {
		nextLocation := from.NextLocation(exit)
		nextRoom := GetRoomByLocation(nextLocation, zone)

		if nextRoom == to {
			return exit
		}
	}

	return database.DirectionNone
}

// vim: nocindent
