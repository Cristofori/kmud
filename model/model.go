package model

import (
	"errors"
	"fmt"
	"sort"

	db "github.com/Cristofori/kmud/database"
	"github.com/Cristofori/kmud/events"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
	"gopkg.in/mgo.v2/bson"
)

func CreateUser(name string, password string, admin bool) types.User {
	return db.NewUser(name, password, admin)
}

func GetOrCreateUser(name string, password string, admin bool) types.User {
	user := GetUserByName(name)

	if user == nil {
		user = CreateUser(name, password, admin)
	}

	return user
}

func GetUsers() types.UserList {
	ids := db.FindAll(types.UserType)
	users := make(types.UserList, len(ids))

	for i, id := range ids {
		users[i] = GetUser(id)
	}

	return users
}

func UserCount() int {
	return len(db.FindAll(types.UserType))
}

func GetUserByName(username string) types.User {
	id := db.FindOne(types.UserType, bson.M{"name": utils.FormatName(username)})
	if id != nil {
		return GetUser(id)
	}
	return nil
}

func DeleteUser(userId types.Id) {
	for _, character := range GetUserCharacters(userId) {
		DeleteCharacter(character.GetId())
	}

	db.DeleteObject(userId)
}

func GetPlayerCharacter(id types.Id) types.PC {
	return db.Retrieve(id, types.PcType).(types.PC)
}

func GetNpc(id types.Id) types.NPC {
	return db.Retrieve(id, types.NpcType).(types.NPC)
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

func GetPlayerCharacterByName(name string) types.PC {
	id := db.FindOne(types.PcType, bson.M{"name": utils.FormatName(name)})
	if id != nil {
		return GetPlayerCharacter(id)
	}
	return nil
}

func GetNpcByName(name string) types.NPC {
	id := db.FindOne(types.NpcType, bson.M{"name": utils.FormatName(name)})
	if id != nil {
		return GetNpc(id)
	}
	return nil
}

func GetNpcs() types.NPCList {
	ids := db.FindAll(types.NpcType)
	npcs := make(types.NPCList, len(ids))

	for i, id := range ids {
		npcs[i] = GetNpc(id)
	}

	return npcs
}

func GetUserCharacters(userId types.Id) types.PCList {
	ids := db.Find(types.PcType, bson.M{"userid": userId})
	pcs := make(types.PCList, len(ids))

	for i, id := range ids {
		pcs[i] = GetPlayerCharacter(id)
	}

	return pcs
}

func CharactersIn(roomId types.Id) types.CharacterList {
	var characters types.CharacterList

	players := PlayerCharactersIn(roomId, nil)
	npcs := NpcsIn(roomId)

	characters = append(characters, players.Characters()...)
	characters = append(characters, npcs.Characters()...)

	return characters
}

func PlayerCharactersIn(roomId types.Id, except types.Id) types.PCList {
	ids := db.Find(types.PcType, bson.M{"roomid": roomId})
	var pcs types.PCList

	for _, id := range ids {
		pc := GetPlayerCharacter(id)

		if pc.IsOnline() && id != except {
			pcs = append(pcs, pc)
		}
	}

	return pcs
}

func NpcsIn(roomId types.Id) types.NPCList {
	ids := db.Find(types.NpcType, bson.M{"roomid": roomId})
	npcs := make(types.NPCList, len(ids))

	for i, id := range ids {
		npcs[i] = GetNpc(id)
	}

	return npcs
}

func GetOnlinePlayerCharacters() []types.PC {
	var pcs []types.PC

	for _, id := range db.FindAll(types.PcType) {
		pc := GetPlayerCharacter(id)
		if pc.IsOnline() {
			pcs = append(pcs, pc)
		}
	}

	return pcs
}

func CreatePlayerCharacter(name string, userId types.Id, startingRoom types.Room) types.PC {
	pc := db.NewPc(name, userId, startingRoom.GetId())
	events.Broadcast(events.EnterEvent{Character: pc, RoomId: startingRoom.GetId(), Direction: types.DirectionNone})
	return pc
}

func GetOrCreatePlayerCharacter(name string, userId types.Id, startingRoom types.Room) types.PC {
	player := GetPlayerCharacterByName(name)
	npc := GetNpcByName(name)

	if player == nil && npc == nil {
		player = CreatePlayerCharacter(name, userId, startingRoom)
	} else if npc != nil {
		return nil
	}

	return player
}

func CreateNpc(name string, roomId types.Id, spawnerId types.Id) types.NPC {
	npc := db.NewNpc(name, roomId, spawnerId)
	events.Broadcast(events.EnterEvent{Character: npc, RoomId: roomId, Direction: types.DirectionNone})
	return npc
}

func DeleteCharacter(charId types.Id) {
	// TODO: Delete (or drop) inventory
	db.DeleteObject(charId)
}

func CreateRoom(zone types.Zone, location types.Coordinate) (types.Room, error) {
	existingRoom := GetRoomByLocation(location, zone.GetId())
	if existingRoom != nil {
		return nil, errors.New("A room already exists at that location")
	}

	return db.NewRoom(zone.GetId(), location), nil
}

func GetRoom(id types.Id) types.Room {
	return db.Retrieve(id, types.RoomType).(types.Room)
}

func GetRooms() types.RoomList {
	ids := db.FindAll(types.RoomType)
	rooms := make(types.RoomList, len(ids))

	for i, id := range ids {
		rooms[i] = GetRoom(id)
	}

	return rooms
}

func GetRoomsInZone(zoneId types.Id) types.RoomList {
	zone := GetZone(zoneId)
	ids := db.Find(types.RoomType, bson.M{"zoneid": zone.GetId()})
	rooms := make(types.RoomList, len(ids))

	for i, id := range ids {
		rooms[i] = GetRoom(id)
	}

	return rooms
}

func GetRoomByLocation(coordinate types.Coordinate, zoneId types.Id) types.Room {
	id := db.FindOne(types.RoomType, bson.M{
		"zoneid":   zoneId,
		"location": coordinate,
	})
	if id != nil {
		return GetRoom(id)
	}
	return nil
}

func GetNeighbors(room types.Room) []types.Room {
	neighbors := []types.Room{}

	for _, dir := range room.GetExits() {
		coords := room.NextLocation(dir)
		neighbor := GetRoomByLocation(coords, room.GetZoneId())
		if neighbor != nil {
			neighbors = append(neighbors, neighbor)
		}
	}

	for _, id := range room.GetLinks() {
		neighbor := GetRoom(id)
		if neighbor != nil {
			neighbors = append(neighbors, neighbor)
		}
	}

	return neighbors
}

func GetZone(id types.Id) types.Zone {
	return db.Retrieve(id, types.ZoneType).(types.Zone)
}

func GetZones() types.ZoneList {
	ids := db.FindAll(types.ZoneType)
	zones := make(types.ZoneList, len(ids))

	for i, id := range ids {
		zones[i] = GetZone(id)
	}

	return zones
}

func CreateZone(name string) (types.Zone, error) {
	if GetZoneByName(name) != nil {
		return nil, errors.New("A zone with that name already exists")
	}

	return db.NewZone(name), nil
}

func DeleteZone(zoneId types.Id) {
	rooms := GetRoomsInZone(zoneId)

	for _, room := range rooms {
		DeleteRoom(room)
	}

	db.DeleteObject(zoneId)
}

func GetZoneByName(name string) types.Zone {
	for _, id := range db.Find(types.ZoneType, bson.M{"name": utils.FormatName(name)}) {
		return GetZone(id)
	}

	return nil
}

func GetAreas(zone types.Zone) types.AreaList {
	ids := db.FindAll(types.AreaType)
	areas := make(types.AreaList, len(ids))
	for i, id := range ids {
		areas[i] = GetArea(id)
	}

	return areas
}

func GetArea(id types.Id) types.Area {
	return db.Retrieve(id, types.AreaType).(types.Area)
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
		return GetArea(id)
	}
	return nil
}

func DeleteArea(areaId types.Id) {
	// TODO - Remove room references to area
	db.DeleteObject(areaId)
}

func GetAreaRooms(areaId types.Id) types.RoomList {
	ids := db.Find(types.RoomType, bson.M{"areaid": areaId})
	rooms := make(types.RoomList, len(ids))
	for i, id := range ids {
		rooms[i] = GetRoom(id)
	}
	return rooms

}

func DeleteRoom(room types.Room) {
	db.DeleteObject(room.GetId())

	// Disconnect all exits leading to this room
	loc := room.GetLocation()

	updateRoom := func(dir types.Direction) {
		next := loc.Next(dir)
		room := GetRoomByLocation(next, room.GetZoneId())

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

func GetUser(id types.Id) types.User {
	return db.Retrieve(id, types.UserType).(types.User)
}

func CreateTemplate(name string) types.Template {
	return db.NewTemplate(name)
}

func GetAllTemplates() types.TemplateList {
	ids := db.FindAll(types.TemplateType)
	templates := make(types.TemplateList, len(ids))
	for i, id := range ids {
		templates[i] = GetTemplate(id)
	}
	sort.Sort(templates)
	return templates
}

func GetTemplate(id types.Id) types.Template {
	return db.Retrieve(id, types.TemplateType).(types.Template)
}

func CreateItem(templateId types.Id) types.Item {
	return db.NewItem(templateId)
}

func GetItem(id types.Id) types.Item {
	return db.Retrieve(id, types.ItemType).(types.Item)
}

func DeleteItemId(itemId types.Id) {
	DeleteItem(itemId)
}

func DeleteItem(itemId types.Id) {
	db.DeleteObject(itemId)
}

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

func MoveCharacter(character types.Character, direction types.Direction) error {
	room := GetRoom(character.GetRoomId())

	if room == nil {
		return errors.New("Character doesn't appear to be in any room")
	}

	if !room.HasExit(direction) {
		return errors.New("Attempted to move through an exit that the room does not contain")
	}

	if room.IsLocked(direction) {
		return errors.New("That way is locked")
	}

	newLocation := room.NextLocation(direction)
	newRoom := GetRoomByLocation(newLocation, room.GetZoneId())

	if newRoom == nil {
		zone := GetZone(room.GetZoneId())
		fmt.Printf("No room found at location %v %v, creating a new one (%s)\n", zone.GetName(), newLocation, character.GetName())

		var err error
		newRoom, err = CreateRoom(GetZone(room.GetZoneId()), newLocation)
		newRoom.SetTitle(room.GetTitle())
		newRoom.SetDescription(room.GetDescription())

		if err != nil {
			return err
		}

		switch direction {
		case types.DirectionNorth:
			newRoom.SetExitEnabled(types.DirectionSouth, true)
		case types.DirectionNorthEast:
			newRoom.SetExitEnabled(types.DirectionSouthWest, true)
		case types.DirectionEast:
			newRoom.SetExitEnabled(types.DirectionWest, true)
		case types.DirectionSouthEast:
			newRoom.SetExitEnabled(types.DirectionNorthWest, true)
		case types.DirectionSouth:
			newRoom.SetExitEnabled(types.DirectionNorth, true)
		case types.DirectionSouthWest:
			newRoom.SetExitEnabled(types.DirectionNorthEast, true)
		case types.DirectionWest:
			newRoom.SetExitEnabled(types.DirectionEast, true)
		case types.DirectionNorthWest:
			newRoom.SetExitEnabled(types.DirectionSouthEast, true)
		case types.DirectionUp:
			newRoom.SetExitEnabled(types.DirectionDown, true)
		case types.DirectionDown:
			newRoom.SetExitEnabled(types.DirectionUp, true)
		default:
			panic("Unexpected code path")
		}
	}

	MoveCharacterToRoom(character, newRoom)
	return nil
}

func BroadcastMessage(from types.Character, message string) {
	events.Broadcast(events.BroadcastEvent{Character: from, Message: message})
}

func Tell(from types.Character, to types.Character, message string) {
	events.Broadcast(events.TellEvent{From: from, To: to, Message: message})
}

func Say(from types.Character, message string) {
	events.Broadcast(events.SayEvent{Character: from, Message: message})
}

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

func ZoneCorners(zone types.Zone) (types.Coordinate, types.Coordinate) {
	var top int
	var bottom int
	var left int
	var right int
	var high int
	var low int

	rooms := GetRoomsInZone(zone.GetId())

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
	for _, exit := range from.GetExits() {
		nextLocation := from.NextLocation(exit)
		nextRoom := GetRoomByLocation(nextLocation, from.GetZoneId())

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
		spawners[i] = GetSpawner(id)
	}

	return spawners
}

func GetSpawner(id types.Id) types.Spawner {
	return db.Retrieve(id, types.SpawnerType).(types.Spawner)
}

func GetAreaSpawners(areaId types.Id) types.SpawnerList {
	ids := db.Find(types.SpawnerType, bson.M{"areaid": areaId})
	spawners := make(types.SpawnerList, len(ids))
	for i, id := range ids {
		spawners[i] = GetSpawner(id)
	}
	return spawners
}

func GetSpawnerNpcs(spawnerId types.Id) types.NPCList {
	ids := db.Find(types.NpcType, bson.M{"spawnerid": spawnerId})
	npcs := make(types.NPCList, len(ids))
	for i, id := range ids {
		npcs[i] = GetNpc(id)
	}
	return npcs
}

func GetSkill(id types.Id) types.Skill {
	return db.Retrieve(id, types.SkillType).(types.Skill)
}

func GetSkillByName(name string) types.Skill {
	id := db.FindOne(types.SkillType, bson.M{"name": utils.FormatName(name)})
	if id != nil {
		return GetSkill(id)
	}
	return nil
}

func GetAllSkills() types.SkillList {
	ids := db.FindAll(types.SkillType)
	skills := make(types.SkillList, len(ids))
	for i, id := range ids {
		skills[i] = GetSkill(id)
	}
	return skills
}

func GetSkills(SkillIds []types.Id) types.SkillList {
	skills := make(types.SkillList, len(SkillIds))
	for i, id := range SkillIds {
		skills[i] = GetSkill(id)
	}
	return skills
}

func CreateSkill(name string) types.Skill {
	return db.NewSkill(name, 10)
}

func DeleteSkill(id types.Id) {
	db.DeleteObject(id)
}

func StoreIn(roomId types.Id) types.Store {
	id := db.FindOne(types.StoreType, bson.M{"roomid": roomId})

	if id != nil {
		return GetStore(id)
	}
	return nil
}

func GetStore(id types.Id) types.Store {
	return db.Retrieve(id, types.StoreType).(types.Store)
}

func CreateStore(name string, roomId types.Id) types.Store {
	return db.NewStore(name, roomId)
}

func DeleteStore(id types.Id) {
	// TODO - Delete or drop items
	db.DeleteObject(id)
}

func GetWorld() types.World {
	id := db.FindOne(types.WorldType, bson.M{})
	if id == nil {
		return db.NewWorld()
	}
	return db.Retrieve(id, types.WorldType).(types.World)
}
