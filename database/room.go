package database

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
	"gopkg.in/mgo.v2/bson"
)

type Room struct {
	DbObject `bson:",inline"`

	ZoneId        bson.ObjectId
	AreaId        bson.ObjectId `bson:",omitempty"`
	Title         string
	Description   string
	Items         []bson.ObjectId
	Location      types.Coordinate
	ExitNorth     bool
	ExitNorthEast bool
	ExitEast      bool
	ExitSouthEast bool
	ExitSouth     bool
	ExitSouthWest bool
	ExitWest      bool
	ExitNorthWest bool
	ExitUp        bool
	ExitDown      bool

	Properties map[string]string
}

func NewRoom(zoneId bson.ObjectId, location types.Coordinate) *Room {
	var room Room

	room.Title = "The Void"
	room.Description = "You are floating in the blackness of space. Complete darkness surrounds " +
		"you in all directions. There is no escape, there is no hope, just the emptiness. " +
		"You are likely to be eaten by a grue."

	room.ExitNorth = false
	room.ExitNorthEast = false
	room.ExitEast = false
	room.ExitSouthEast = false
	room.ExitSouth = false
	room.ExitSouthWest = false
	room.ExitWest = false
	room.ExitNorthWest = false
	room.ExitUp = false
	room.ExitDown = false

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
		areaStr = fmt.Sprintf(" - %s", area.GetName())
	}

	str = fmt.Sprintf("\r\n %v>>> %v%s%s %v<<< %v(%v %v %v)\r\n\r\n %v%s\r\n\r\n",
		types.ColorWhite, types.ColorBlue,
		self.GetTitle(), areaStr,
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

	str = str + "\r\n"

	return str
}

func (self *Room) HasExit(dir types.Direction) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	switch dir {
	case types.DirectionNorth:
		return self.ExitNorth
	case types.DirectionNorthEast:
		return self.ExitNorthEast
	case types.DirectionEast:
		return self.ExitEast
	case types.DirectionSouthEast:
		return self.ExitSouthEast
	case types.DirectionSouth:
		return self.ExitSouth
	case types.DirectionSouthWest:
		return self.ExitSouthWest
	case types.DirectionWest:
		return self.ExitWest
	case types.DirectionNorthWest:
		return self.ExitNorthWest
	case types.DirectionUp:
		return self.ExitUp
	case types.DirectionDown:
		return self.ExitDown
	}

	panic("Unexpected code path")
}

func (self *Room) SetExitEnabled(dir types.Direction, enabled bool) {
	self.WriteLock()
	defer self.WriteUnlock()

	switch dir {
	case types.DirectionNorth:
		self.ExitNorth = enabled
	case types.DirectionNorthEast:
		self.ExitNorthEast = enabled
	case types.DirectionEast:
		self.ExitEast = enabled
	case types.DirectionSouthEast:
		self.ExitSouthEast = enabled
	case types.DirectionSouth:
		self.ExitSouth = enabled
	case types.DirectionSouthWest:
		self.ExitSouthWest = enabled
	case types.DirectionWest:
		self.ExitWest = enabled
	case types.DirectionNorthWest:
		self.ExitNorthWest = enabled
	case types.DirectionUp:
		self.ExitUp = enabled
	case types.DirectionDown:
		self.ExitDown = enabled
	}

	objectModified(self)
}

func (self *Room) AddItem(id bson.ObjectId) {
	if !self.HasItem(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		self.Items = append(self.Items, id)
		objectModified(self)
	}
}

func (self *Room) RemoveItem(id bson.ObjectId) {
	if self.HasItem(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		for i, itemId := range self.Items {
			if itemId == id {
				// TODO: Potential memory leak. See http://code.google.com/p/go-wiki/wiki/SliceTricks
				self.Items = append(self.Items[:i], self.Items[i+1:]...)
				break
			}
		}

		objectModified(self)
	}
}

func (self *Room) HasItem(id bson.ObjectId) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	for _, itemId := range self.Items {
		if itemId == id {
			return true
		}
	}

	return false
}

func (self *Room) GetItemIds() []bson.ObjectId {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Items
}

func (self *Room) SetTitle(title string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if title != self.Title {
		self.Title = title
		objectModified(self)
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
		objectModified(self)
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
		objectModified(self)
	}
}

func (self *Room) GetLocation() types.Coordinate {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Location
}

func (self *Room) SetZoneId(zoneId bson.ObjectId) {
	self.WriteLock()
	defer self.WriteUnlock()

	if zoneId != self.ZoneId {
		self.ZoneId = zoneId
		objectModified(self)
	}
}

func (self *Room) GetZoneId() bson.ObjectId {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.ZoneId
}

func (self *Room) SetAreaId(areaId bson.ObjectId) {
	self.WriteLock()
	defer self.WriteUnlock()

	if areaId != self.AreaId {
		self.AreaId = areaId
		objectModified(self)
	}
}

func (self *Room) GetAreaId() bson.ObjectId {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.AreaId
}

func (self *Room) NextLocation(direction types.Direction) types.Coordinate {
	loc := self.GetLocation()
	return loc.Next(direction)
}

func (self *Room) GetExits() []types.Direction {
	var exits []types.Direction

	appendIfExists := func(direction types.Direction) {
		if self.HasExit(direction) {
			exits = append(exits, direction)
		}
	}

	appendIfExists(types.DirectionNorth)
	appendIfExists(types.DirectionNorthEast)
	appendIfExists(types.DirectionEast)
	appendIfExists(types.DirectionSouthEast)
	appendIfExists(types.DirectionSouth)
	appendIfExists(types.DirectionSouthWest)
	appendIfExists(types.DirectionWest)
	appendIfExists(types.DirectionNorthWest)
	appendIfExists(types.DirectionUp)
	appendIfExists(types.DirectionDown)

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
		objectModified(self)
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
	objectModified(self)
}

type Rooms []*Room

// vim: nocindent
