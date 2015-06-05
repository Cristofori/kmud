package database

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"kmud/utils"
	"sort"
	"strings"
)

type Room struct {
	DbObject `bson:",inline"`

	ZoneId        bson.ObjectId
	AreaId        bson.ObjectId `bson:",omitempty"`
	Title         string
	Description   string
	Items         []bson.ObjectId
	Location      Coordinate
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

func NewRoom(zoneId bson.ObjectId, location Coordinate) *Room {
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

	room.initDbObject()
	commitObject(&room)
	return &room
}

func (self *Room) GetType() ObjectType {
	return RoomType
}

func (self *Room) ToString(players []*PlayerChar, npcs []*NonPlayerChar, items []*Item, area *Area) string {
	var str string

	areaStr := ""
	if area != nil {
		areaStr = fmt.Sprintf(" - %s", area.GetName())
	}

	str = fmt.Sprintf("\r\n %v>>> %v%s%s %v<<< %v(%v %v %v)\r\n\r\n %v%s\r\n\r\n",
		utils.ColorWhite, utils.ColorBlue,
		self.GetTitle(), areaStr,
		utils.ColorWhite, utils.ColorBlue,
		self.GetLocation().X, self.GetLocation().Y, self.GetLocation().Z,
		utils.ColorWhite,
		self.GetDescription())

	extraNewLine := ""

	if len(players) > 0 {
		str = fmt.Sprintf("%s %sAlso here:", str, utils.ColorBlue)

		var names []string
		for _, char := range players {
			names = append(names, utils.Colorize(utils.ColorWhite, char.GetName()))
		}
		str = str + strings.Join(names, utils.Colorize(utils.ColorBlue, ", ")) + "\n"

		extraNewLine = "\r\n"
	}

	if len(npcs) > 0 {
		str = str + " " + utils.Colorize(utils.ColorBlue, "NPCs: ")

		var names []string
		for _, npc := range npcs {
			names = append(names, utils.Colorize(utils.ColorWhite, npc.GetName()))
		}
		str = str + strings.Join(names, utils.Colorize(utils.ColorBlue, ", ")) + "\r\n"

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

		str = str + " " + utils.Colorize(utils.ColorBlue, "Items: ")

		var names []string
		for _, name := range nameList {
			if itemMap[name] > 1 {
				name = fmt.Sprintf("%s x%v", name, itemMap[name])
			}
			names = append(names, utils.Colorize(utils.ColorWhite, name))
		}
		str = str + strings.Join(names, utils.Colorize(utils.ColorBlue, ", ")) + "\r\n"

		extraNewLine = "\r\n"
	}

	str = str + extraNewLine + " " + utils.Colorize(utils.ColorBlue, "Exits: ")

	var exitList []string
	for _, direction := range self.GetExits() {
		exitList = append(exitList, directionToExitString(direction))
	}

	if len(exitList) == 0 {
		str = str + utils.Colorize(utils.ColorWhite, "None")
	} else {
		str = str + strings.Join(exitList, " ")
	}

	str = str + "\r\n"

	return str
}

func (self *Room) HasExit(dir Direction) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	switch dir {
	case DirectionNorth:
		return self.ExitNorth
	case DirectionNorthEast:
		return self.ExitNorthEast
	case DirectionEast:
		return self.ExitEast
	case DirectionSouthEast:
		return self.ExitSouthEast
	case DirectionSouth:
		return self.ExitSouth
	case DirectionSouthWest:
		return self.ExitSouthWest
	case DirectionWest:
		return self.ExitWest
	case DirectionNorthWest:
		return self.ExitNorthWest
	case DirectionUp:
		return self.ExitUp
	case DirectionDown:
		return self.ExitDown
	}

	panic("Unexpected code path")
}

func (self *Room) SetExitEnabled(dir Direction, enabled bool) {
	self.WriteLock()
	defer self.WriteUnlock()

	switch dir {
	case DirectionNorth:
		self.ExitNorth = enabled
	case DirectionNorthEast:
		self.ExitNorthEast = enabled
	case DirectionEast:
		self.ExitEast = enabled
	case DirectionSouthEast:
		self.ExitSouthEast = enabled
	case DirectionSouth:
		self.ExitSouth = enabled
	case DirectionSouthWest:
		self.ExitSouthWest = enabled
	case DirectionWest:
		self.ExitWest = enabled
	case DirectionNorthWest:
		self.ExitNorthWest = enabled
	case DirectionUp:
		self.ExitUp = enabled
	case DirectionDown:
		self.ExitDown = enabled
	}

	objectModified(self)
}

func (self *Room) AddItem(item *Item) {
	if !self.HasItem(item) {
		self.WriteLock()
		defer self.WriteUnlock()

		self.Items = append(self.Items, item.GetId())
		objectModified(self)
	}
}

func (self *Room) RemoveItem(item *Item) {
	if self.HasItem(item) {
		self.WriteLock()
		defer self.WriteUnlock()

		for i, itemId := range self.Items {
			if itemId == item.GetId() {
				// TODO: Potential memory leak. See http://code.google.com/p/go-wiki/wiki/SliceTricks
				self.Items = append(self.Items[:i], self.Items[i+1:]...)
				break
			}
		}

		objectModified(self)
	}
}

func (self *Room) HasItem(item *Item) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	for _, itemId := range self.Items {
		if itemId == item.GetId() {
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

func (self *Room) SetLocation(location Coordinate) {
	self.WriteLock()
	defer self.WriteUnlock()

	if location != self.Location {
		self.Location = location
		objectModified(self)
	}
}

func (self *Room) GetLocation() Coordinate {
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

func (self *Room) NextLocation(direction Direction) Coordinate {
	loc := self.GetLocation()
	return loc.Next(direction)
}

func (self *Room) GetExits() []Direction {
	var exits []Direction

	appendIfExists := func(direction Direction) {
		if self.HasExit(direction) {
			exits = append(exits, direction)
		}
	}

	appendIfExists(DirectionNorth)
	appendIfExists(DirectionNorthEast)
	appendIfExists(DirectionEast)
	appendIfExists(DirectionSouthEast)
	appendIfExists(DirectionSouth)
	appendIfExists(DirectionSouthWest)
	appendIfExists(DirectionWest)
	appendIfExists(DirectionNorthWest)
	appendIfExists(DirectionUp)
	appendIfExists(DirectionDown)

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
