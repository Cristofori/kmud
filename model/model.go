package model

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	db "github.com/Cristofori/kmud/database"
	ds "github.com/Cristofori/kmud/datastore"
	"github.com/Cristofori/kmud/utils"
)

// CreateUser creates a new User object in the database and adds it to the model.
// A pointer to the new User object is returned.
func CreateUser(name string, password string) *db.User {
	return db.NewUser(name, password)
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
	var users db.Users

	for _, id := range db.FindAll(db.UserType) {
		users = append(users, ds.Get(id).(*db.User))
	}

	return users
}

// GetUserByName searches for the User object with the given name. Returns a
// nil User if one was not found.
func GetUserByName(username string) *db.User {
	for _, id := range db.Find(db.UserType, "name", utils.FormatName(username)) {
		return ds.Get(id).(*db.User)
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

	ds.Remove(user)
	utils.HandleError(db.DeleteObject(user))
}

// GetPlayerCharacter returns the Character object associated the given Id
func GetPlayerCharacter(id bson.ObjectId) *db.PlayerChar {
	return ds.Get(id).(*db.PlayerChar)
}

func GetNpc(id bson.ObjectId) *db.NonPlayerChar {
	return ds.Get(id).(*db.NonPlayerChar)
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
	for _, id := range db.Find(db.PcType, "name", utils.FormatName(name)) {
		return ds.Get(id).(*db.PlayerChar)
	}

	return nil
}

func GetNpcByName(name string) *db.NonPlayerChar {
	for _, id := range db.Find(db.NpcType, "name", utils.FormatName(name)) {
		return ds.Get(id).(*db.NonPlayerChar)
	}

	return nil
}

func GetNpcs() db.NonPlayerCharList {
	var npcs db.NonPlayerCharList

	for _, id := range db.FindAll(db.NpcType) {
		npcs = append(npcs, ds.Get(id).(*db.NonPlayerChar))
	}

	return npcs
}

/*
func GetAllNpcTemplates() []*db.Character {
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
	var pcs []*db.PlayerChar

	id := user.GetId()

	for _, id := range db.Find(db.PcType, "userid", id) {
		pcs = append(pcs, ds.Get(id).(*db.PlayerChar))
	}

	return pcs
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
	var pcs []*db.PlayerChar

	for _, id := range db.Find(db.PcType, "roomid", room.GetId()) {
		pc := ds.Get(id).(*db.PlayerChar)

		if pc.IsOnline() && pc != except {
			pcs = append(pcs, pc)
		}
	}

	return pcs
}

// NpcsIn returns all of the NPC characters that are in the given room
func NpcsIn(room *db.Room) db.NonPlayerCharList {
	var npcs db.NonPlayerCharList

	for _, id := range db.Find(db.NpcType, "roomid", room.GetId()) {
		npcs = append(npcs, ds.Get(id).(*db.NonPlayerChar))
	}

	return npcs
}

// GetOnlinePlayerCharacters returns a list of all of the characters who are online
func GetOnlinePlayerCharacters() []*db.PlayerChar {
	var pcs []*db.PlayerChar

	for _, id := range db.FindAll(db.PcType) {
		pc := ds.Get(id).(*db.PlayerChar)
		if pc.IsOnline() {
			pcs = append(pcs, pc)
		}
	}

	return pcs
}

// CreatePlayerCharacter creates a new player-controlled Character object in the
// database and adds it to the model.  A pointer to the new character object is
// returned.
func CreatePlayerCharacter(name string, parentUser *db.User, startingRoom *db.Room) *db.PlayerChar {
	return db.NewPlayerChar(name, parentUser.GetId(), startingRoom.GetId())
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
	return db.NewNonPlayerChar(name, room.GetId())
}

/*
func CreateNpcTemplate(name string) *db.Character {
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
func DeletePlayerCharacter(pc *db.PlayerChar) {
	ds.Remove(pc)
	utils.HandleError(db.DeleteObject(pc))
}

func DeleteNpc(npc *db.NonPlayerChar) {
	ds.Remove(npc)
	utils.HandleError(db.DeleteObject(npc))
}

// CreateRoom creates a new Room object in the database and adds it to the model.
// A pointer to the new Room object is returned.
func CreateRoom(zone *db.Zone, location db.Coordinate) (*db.Room, error) {
	existingRoom := GetRoomByLocation(location, zone)
	if existingRoom != nil {
		return nil, errors.New("A room already exists at that location")
	}

	return db.NewRoom(zone.GetId(), location), nil
}

// GetRoom returns the room object associated with the given id
func GetRoom(id bson.ObjectId) *db.Room {
	return ds.Get(id).(*db.Room)
}

// GetRooms returns a list of all of the rooms in the entire model
func GetRooms() db.Rooms {
	var rooms []*db.Room
	for _, id := range db.FindAll(db.RoomType) {
		rooms = append(rooms, ds.Get(id).(*db.Room))
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
	for _, id := range db.Find(db.RoomType, "zoneid", zone.GetId()) {
		// TODO move this check into the DB query
		room := ds.Get(id).(*db.Room)
		if room.GetLocation() == coordinate {
			return room
		}
	}

	return nil
}

// GetZone returns the zone object associated with the given id
func GetZone(zoneId bson.ObjectId) *db.Zone {
	return ds.Get(zoneId).(*db.Zone)
}

// GetZones returns all of the zones in the model
func GetZones() db.Zones {
	var zones db.Zones

	for _, id := range db.FindAll(db.ZoneType) {
		zones = append(zones, ds.Get(id).(*db.Zone))
	}

	return zones
}

// CreateZone creates a new Zone object in the database and adds it to the model.
// A pointer to the new Zone object is returned.
func CreateZone(name string) (*db.Zone, error) {
	if GetZoneByName(name) != nil {
		return nil, errors.New("A zone with that name already exists")
	}

	return db.NewZone(name), nil
}

// Removes the given Zone from the model and the database
func DeleteZone(zone *db.Zone) {
	ds.Remove(zone)
	utils.HandleError(db.DeleteObject(zone))
}

// GetZoneByName name searches for a zone with the given name
func GetZoneByName(name string) *db.Zone {
	for _, id := range db.Find(db.ZoneType, "name", utils.FormatName(name)) {
		return ds.Get(id).(*db.Zone)
	}

	return nil
}

func GetAreas(zone *db.Zone) db.Areas {
	var areas db.Areas
	for _, id := range db.FindAll(db.AreaType) {
		areas = append(areas, ds.Get(id).(*db.Area))
	}

	return areas
}

func GetArea(areaId bson.ObjectId) *db.Area {
	if ds.ContainsId(areaId) {
		return ds.Get(areaId).(*db.Area)
	}

	return nil
}

func CreateArea(name string, zone *db.Zone) (*db.Area, error) {
	if GetAreaByName(name) != nil {
		return nil, errors.New("An area with that name already exists")
	}

	return db.NewArea(name, zone.GetId()), nil
}

func GetAreaByName(name string) *db.Area {
	for _, id := range db.Find(db.AreaType, "name", utils.FormatName(name)) {
		return ds.Get(id).(*db.Area)
	}

	return nil
}

func DeleteArea(area *db.Area) {
	ds.Remove(area)
	utils.HandleError(db.DeleteObject(area))
}

// DeleteRoom removes the given room object from the model and the database. It
// also disables all exits in neighboring rooms that lead to the given room.
func DeleteRoom(room *db.Room) {
	ds.Remove(room)

	utils.HandleError(db.DeleteObject(room))

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
	return ds.Get(id).(*db.User)
}

// CreateItem creates an item object in the database with the given name and
// adds it to the model. It's up to the caller to ensure that the item actually
// gets put somewhere meaningful.
func CreateItem(name string) *db.Item {
	return db.NewItem(name)
}

// GetItem returns the Item object associated the given id
func GetItem(id bson.ObjectId) *db.Item {
	return ds.Get(id).(*db.Item)
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
	ds.Remove(item)

	utils.HandleError(db.DeleteObject(item))
}

func DeleteObject(obj ds.Identifiable) {
	ds.Remove(obj)
	utils.HandleError(db.DeleteObject(obj))
}

// Initializes the global model object and starts up the main event loop
func Init(session db.Session, dbName string) error {
	ds.Init()
	db.Init(session, dbName)

	users := []*db.User{}
	err := db.RetrieveObjects(db.UserType, &users)
	utils.HandleError(err)

	for _, user := range users {
		ds.Set(user)
	}

	pcs := []*db.PlayerChar{}
	err = db.RetrieveObjects(db.PcType, &pcs)
	utils.HandleError(err)

	for _, pc := range pcs {
		pc.SetObjectType(db.PcType)
		ds.Set(pc)
	}

	npcs := []*db.NonPlayerChar{}
	err = db.RetrieveObjects(db.NpcType, &npcs)
	utils.HandleError(err)

	for _, npc := range npcs {
		npc.SetObjectType(db.NpcType)
		ds.Set(npc)
	}

	zones := []*db.Zone{}
	err = db.RetrieveObjects(db.ZoneType, &zones)
	utils.HandleError(err)

	for _, zone := range zones {
		ds.Set(zone)
	}

	areas := []*db.Area{}
	err = db.RetrieveObjects(db.AreaType, &areas)
	utils.HandleError(err)

	for _, area := range areas {
		ds.Set(area)
	}

	rooms := []*db.Room{}
	err = db.RetrieveObjects(db.RoomType, &rooms)
	utils.HandleError(err)

	for _, room := range rooms {
		ds.Set(room)
	}

	items := []*db.Item{}
	err = db.RetrieveObjects(db.ItemType, &items)
	utils.HandleError(err)

	for _, item := range items {
		ds.Set(item)
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
