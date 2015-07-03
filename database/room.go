package database

import (
	"fmt"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strings"
)

type Exit struct {
	Locked bool
}

type Room struct {
	DbObject `bson:",inline"`

	ZoneId      types.Id
	AreaId      types.Id `bson:",omitempty"`
	Title       string
	Description string
	Items       map[string]bool
	Links       map[string]types.Id
	Location    types.Coordinate

	Exits map[types.Direction]*Exit

	Properties map[string]string
}

func NewRoom(zoneId types.Id, location types.Coordinate) *Room {
	var room Room

	room.Title = "The Void"
	room.Description = "You are floating in the blackness of space. Complete darkness surrounds " +
		"you in all directions. There is no escape, there is no hope, just the emptiness. " +
		"You are likely to be eaten by a grue."

	room.Location = location
	room.ZoneId = zoneId

	room.initDbObject(&room)
	return &room
}

func (self *Room) GetType() types.ObjectType {
	return types.RoomType
}

func (self *Room) ToString(players types.PCList, npcs types.NPCList, items types.ItemList, area types.Area) string {
	var str string

	areaStr := ""
	if area != nil {
		areaStr = fmt.Sprintf("%s - ", area.GetName())
	}

	str = fmt.Sprintf("\r\n %v>>> %v%s%s %v<<< %v(%v %v %v)\r\n\r\n %v%s\r\n\r\n",
		types.ColorWhite, types.ColorBlue,
		areaStr, self.GetTitle(),
		types.ColorWhite, types.ColorBlue,
		self.GetLocation().X, self.GetLocation().Y, self.GetLocation().Z,
		types.ColorWhite,
		self.GetDescription())

	extraNewLine := ""

	if len(players) > 0 {
		str = fmt.Sprintf("%s %sAlso here:", str, types.ColorBlue)

		names := make([]string, len(players))
		for i, char := range players {
			names[i] = types.Colorize(types.ColorWhite, char.GetName())
		}
		str = fmt.Sprintf("%s %s \r\n", str, strings.Join(names, types.Colorize(types.ColorBlue, ", ")))

		extraNewLine = "\r\n"
	}

	if len(npcs) > 0 {
		str = fmt.Sprintf("%s %s", str, types.Colorize(types.ColorBlue, "NPCs: "))

		names := make([]string, len(npcs))
		for i, npc := range npcs {
			names[i] = types.Colorize(types.ColorWhite, npc.GetName())
		}
		str = fmt.Sprintf("%s %s \r\n", str, strings.Join(names, types.Colorize(types.ColorBlue, ", ")))

		extraNewLine = "\r\n"
	}

	if len(items) > 0 {
		itemMap := make(map[string]int)
		var nameList []string

		for _, item := range items {
			if item == nil {
				continue
			}

			_, found := itemMap[item.GetName()]
			if !found {
				nameList = append(nameList, item.GetName())
			}
			itemMap[item.GetName()]++
		}

		sort.Strings(nameList)

		str = str + " " + types.Colorize(types.ColorBlue, "Items: ")

		var names []string
		for _, name := range nameList {
			if itemMap[name] > 1 {
				name = fmt.Sprintf("%s x%v", name, itemMap[name])
			}
			names = append(names, types.Colorize(types.ColorWhite, name))
		}
		str = str + strings.Join(names, types.Colorize(types.ColorBlue, ", ")) + "\r\n"

		extraNewLine = "\r\n"
	}

	str = str + extraNewLine + " " + types.Colorize(types.ColorBlue, "Exits: ")

	var exitList []string
	for _, direction := range self.GetExits() {
		exitList = append(exitList, utils.DirectionToExitString(direction))
	}

	if len(exitList) == 0 {
		str = str + types.Colorize(types.ColorWhite, "None")
	} else {
		str = str + strings.Join(exitList, " ")
	}

	if len(self.GetLinks()) > 0 {
		str = fmt.Sprintf("%s\r\n\r\n %s %s",
			str,
			types.Colorize(types.ColorBlue, "Other exits:"),
			types.Colorize(types.ColorWhite, strings.Join(self.LinkNames(), ", ")),
		)
	}

	str = str + "\r\n"

	return str
}

func (self *Room) HasExit(dir types.Direction) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	_, found := self.Exits[dir]
	return found
}

func (self *Room) SetExitEnabled(dir types.Direction, enabled bool) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Exits == nil {
		self.Exits = map[types.Direction]*Exit{}
	}

	if enabled {
		self.Exits[dir] = &Exit{}
	} else {
		delete(self.Exits, dir)
	}

	self.modified()
}

func (self *Room) AddItem(id types.Id) {
	if !self.HasItem(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		if self.Items == nil {
			self.Items = map[string]bool{}
		}

		self.Items[id.Hex()] = true
		self.modified()
	}
}

func (self *Room) RemoveItem(id types.Id) {
	if self.HasItem(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		delete(self.Items, id.Hex())
		self.modified()
	}
}

func (self *Room) SetLink(name string, roomId types.Id) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Links == nil {
		self.Links = map[string]types.Id{}
	}

	self.Links[name] = roomId

	self.modified()
}

func (self *Room) RemoveLink(name string) {
	self.WriteLock()
	defer self.WriteUnlock()

	delete(self.Links, name)
	self.modified()
}

func (self *Room) GetLinks() map[string]types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Links
}

func (self *Room) LinkNames() []string {
	names := make([]string, len(self.GetLinks()))

	i := 0
	for name := range self.Links {
		names[i] = name
		i++
	}
	return names
}

func (self *Room) HasItem(id types.Id) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	_, found := self.Items[id.Hex()]
	return found
}

func (self *Room) GetItems() []types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	ids := make([]types.Id, len(self.Items))
	i := 0
	for id := range self.Items {
		ids[i] = bson.ObjectIdHex(id)
		i++
	}

	return ids
}

func (self *Room) SetTitle(title string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if title != self.Title {
		self.Title = title
		self.modified()
	}
}

func (self *Room) GetTitle() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Title
}

func (self *Room) SetDescription(description string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Description != description {
		self.Description = description
		self.modified()
	}
}

func (self *Room) GetDescription() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Description
}

func (self *Room) SetLocation(location types.Coordinate) {
	self.WriteLock()
	defer self.WriteUnlock()

	if location != self.Location {
		self.Location = location
		self.modified()
	}
}

func (self *Room) GetLocation() types.Coordinate {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Location
}

func (self *Room) SetZoneId(zoneId types.Id) {
	self.WriteLock()
	defer self.WriteUnlock()

	if zoneId != self.ZoneId {
		self.ZoneId = zoneId
		self.modified()
	}
}

func (self *Room) GetZoneId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.ZoneId
}

func (self *Room) SetAreaId(areaId types.Id) {
	self.WriteLock()
	defer self.WriteUnlock()

	if areaId != self.AreaId {
		self.AreaId = areaId
		self.modified()
	}
}

func (self *Room) GetAreaId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.AreaId
}

func (self *Room) NextLocation(direction types.Direction) types.Coordinate {
	loc := self.GetLocation()
	return loc.Next(direction)
}

func (self *Room) GetExits() []types.Direction {
	self.ReadLock()
	defer self.ReadUnlock()

	exits := make([]types.Direction, len(self.Exits))

	i := 0
	for dir := range self.Exits {
		exits[i] = dir
		i++
	}

	return exits
}

func (self *Room) SetProperty(name, value string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Properties == nil {
		self.Properties = map[string]string{}
	}

	if self.Properties[name] != value {
		self.Properties[name] = value
		self.modified()
	}
}

func (self *Room) GetProperty(name string) string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Properties[name]
}

func (self *Room) GetProperties() map[string]string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Properties
}

func (self *Room) RemoveProperty(key string) {
	self.WriteLock()
	defer self.WriteUnlock()

	delete(self.Properties, key)
	self.modified()
}

func (self *Room) SetLocked(dir types.Direction, locked bool) {
	if self.HasExit(dir) {
		self.WriteLock()
		defer self.WriteUnlock()

		self.Exits[dir].Locked = locked
		self.modified()
	}
}

func (self *Room) IsLocked(dir types.Direction) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	if self.HasExit(dir) {
		return self.Exits[dir].Locked
	}

	return false
}

// vim: nocindent
