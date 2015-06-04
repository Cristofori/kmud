package model

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	db "kmud/database"
	"kmud/utils"
	"sync"
)

var _objects map[bson.ObjectId]interface{}

var _chars map[bson.ObjectId]*db.PlayerChar
var _npcs map[bson.ObjectId]*db.NonPlayerChar
var _zones map[bson.ObjectId]*db.Zone
var _areas map[bson.ObjectId]*db.Area
var _rooms map[bson.ObjectId]*db.Room
var _items map[bson.ObjectId]*db.Item

var mutex sync.RWMutex

// CreateUser creates a new User object in the database and adds it to the model.
// A pointer to the new User object is returned.
func CreateUser(name string, password string) *db.User {
	mutex.Lock()
	defer mutex.Unlock()

	user := db.NewUser(name, password)
	_objects[user.GetId()] = user

	return user
}

// GetOrCreateUser attempts to retrieve the existing user from the model by the given name.
// if none exists, then a new one is created with the given credentials.
func GetOrCreateUser(name string, password string) *db.User {
	user := GetUserByName(name)

	if user == nil {
		user = CreateUser(name, password)
	}

	return user
}

// GetUsers returns all of the User objects in the model
func GetUsers() db.Users {
	mutex.RLock()
	defer mutex.RUnlock()

	var users db.Users

	for _, id := range db.FindAll(db.UserType) {
		users = append(users, _objects[id].(*db.User))
	}

	return users
}

// GetUserByName searches for the User object with the given name. Returns a
// nil User if one was not found.
func GetUserByName(username string) *db.User {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, id := range db.Find(db.UserType, "name", utils.FormatName(username)) {
		return _objects[id].(*db.User)
	}

	return nil
}

func DeleteUserId(userId bson.ObjectId) {
	DeleteUser(GetUser(userId))
}

// Removes the given User from the model. Removes it from the database as well.
func DeleteUser(user *db.User) {
	for _, character := range GetUserCharacters(user) {
		DeletePlayerCharacter(character)
	}

	mutex.Lock()
	defer mutex.Unlock()

	delete(_objects, user.GetId())
	utils.HandleError(db.DeleteObject(user))
}

// GetPlayerCharacter returns the Character object associated the given Id
func GetPlayerCharacter(id bson.ObjectId) *db.PlayerChar {
	mutex.RLock()
	defer mutex.RUnlock()

	return _chars[id]
}

func GetNpc(id bson.ObjectId) *db.NonPlayerChar {
	mutex.RLock()
	defer mutex.RUnlock()

	return _npcs[id]
}

func GetCharacterByName(name string) *db.Character {
	char := GetPlayerCharacterByName(name)

	if char != nil {
		return &char.Character
	}

	npc := GetNpcByName(name)

	if npc != nil {
		return &npc.Character
	}

	return nil
}

// GetPlayerCharacaterByName searches for a character with the given name. Returns a
// character object, or nil if it wasn't found.
func GetPlayerCharacterByName(name string) *db.PlayerChar {
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

func GetNpcByName(name string) *db.NonPlayerChar {
	mutex.RLock()
	defer mutex.RUnlock()

	name = utils.Simplify(name)

	for _, npc := range _npcs {
		if npc.GetName() == name {
			return npc
		}
	}

	return nil
}

func GetNpcs() []*db.NonPlayerChar {
	mutex.RLock()
	defer mutex.RUnlock()

	npcs := []*db.NonPlayerChar{}

	for _, npc := range _npcs {
		npcs = append(npcs, npc)
	}

	return npcs
}

/*
func GetAllNpcTemplates() []*db.Character {
	mutex.RLock()
	defer mutex.RUnlock()

	templates := []*db.Character{}

	for _, character := range _chars {
		if character.IsNpcTemplate() {
			templates = append(templates, character)
		}
	}

	return templates
}
*/

// GetUserCharacters returns all of the Character objects associated with the
// given user id
func GetUserCharacters(user *db.User) []*db.PlayerChar {
	mutex.RLock()
	defer mutex.RUnlock()

	var characters []*db.PlayerChar

	for _, character := range _chars {
		if character.GetUserId() == user.GetId() {
			characters = append(characters, character)
		}
	}

	return characters
}

func CharactersIn(room *db.Room) db.CharacterList {
	var characters db.CharacterList

	players := PlayerCharactersIn(room, nil)
	npcs := NpcsIn(room)

	characters = append(characters, players.Characters()...)
	characters = append(characters, npcs.Characters()...)

	return characters
}

// PlayerCharactersIn returns a list of player characters that are in the given room
func PlayerCharactersIn(room *db.Room, except *db.PlayerChar) db.PlayerCharList {
	mutex.RLock()
	defer mutex.RUnlock()

	var characters []*db.PlayerChar

	for _, char := range _chars {
		if char.GetRoomId() == room.GetId() && char.IsOnline() {
			characters = append(characters, char)
		}
	}

	return characters
}

// NpcsIn returns all of the NPC characters that are in the given room
func NpcsIn(room *db.Room) db.NonPlayerCharList {
	mutex.RLock()
	defer mutex.RUnlock()

	var npcs []*db.NonPlayerChar

	for _, npc := range _npcs {
		if npc.GetRoomId() == room.GetId() {
			npcs = append(npcs, npc)
		}
	}

	return npcs
}

// GetOnlinePlayerCharacters returns a list of all of the characters who are online
func GetOnlinePlayerCharacters() []*db.PlayerChar {
	mutex.RLock()
	defer mutex.RUnlock()

	var characters []*db.PlayerChar

	for _, char := range _chars {
		if char.IsOnline() {
			characters = append(characters, char)
		}
	}

	return characters
}

// CreatePlayerCharacter creates a new player-controlled Character object in the
// database and adds it to the model.  A pointer to the new character object is
// returned.
func CreatePlayerCharacter(name string, parentUser *db.User, startingRoom *db.Room) *db.PlayerChar {
	mutex.Lock()
	defer mutex.Unlock()

	character := db.NewPlayerChar(name, parentUser.GetId(), startingRoom.GetId())
	_chars[character.GetId()] = character

	return character
}

// GetOrCreatePlayerCharacter attempts to retrieve the existing user from the model by the given name.
// if none exists, then a new one is created. If the name matches an NPC (rather than a player)
// then nil will be returned.
func GetOrCreatePlayerCharacter(name string, parentUser *db.User, startingRoom *db.Room) *db.PlayerChar {
	player := GetPlayerCharacterByName(name)
	npc := GetNpcByName(name)

	if player == nil && npc == nil {
		player = CreatePlayerCharacter(name, parentUser, startingRoom)
	} else if npc != nil {
		return nil
	}

	return player
}

// CreateNpc is a convenience function for creating a new character object that
// is an NPC (as opposed to an actual player-controlled character)
func CreateNpc(name string, room *db.Room) *db.NonPlayerChar {
	mutex.Lock()
	defer mutex.Unlock()

	npc := db.NewNonPlayerChar(name, room.GetId())
	_npcs[npc.GetId()] = npc

	return npc
}

/*
func CreateNpcTemplate(name string) *db.Character {
	mutex.Lock()
	defer mutex.Unlock()

	template := db.NewNpcTemplate(name)
	_chars[template.GetId()] = template

	return template
}
*/

func DeletePlayerCharacterId(id bson.ObjectId) {
	DeletePlayerCharacter(GetPlayerCharacter(id))
}

func DeleteNpcId(id bson.ObjectId) {
	DeleteNpc(GetNpc(id))
}

// DeletePlayerCharacter removes the character (either NPC or player-controlled)
// associated with the given id from the model and from the database
func DeletePlayerCharacter(character *db.PlayerChar) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(_chars, character.GetId())
	utils.HandleError(db.DeleteObject(character))
}

func DeleteNpc(npc *db.NonPlayerChar) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(_npcs, npc.GetId())
	utils.HandleError(db.DeleteObject(npc))
}

// CreateRoom creates a new Room object in the database and adds it to the model.
// A pointer to the new Room object is returned.
func CreateRoom(zone *db.Zone, location db.Coordinate) (*db.Room, error) {
	existingRoom := GetRoomByLocation(location, zone)
	if existingRoom != nil {
		return nil, errors.New("A room already exists at that location")
	}

	mutex.Lock()
	defer mutex.Unlock()

	room := db.NewRoom(zone.GetId(), location)
	_rooms[room.GetId()] = room

	return room, nil
}

// GetRoom returns the room object associated with the given id
func GetRoom(id bson.ObjectId) *db.Room {
	mutex.RLock()
	defer mutex.RUnlock()

	return _rooms[id]
}

// GetRooms returns a list of all of the rooms in the entire model
func GetRooms() []*db.Room {
	mutex.RLock()
	defer mutex.RUnlock()

	var rooms []*db.Room

	for _, room := range _rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

// GetRoomsInZone returns a slice containing all of the rooms that belong to
// the given zone
func GetRoomsInZone(zone *db.Zone) []*db.Room {
	allRooms := GetRooms()

	var rooms []*db.Room

	for _, room := range allRooms {
		if room.GetZoneId() == zone.GetId() {
			rooms = append(rooms, room)
		}
	}

	return rooms
}

// GetRoomByLocation searches for the room associated with the given coordinate
// in the given zone. Returns a nil room object if it was not found.
func GetRoomByLocation(coordinate db.Coordinate, zone *db.Zone) *db.Room {
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
func GetZone(zoneId bson.ObjectId) *db.Zone {
	mutex.RLock()
	defer mutex.RUnlock()

	return _zones[zoneId]
}

// GetZones returns all of the zones in the model
func GetZones() db.Zones {
	mutex.RLock()
	defer mutex.RUnlock()

	var zones db.Zones

	for _, zone := range _zones {
		zones = append(zones, zone)
	}

	return zones
}

// CreateZone creates a new Zone object in the database and adds it to the model.
// A pointer to the new Zone object is returned.
func CreateZone(name string) (*db.Zone, error) {
	if GetZoneByName(name) != nil {
		return nil, errors.New("A zone with that name already exists")
	}

	mutex.Lock()
	defer mutex.Unlock()

	zone := db.NewZone(name)
	_zones[zone.GetId()] = zone

	return zone, nil
}

// Removes the given Zone from the model. Removes it from the database as well.
func DeleteZone(zone *db.Zone) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(_zones, zone.GetId())
	utils.HandleError(db.DeleteObject(zone))
}

// GetZoneByName name searches for a zone with the given name, returns a zone
// object and whether or not it was found
func GetZoneByName(name string) *db.Zone {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, zone := range _zones {
		if utils.Compare(zone.GetName(), name) {
			return zone
		}
	}

	return nil
}

func GetAreas(zone *db.Zone) db.Areas {
	mutex.RLock()
	defer mutex.RUnlock()

	var areas db.Areas

	for _, area := range _areas {
		if area.GetZoneId() == zone.GetId() {
			areas = append(areas, area)
		}
	}

	return areas
}

func GetArea(areaId bson.ObjectId) *db.Area {
	mutex.RLock()
	defer mutex.RUnlock()

	return _areas[areaId]
}

func CreateArea(name string, zone *db.Zone) (*db.Area, error) {
	if GetAreaByName(name) != nil {
		return nil, errors.New("An area with that name already exists")
	}

	mutex.Lock()
	defer mutex.Unlock()

	area := db.NewArea(name, zone.GetId())
	_areas[area.GetId()] = area

	return area, nil
}

func GetAreaByName(name string) *db.Area {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, area := range _areas {
		if utils.Compare(area.GetName(), name) {
			return area
		}
	}

	return nil
}

func DeleteArea(area *db.Area) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(_areas, area.GetId())
	utils.HandleError(db.DeleteObject(area))
}

// DeleteRoom removes the given room object from the model and the database. It
// also disables all exits in neighboring rooms that lead to the given room.
func DeleteRoom(room *db.Room) {
	mutex.Lock()
	delete(_rooms, room.GetId())

	utils.HandleError(db.DeleteObject(room))
	mutex.Unlock()

	// Disconnect all exits leading to this room
	loc := room.GetLocation()

	updateRoom := func(dir db.Direction) {
		next := loc.Next(dir)
		room := GetRoomByLocation(next, GetZone(room.GetZoneId()))

		if room != nil {
			room.SetExitEnabled(dir.Opposite(), false)
		}
	}

	updateRoom(db.DirectionNorth)
	updateRoom(db.DirectionNorthEast)
	updateRoom(db.DirectionEast)
	updateRoom(db.DirectionSouthEast)
	updateRoom(db.DirectionSouth)
	updateRoom(db.DirectionSouthWest)
	updateRoom(db.DirectionWest)
	updateRoom(db.DirectionNorthWest)
	updateRoom(db.DirectionUp)
	updateRoom(db.DirectionDown)
}

// GetUser returns the User object associated with the given id
func GetUser(id bson.ObjectId) *db.User {
	mutex.RLock()
	defer mutex.RUnlock()

	return _objects[id].(*db.User)
}

// CreateItem creates an item object in the database with the given name and
// adds it to the model. It's up to the caller to ensure that the item actually
// gets put somewhere meaningful.
func CreateItem(name string) *db.Item {
	mutex.Lock()
	defer mutex.Unlock()

	item := db.NewItem(name)
	_items[item.GetId()] = item

	return item
}

// GetItem returns the Item object associated the given id
func GetItem(id bson.ObjectId) *db.Item {
	mutex.RLock()
	defer mutex.RUnlock()

	return _items[id]
}

// GetItems returns the Items object associated the given ids
func GetItems(itemIds []bson.ObjectId) []*db.Item {
	var items []*db.Item

	for _, itemId := range itemIds {
		items = append(items, GetItem(itemId))
	}

	return items
}

// ItemsIn returns a slice containing all of the items in the given room
func ItemsIn(room *db.Room) []*db.Item {
	return GetItems(room.GetItemIds())
}

func DeleteItemId(itemId bson.ObjectId) {
	DeleteItem(GetItem(itemId))
}

// DeleteItem removes the item associated with the given id from the
// model and from the database
func DeleteItem(item *db.Item) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(_items, item.GetId())

	utils.HandleError(db.DeleteObject(item))
}

// Initializes the global model object and starts up the main event loop
func Init(session db.Session) error {
	db.Init(session)

	_objects = map[bson.ObjectId]interface{}{}

	_chars = map[bson.ObjectId]*db.PlayerChar{}
	_npcs = map[bson.ObjectId]*db.NonPlayerChar{}
	_zones = map[bson.ObjectId]*db.Zone{}
	_areas = map[bson.ObjectId]*db.Area{}
	_rooms = map[bson.ObjectId]*db.Room{}
	_items = map[bson.ObjectId]*db.Item{}

	users := []*db.User{}
	err := db.RetrieveObjects(db.UserType, &users)
	utils.HandleError(err)

	for _, user := range users {
		_objects[user.GetId()] = user
	}

	characters := []*db.PlayerChar{}
	err = db.RetrieveObjects(db.PcType, &characters)
	utils.HandleError(err)

	for _, character := range characters {
		_chars[character.GetId()] = character
	}

	npcs := []*db.NonPlayerChar{}
	err = db.RetrieveObjects(db.NpcType, &npcs)
	utils.HandleError(err)

	for _, npc := range npcs {
		_npcs[npc.GetId()] = npc
	}

	zones := []*db.Zone{}
	err = db.RetrieveObjects(db.ZoneType, &zones)
	utils.HandleError(err)

	for _, zone := range zones {
		_zones[zone.GetId()] = zone
	}

	areas := []*db.Area{}
	err = db.RetrieveObjects(db.AreaType, &areas)
	utils.HandleError(err)

	for _, area := range areas {
		_areas[area.GetId()] = area
	}

	rooms := []*db.Room{}
	err = db.RetrieveObjects(db.RoomType, &rooms)
	utils.HandleError(err)

	for _, room := range rooms {
		_rooms[room.GetId()] = room
	}

	items := []*db.Item{}
	err = db.RetrieveObjects(db.ItemType, &items)
	utils.HandleError(err)

	for _, item := range items {
		_items[item.GetId()] = item
	}

	// Start the event loop
	go eventLoop()

	fights = map[*db.Character]*db.Character{}
	go combatLoop()

	return err
}

// MoveCharacter attempts to move the character to the given coordinates
// specific by location. Returns an error if there is no room to move to.
func MoveCharacterToLocation(character *db.Character, zone *db.Zone, location db.Coordinate) (*db.Room, error) {
	newRoom := GetRoomByLocation(location, zone)

	if newRoom == nil {
		return nil, errors.New("Invalid location")
	}

	MoveCharacterToRoom(character, newRoom)
	return newRoom, nil
}

// MoveCharacterTo room moves the character to the given room
func MoveCharacterToRoom(character *db.Character, newRoom *db.Room) {
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
func MoveCharacter(character *db.Character, direction db.Direction) (*db.Room, error) {
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
		case db.DirectionNorth:
			room.SetExitEnabled(db.DirectionSouth, true)
		case db.DirectionNorthEast:
			room.SetExitEnabled(db.DirectionSouthWest, true)
		case db.DirectionEast:
			room.SetExitEnabled(db.DirectionWest, true)
		case db.DirectionSouthEast:
			room.SetExitEnabled(db.DirectionNorthWest, true)
		case db.DirectionSouth:
			room.SetExitEnabled(db.DirectionNorth, true)
		case db.DirectionSouthWest:
			room.SetExitEnabled(db.DirectionNorthEast, true)
		case db.DirectionWest:
			room.SetExitEnabled(db.DirectionEast, true)
		case db.DirectionNorthWest:
			room.SetExitEnabled(db.DirectionSouthEast, true)
		case db.DirectionUp:
			room.SetExitEnabled(db.DirectionDown, true)
		case db.DirectionDown:
			room.SetExitEnabled(db.DirectionUp, true)
		default:
			panic("Unexpected code path")
		}
	} else {
		room = newRoom
	}

	return MoveCharacterToLocation(character, GetZone(room.GetZoneId()), room.GetLocation())
}

// BroadcastMessage sends a message to all users that are logged in
func BroadcastMessage(from *db.Character, message string) {
	queueEvent(BroadcastEvent{from, message})
}

// Tell sends a message to the specified character
func Tell(from *db.Character, to *db.Character, message string) {
	queueEvent(TellEvent{from, to, message})
}

// Say sends a message to all characters in the given character's room
func Say(from *db.Character, message string) {
	queueEvent(SayEvent{from, message})
}

// Emote sends an emote message to all characters in the given character's room
func Emote(from *db.Character, message string) {
	queueEvent(EmoteEvent{from, message})
}

// ZoneCorners returns cordinates that indiate the highest and lowest points of
// the map in 3 dimensions
func ZoneCorners(zone *db.Zone) (db.Coordinate, db.Coordinate) {
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

	return db.Coordinate{X: left, Y: top, Z: high},
		db.Coordinate{X: right, Y: bottom, Z: low}
}

// Returns the exit direction of the given room if it is adjacent, otherwise DirectionNone
func DirectionBetween(from, to *db.Room) db.Direction {
	zone := GetZone(from.GetZoneId())
	for _, exit := range from.GetExits() {
		nextLocation := from.NextLocation(exit)
		nextRoom := GetRoomByLocation(nextLocation, zone)

		if nextRoom == to {
			return exit
		}
	}

	return db.DirectionNone
}

// vim: nocindent
