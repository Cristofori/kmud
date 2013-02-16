package database

import (
	"fmt"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
	"sort"
	"strings"
)

const (
	roomZoneId        string = "zoneid"
	roomTitle         string = "title"
	roomDescription   string = "description"
	roomItems         string = "items"
	roomLocation      string = "location"
	roomExitNorth     string = "exitnorth"
	roomExitNorthEast string = "exitnortheast"
	roomExitEast      string = "exiteast"
	roomExitSouthEast string = "exitsoutheast"
	roomExitSouth     string = "exitsouth"
	roomExitSouthWest string = "exitsouthwest"
	roomExitWest      string = "exitwest"
	roomExitNorthWest string = "exitnorthwest"
	roomExitUp        string = "exitup"
	roomExitDown      string = "exitdown"
)

type Room struct {
	DbObject `bson:",inline"`
}

func NewRoom(zoneId bson.ObjectId) *Room {
	var room Room
	room.initDbObject("", roomType)

	room.SetTitle("The Void")
	room.SetDescription("You are floating in the blackness of space. Complete darkness surrounds " +
		"you in all directions. There is no escape, there is no hope, just the emptiness. " +
		"You are likely to be eaten by a grue.")

	room.SetExitEnabled(DirectionNorth, false)
	room.SetExitEnabled(DirectionNorthEast, false)
	room.SetExitEnabled(DirectionEast, false)
	room.SetExitEnabled(DirectionSouthEast, false)
	room.SetExitEnabled(DirectionSouth, false)
	room.SetExitEnabled(DirectionSouthWest, false)
	room.SetExitEnabled(DirectionWest, false)
	room.SetExitEnabled(DirectionNorthWest, false)
	room.SetExitEnabled(DirectionUp, false)
	room.SetExitEnabled(DirectionDown, false)

	room.SetLocation(Coordinate{0, 0, 0})
	room.SetZoneId(zoneId)

	return &room
}

func (self *Room) ToString(mode PrintMode, colorMode utils.ColorMode, players []*Character, npcs []*Character, items []*Item) string {
	var str string

	if mode == ReadMode {
		str = fmt.Sprintf("\n %v %v %v (%v %v %v)\n\n %v\n\n",
			utils.Colorize(colorMode, utils.ColorWhite, ">>>"),
			utils.Colorize(colorMode, utils.ColorBlue, self.GetTitle()),
			utils.Colorize(colorMode, utils.ColorWhite, "<<<"),
			self.GetLocation().X,
			self.GetLocation().Y,
			self.GetLocation().Z,
			utils.Colorize(colorMode, utils.ColorWhite, self.GetDescription()))

		extraNewLine := ""

		if len(players) > 0 {
			str = str + " " + utils.Colorize(colorMode, utils.ColorBlue, "Also here: ")

			var names []string
			for _, char := range players {
				names = append(names, utils.Colorize(colorMode, utils.ColorWhite, char.PrettyName()))
			}
			str = str + strings.Join(names, utils.Colorize(colorMode, utils.ColorBlue, ", ")) + "\n"

			extraNewLine = "\n"
		}

		if len(npcs) > 0 {
			str = str + " " + utils.Colorize(colorMode, utils.ColorBlue, "NPCs: ")

			var names []string
			for _, npc := range npcs {
				names = append(names, utils.Colorize(colorMode, utils.ColorWhite, npc.PrettyName()))
			}
			str = str + strings.Join(names, utils.Colorize(colorMode, utils.ColorBlue, ", ")) + "\n"

			extraNewLine = "\n"
		}

		if len(items) > 0 {

			itemMap := make(map[string]int)
			var nameList []string

			for _, item := range items {
				_, found := itemMap[item.PrettyName()]
				if !found {
					nameList = append(nameList, item.PrettyName())
				}
				itemMap[item.PrettyName()]++
			}

			sort.Strings(nameList)

			str = str + " " + utils.Colorize(colorMode, utils.ColorBlue, "Items: ")

			var names []string
			for _, name := range nameList {
				if itemMap[name] > 1 {
					name = fmt.Sprintf("%s x%v", name, itemMap[name])
				}
				names = append(names, utils.Colorize(colorMode, utils.ColorWhite, name))
			}
			str = str + strings.Join(names, utils.Colorize(colorMode, utils.ColorBlue, ", ")) + "\n"

			extraNewLine = "\n"
		}

		str = str + extraNewLine + " " + utils.Colorize(colorMode, utils.ColorBlue, "Exits: ")

	} else {
		str = fmt.Sprintf(" [1] %v \n\n [2] %v \n\n [3] Exits: ", self.GetTitle(), self.GetDescription())
	}

	var exitList []string

	appendIfExists := func(direction ExitDirection) {
		if self.HasExit(direction) {
			exitList = append(exitList, directionToExitString(colorMode, direction))
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

	if len(exitList) == 0 {
		str = str + utils.Colorize(colorMode, utils.ColorWhite, "None")
	} else {
		str = str + strings.Join(exitList, " ")
	}

	str = str + "\n"

	return str
}

func (self *Room) HasExit(dir ExitDirection) bool {
	switch dir {
	case DirectionNorth:
		return self.getField(roomExitNorth).(bool)
	case DirectionNorthEast:
		return self.getField(roomExitNorthEast).(bool)
	case DirectionEast:
		return self.getField(roomExitEast).(bool)
	case DirectionSouthEast:
		return self.getField(roomExitSouthEast).(bool)
	case DirectionSouth:
		return self.getField(roomExitSouth).(bool)
	case DirectionSouthWest:
		return self.getField(roomExitSouthWest).(bool)
	case DirectionWest:
		return self.getField(roomExitWest).(bool)
	case DirectionNorthWest:
		return self.getField(roomExitNorthWest).(bool)
	case DirectionUp:
		return self.getField(roomExitUp).(bool)
	case DirectionDown:
		return self.getField(roomExitDown).(bool)
	}

	panic("Unexpected code path")
}

func (self *Room) SetExitEnabled(dir ExitDirection, enabled bool) {
	switch dir {
	case DirectionNorth:
		self.setField(roomExitNorth, enabled)
	case DirectionNorthEast:
		self.setField(roomExitNorthEast, enabled)
	case DirectionEast:
		self.setField(roomExitEast, enabled)
	case DirectionSouthEast:
		self.setField(roomExitSouthEast, enabled)
	case DirectionSouth:
		self.setField(roomExitSouth, enabled)
	case DirectionSouthWest:
		self.setField(roomExitSouthWest, enabled)
	case DirectionWest:
		self.setField(roomExitWest, enabled)
	case DirectionNorthWest:
		self.setField(roomExitNorthWest, enabled)
	case DirectionUp:
		self.setField(roomExitUp, enabled)
	case DirectionDown:
		self.setField(roomExitDown, enabled)
	}
}

func (self *Room) AddItem(item *Item) {
	itemIds := self.GetItemIds()
	itemIds = append(itemIds, item.GetId())
	self.setField(roomItems, itemIds)
}

func (self *Room) RemoveItem(item *Item) {
	itemIds := self.GetItemIds()
	for i, itemId := range itemIds {
		if itemId == item.GetId() {
			// TODO: Potential memory leak. See http://code.google.com/p/go-wiki/wiki/SliceTricks
			itemIds = append(itemIds[:i], itemIds[i+1:]...)
			self.setField(roomItems, itemIds)
			return
		}
	}
}

func (self *Room) SetTitle(title string) {
	self.setField(roomTitle, title)
}

func (self *Room) GetTitle() string {
	return self.getField(roomTitle).(string)
}

func (self *Room) SetDescription(description string) {
	self.setField(roomDescription, description)
}

func (self *Room) GetDescription() string {
	return self.getField(roomDescription).(string)
}

func (self *Room) SetLocation(location Coordinate) {
	self.setField(roomLocation, location)
}

func (self *Room) GetLocation() Coordinate {
	return self.getField(roomLocation).(Coordinate)
}

func (self *Room) SetZoneId(zoneId bson.ObjectId) {
	self.setField(roomZoneId, zoneId)
}

func (self *Room) GetZoneId() bson.ObjectId {
	return self.getField(roomZoneId).(bson.ObjectId)
}

func (self *Room) GetItemIds() []bson.ObjectId {
	if self.hasField(roomItems) {
		return self.getField(roomItems).([]bson.ObjectId)
	}

	return []bson.ObjectId{}
}

func (self *Room) NextLocation(direction ExitDirection) Coordinate {
	loc := self.GetLocation()
	return loc.Next(direction)
}

// vim: nocindent
