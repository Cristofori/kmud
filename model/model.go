package model

import (
	"errors"
	"fmt"

	db "github.com/Cristofori/kmud/database"
	ds "github.com/Cristofori/kmud/datastore"
	"github.com/Cristofori/kmud/events"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
	"gopkg.in/mgo.v2/bson"
)

// CreateUser creates a new User object in the database and adds it to the model.
// A pointer to the new User object is returned.
func CreateUser(name string, password string) types.User {
	return db.NewUser(name, password)
}

// GetOrCreateUser attempts to retrieve the existing user from the model by the given name.
// if none exists, then a new one is created with the given credentials.
func GetOrCreateUser(name string, password string) types.User {
	user := GetUserByName(name)

	if user == nil {
		user = CreateUser(name, password)
	}

	return user
}

// GetUsers returns all of the User objects in the model
func GetUsers() types.UserList {
	ids := db.FindAll(types.UserType)
	users := make(types.UserList, len(ids))

	for i, id := range ids {
		users[i] = ds.Get(id).(types.User)
	}

	return users
}

// GetUserByName searches for the User object with the given name. Returns a
// nil User if one was not found.
func GetUserByName(username string) types.User {
	id := db.FindOne(types.UserType, bson.M{"name": utils.FormatName(username)})
	if id != nil {
		return ds.Get(id).(types.User)
	}
	return nil
}

func DeleteUserId(userId types.Id) {
	DeleteUser(GetUser(userId))
}

// Removes the given User from the model. Removes it from the database as well.
func DeleteUser(user types.User) {
	for _, character := range GetUserCharacters(user) {
		DeleteCharacter(character)
	}

	ds.Remove(user)
	utils.HandleError(db.DeleteObject(user))
}

// GetPlayerCharacter returns the Character object associated the given Id
func GetPlayerCharacter(id types.Id) types.PC {
	return ds.Get(id).(types.PC)
}

func GetNpc(id types.Id) types.NPC {
	return ds.Get(id).(types.NPC)
}

func GetCharacterByName(name string) types.Character {
	char := GetPlayerCharacterByName(name)

	if char != nil {
		return char
	}

	npc := GetNpcByName(name)

	if npc != nil {
		return npc
	}

	return nil
}

// GetPlayerCharacaterByName searches for a character with the given name. Returns a
// character object, or nil if it wasn't found.
func GetPlayerCharacterByName(name string) types.PC {
	id := db.FindOne(types.PcType, bson.M{"name": utils.FormatName(name)})
	if id != nil {
		return ds.Get(id).(types.PC)
	}
	return nil
}

func GetNpcByName(name string) types.NPC {
	id := db.FindOne(types.NpcType, bson.M{"name": utils.FormatName(name)})
	if id != nil {
		return ds.Get(id).(types.NPC)
	}
	return nil
}

func GetNpcs() types.NPCList {
	ids := db.FindAll(types.NpcType)
	npcs := make(types.NPCList, len(ids))

	for i, id := range ids {
		npcs[i] = ds.Get(id).(types.NPC)
	}

	return npcs
}

// GetUserCharacters returns all of the Character objects associated with the
// given user id
func GetUserCharacters(user types.User) types.PCList {
	ids := db.Find(types.PcType, bson.M{"userid": user.GetId()})
	pcs := make(types.PCList, len(ids))

	for i, id := range ids {
		pcs[i] = ds.Get(id).(types.PC)
	}

	return pcs
}

func CharactersIn(room types.Room) types.CharacterList {
	var characters types.CharacterList

	players := PlayerCharactersIn(room, nil)
	npcs := NpcsIn(room)

	characters = append(characters, players.Characters()...)
	characters = append(characters, npcs.Characters()...)

	return characters
}

// PlayerCharactersIn returns a list of player characters that are in the given room
func PlayerCharactersIn(room types.Room, except types.Character) types.PCList {
	ids := db.Find(types.PcType, bson.M{"roomid": room.GetId()})
	var pcs types.PCList

	for _, id := range ids {
		pc := ds.Get(id).(types.PC)

		if pc.IsOnline() && pc != except {
			pcs = append(pcs, pc)
		}
	}

	return pcs
}

// NpcsIn returns all of the NPC characters that are in the given room
func NpcsIn(room types.Room) types.NPCList {
	ids := db.Find(types.NpcType, bson.M{"roomid": room.GetId()})
	npcs := make(types.NPCList, len(ids))

	for i, id := range ids {
		npcs[i] = ds.Get(id).(types.NPC)
	}

	return npcs
}

// GetOnlinePlayerCharacters returns a list of all of the characters who are online
func GetOnlinePlayerCharacters() []types.PC {
	var pcs []types.PC

	for _, id := range db.FindAll(types.PcType) {
		pc := ds.Get(id).(types.PC)
		if pc.IsOnline() {
			pcs = append(pcs, pc)
		}
	}

	return pcs
}

// CreatePlayerCharacter creates a new player-controlled Character object in the
// database and adds it to the model.  A pointer to the new character object is
// returned.
func CreatePlayerCharacter(name string, parentUser types.User, startingRoom types.Room) types.PC {
	pc := db.NewPlayerChar(name, parentUser.GetId(), startingRoom.GetId())
	events.Broadcast(events.EnterEvent{Character: pc, RoomId: startingRoom.GetId(), Direction: types.DirectionNone})
	return pc
}

// GetOrCreatePlayerCharacter attempts to retrieve the existing user from the model by the given name.
// if none exists, then a new one is created. If the name matches an NPC (rather than a player)
// then nil will be returned.
func GetOrCreatePlayerCharacter(name string, parentUser types.User, startingRoom types.Room) types.PC {
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
func CreateNpc(name string, roomId types.Id, spawnerId types.Id) types.NPC {
	npc := db.NewNonPlayerChar(name, roomId, spawnerId)
	events.Broadcast(events.EnterEvent{Character: npc, RoomId: roomId, Direction: types.DirectionNone})
	return npc
}

func DeleteCharacterId(id types.Id) {
	DeleteCharacter(ds.Get(id).(types.Character))
}

// DeleteCharacter removes the character associated with the given id from the model and from the database
func DeleteCharacter(c types.Character) {
	ds.Remove(c)
	utils.HandleError(db.DeleteObject(c))
}

// CreateRoom creates a new Room object in the database and adds it to the model.
// A pointer to the new Room object is returned.
func CreateRoom(zone types.Zone, location types.Coordinate) (types.Room, error) {
	existingRoom := GetRoomByLocation(location, zone)
	if existingRoom != nil {
		return nil, errors.New("A room already exists at that location")
	}

	return db.NewRoom(zone.GetId(), location), nil
}

// GetRoom returns the room object associated with the given id
func GetRoom(id types.Id) types.Room {
	return ds.Get(id).(types.Room)
}

// GetRooms returns a list of all of the rooms in the entire model
func GetRooms() types.RoomList {
	ids := db.FindAll(types.RoomType)
	rooms := make(types.RoomList, len(ids))

	for i, id := range ids {
		rooms[i] = ds.Get(id).(types.Room)
	}

	return rooms
}

// GetRoomsInZone returns a slice containing all of the rooms that belong to
// the given zone
func GetRoomsInZone(zone types.Zone) types.RoomList {
	ids := db.Find(types.RoomType, bson.M{"zoneid": zone.GetId()})
	rooms := make(types.RoomList, len(ids))

	for i, id := range ids {
		rooms[i] = ds.Get(id).(types.Room)
	}

	return rooms
}

// GetRoomByLocation searches for the room associated with the given coordinate
// in the given zone. Returns a nil room object if it was not found.
func GetRoomByLocation(coordinate types.Coordinate, zone types.Zone) types.Room {
	id := db.FindOne(types.RoomType, bson.M{
		"zoneid":   zone.GetId(),
		"location": coordinate,
	})
	if id != nil {
		return ds.Get(id).(types.Room)
	}
	return nil
}

// GetZone returns the zone object associated with the given id
func GetZone(zoneId types.Id) types.Zone {
	return ds.Get(zoneId).(types.Zone)
}

// GetZones returns all of the zones in the model
func GetZones() types.ZoneList {
	ids := db.FindAll(types.ZoneType)
	zones := make(types.ZoneList, len(ids))

	for i, id := range ids {
		zones[i] = ds.Get(id).(types.Zone)
	}

	return zones
}

// CreateZone creates a new Zone object in the database and adds it to the model.
// A pointer to the new Zone object is returned.
func CreateZone(name string) (types.Zone, error) {
	if GetZoneByName(name) != nil {
		return nil, errors.New("A zone with that name already exists")
	}

	return db.NewZone(name), nil
}

// Removes the given Zone from the model and the database
func DeleteZone(zone types.Zone) {
	ds.Remove(zone)
	utils.HandleError(db.DeleteObject(zone))
}

// GetZoneByName name searches for a zone with the given name
func GetZoneByName(name string) types.Zone {
	for _, id := range db.Find(types.ZoneType, bson.M{"name": utils.FormatName(name)}) {
		return ds.Get(id).(types.Zone)
	}

	return nil
}

func GetAreas(zone types.Zone) types.AreaList {
	ids := db.FindAll(types.AreaType)
	areas := make(types.AreaList, len(ids))
	for i, id := range ids {
		areas[i] = ds.Get(id).(types.Area)
	}

	return areas
}

func GetArea(areaId types.Id) types.Area {
	if ds.ContainsId(areaId) {
		return ds.Get(areaId).(types.Area)
	}

	return nil
}

func CreateArea(name string, zone types.Zone) (types.Area, error) {
	if GetAreaByName(name) != nil {
		return nil, errors.New("An area with that name already exists")
	}

	return db.NewArea(name, zone.GetId()), nil
}

func GetAreaByName(name string) types.Area {
	id := db.FindOne(types.AreaType, bson.M{"name": utils.FormatName(name)})
	if id != nil {
		return ds.Get(id).(types.Area)
	}
	return nil
}

func DeleteArea(area types.Area) {
	ds.Remove(area)
	utils.HandleError(db.DeleteObject(area))
}

func GetAreaRooms(areaId types.Id) types.RoomList {
	ids := db.Find(types.RoomType, bson.M{"areaid": areaId})
	rooms := make(types.RoomList, len(ids))
	for i, id := range ids {
		rooms[i] = ds.Get(id).(types.Room)
	}
	return rooms

}

// DeleteRoom removes the given room object from the model and the database. It
// also disables all exits in neighboring rooms that lead to the given room.
func DeleteRoom(room types.Room) {
	ds.Remove(room)

	utils.HandleError(db.DeleteObject(room))

	// Disconnect all exits leading to this room
	loc := room.GetLocation()

	updateRoom := func(dir types.Direction) {
		next := loc.Next(dir)
		room := GetRoomByLocation(next, GetZone(room.GetZoneId()))

		if room != nil {
			room.SetExitEnabled(dir.Opposite(), false)
		}
	}

	updateRoom(types.DirectionNorth)
	updateRoom(types.DirectionNorthEast)
	updateRoom(types.DirectionEast)
	updateRoom(types.DirectionSouthEast)
	updateRoom(types.DirectionSouth)
	updateRoom(types.DirectionSouthWest)
	updateRoom(types.DirectionWest)
	updateRoom(types.DirectionNorthWest)
	updateRoom(types.DirectionUp)
	updateRoom(types.DirectionDown)
}

// GetUser returns the User object associated with the given id
func GetUser(id types.Id) types.User {
	return ds.Get(id).(types.User)
}

// CreateItem creates an item object in the database with the given name and
// adds it to the model. It's up to the caller to ensure that the item actually
// gets put somewhere meaningful.
func CreateItem(name string) types.Item {
	return db.NewItem(name)
}

// GetItem returns the Item object associated the given id
func GetItem(id types.Id) types.Item {
	return ds.Get(id).(types.Item)
}

// GetItems returns the Items object associated the given ids
func GetItems(itemIds []types.Id) types.ItemList {
	items := make(types.ItemList, len(itemIds))

	for i, itemId := range itemIds {
		items[i] = GetItem(itemId)
	}

	return items
}

// ItemsIn returns a slice containing all of the items in the given room
func ItemsIn(room types.Room) types.ItemList {
	return GetItems(room.GetItemIds())
}

func DeleteItemId(itemId types.Id) {
	DeleteItem(GetItem(itemId))
}

// DeleteItem removes the item associated with the given id from the
// model and from the database
func DeleteItem(item types.Item) {
	ds.Remove(item)

	utils.HandleError(db.DeleteObject(item))
}

func DeleteObject(obj types.Object) {
	ds.Remove(obj)
	utils.HandleError(db.DeleteObject(obj))
}

func Init(session db.Session, dbName string) {
	ds.Init()
	db.Init(session, dbName)

	users := []*db.User{}
	db.RetrieveObjects(types.UserType, &users)

	for _, user := range users {
		ds.Set(user)
	}

	pcs := []*db.PlayerChar{}
	db.RetrieveObjects(types.PcType, &pcs)

	for _, pc := range pcs {
		pc.SetObjectType(types.PcType)
		ds.Set(pc)
	}

	npcs := []*db.NonPlayerChar{}
	db.RetrieveObjects(types.NpcType, &npcs)

	for _, npc := range npcs {
		npc.SetObjectType(types.NpcType)
		ds.Set(npc)
	}

	spawners := []*db.Spawner{}
	db.RetrieveObjects(types.SpawnerType, &spawners)

	for _, spawner := range spawners {
		spawner.SetObjectType(types.SpawnerType)
		ds.Set(spawner)
	}

	zones := []*db.Zone{}
	db.RetrieveObjects(types.ZoneType, &zones)

	for _, zone := range zones {
		ds.Set(zone)
	}

	areas := []*db.Area{}
	db.RetrieveObjects(types.AreaType, &areas)

	for _, area := range areas {
		ds.Set(area)
	}

	rooms := []*db.Room{}
	db.RetrieveObjects(types.RoomType, &rooms)

	for _, room := range rooms {
		ds.Set(room)
	}

	items := []*db.Item{}
	db.RetrieveObjects(types.ItemType, &items)

	for _, item := range items {
		ds.Set(item)
	}
}

// MoveCharacter attempts to move the character to the given coordinates
// specific by location. Returns an error if there is no room to move to.
func MoveCharacterToLocation(character types.Character, zone types.Zone, location types.Coordinate) (types.Room, error) {
	newRoom := GetRoomByLocation(location, zone)

	if newRoom == nil {
		return nil, errors.New("Invalid location")
	}

	MoveCharacterToRoom(character, newRoom)
	return newRoom, nil
}

// MoveCharacterTo room moves the character to the given room
func MoveCharacterToRoom(character types.Character, newRoom types.Room) {
	oldRoomId := character.GetRoomId()
	character.SetRoomId(newRoom.GetId())

	oldRoom := GetRoom(oldRoomId)

	// Leave
	dir := DirectionBetween(oldRoom, newRoom)
	events.Broadcast(events.LeaveEvent{Character: character, RoomId: oldRoomId, Direction: dir})

	// Enter
	dir = DirectionBetween(newRoom, oldRoom)
	events.Broadcast(events.EnterEvent{Character: character, RoomId: newRoom.GetId(), Direction: dir})
}

// MoveCharacter moves the given character in the given direction. If there is
// no exit in that direction, and error is returned. If there is an exit, but no
// room connected to it, then a room is automatically created for the character
// to move in to.
func MoveCharacter(character types.Character, direction types.Direction) (types.Room, error) {
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
		case types.DirectionNorth:
			room.SetExitEnabled(types.DirectionSouth, true)
		case types.DirectionNorthEast:
			room.SetExitEnabled(types.DirectionSouthWest, true)
		case types.DirectionEast:
			room.SetExitEnabled(types.DirectionWest, true)
		case types.DirectionSouthEast:
			room.SetExitEnabled(types.DirectionNorthWest, true)
		case types.DirectionSouth:
			room.SetExitEnabled(types.DirectionNorth, true)
		case types.DirectionSouthWest:
			room.SetExitEnabled(types.DirectionNorthEast, true)
		case types.DirectionWest:
			room.SetExitEnabled(types.DirectionEast, true)
		case types.DirectionNorthWest:
			room.SetExitEnabled(types.DirectionSouthEast, true)
		case types.DirectionUp:
			room.SetExitEnabled(types.DirectionDown, true)
		case types.DirectionDown:
			room.SetExitEnabled(types.DirectionUp, true)
		default:
			panic("Unexpected code path")
		}
	} else {
		room = newRoom
	}

	return MoveCharacterToLocation(character, GetZone(room.GetZoneId()), room.GetLocation())
}

// BroadcastMessage sends a message to all users that are logged in
func BroadcastMessage(from types.Character, message string) {
	events.Broadcast(events.BroadcastEvent{Character: from, Message: message})
}

// Tell sends a message to the specified character
func Tell(from types.Character, to types.Character, message string) {
	events.Broadcast(events.TellEvent{From: from, To: to, Message: message})
}

// Say sends a message to all characters in the given character's room
func Say(from types.Character, message string) {
	events.Broadcast(events.SayEvent{Character: from, Message: message})
}

// Emote sends an emote message to all characters in the given character's room
func Emote(from types.Character, message string) {
	events.Broadcast(events.EmoteEvent{Character: from, Emote: message})
}

func Login(character types.PC) {
	character.SetOnline(true)
	events.Broadcast(events.LoginEvent{Character: character})
}

func Logout(character types.PC) {
	character.SetOnline(false)
	events.Broadcast(events.LogoutEvent{Character: character})
}

// ZoneCorners returns cordinates that indiate the highest and lowest points of
// the map in 3 dimensions
func ZoneCorners(zone types.Zone) (types.Coordinate, types.Coordinate) {
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

	return types.Coordinate{X: left, Y: top, Z: high},
		types.Coordinate{X: right, Y: bottom, Z: low}
}

func DirectionBetween(from, to types.Room) types.Direction {
	zone := GetZone(from.GetZoneId())
	for _, exit := range from.GetExits() {
		nextLocation := from.NextLocation(exit)
		nextRoom := GetRoomByLocation(nextLocation, zone)

		if nextRoom == to {
			return exit
		}
	}

	return types.DirectionNone
}

func CreateSpawner(name string, areaId types.Id) types.Spawner {
	return db.NewSpawner(name, areaId)
}

func GetSpawners() types.SpawnerList {
	ids := db.FindAll(types.SpawnerType)
	spawners := make(types.SpawnerList, len(ids))

	for i, id := range ids {
		spawners[i] = ds.Get(id).(types.Spawner)
	}

	return spawners
}

func GetSpawner(id types.Id) types.Spawner {
	return ds.Get(id).(types.Spawner)
}

func GetAreaSpawners(areaId types.Id) types.SpawnerList {
	ids := db.Find(types.SpawnerType, bson.M{"areaid": areaId})
	spawners := make(types.SpawnerList, len(ids))
	for i, id := range ids {
		spawners[i] = ds.Get(id).(types.Spawner)
	}
	return spawners
}

func GetSpawnerNpcs(spawnerId types.Id) types.NPCList {
	ids := db.Find(types.NpcType, bson.M{"spawnerid": spawnerId})
	npcs := make(types.NPCList, len(ids))
	for i, id := range ids {
		npcs[i] = ds.Get(id).(types.NPC)
	}
	return npcs
}

// vim: nocindent
