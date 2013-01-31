package database

import (
	"fmt"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
	"strings"
)

type Room struct {
	DbObject      `bson:",inline"`
	ZoneId        bson.ObjectId `bson:",omitempty"`
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
	Default       bool
}

func NewRoom(zoneId bson.ObjectId) Room {
	var room Room
	room.Id = bson.NewObjectId()
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

	room.Location = Coordinate{0, 0, 0}

	room.Default = false

	room.ZoneId = zoneId

	return room
}

func (self *Room) ToString(mode PrintMode, colorMode utils.ColorMode, chars []*Character, npcs []*Character, items []Item) string {
	var str string

	if mode == ReadMode {
		str = fmt.Sprintf("\n %v %v %v (%v %v %v)\n\n %v\n\n",
			utils.Colorize(colorMode, utils.ColorWhite, ">>>"),
			utils.Colorize(colorMode, utils.ColorBlue, self.Title),
			utils.Colorize(colorMode, utils.ColorWhite, "<<<"),
			self.Location.X,
			self.Location.Y,
			self.Location.Z,
			utils.Colorize(colorMode, utils.ColorWhite, self.Description))

		extraNewLine := ""

		if len(chars) > 0 {
			str = str + " " + utils.Colorize(colorMode, utils.ColorBlue, "Also here: ")

			var names []string
			for _, char := range chars {
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
			str = str + " " + utils.Colorize(colorMode, utils.ColorBlue, "Items: ")

			var names []string
			for _, item := range items {
				names = append(names, utils.Colorize(colorMode, utils.ColorWhite, item.PrettyName()))
			}
			str = str + strings.Join(names, utils.Colorize(colorMode, utils.ColorBlue, ", ")) + "\n"

			extraNewLine = "\n"
		}

		str = str + extraNewLine + " " + utils.Colorize(colorMode, utils.ColorBlue, "Exits: ")

	} else {
		str = fmt.Sprintf(" [1] %v \n\n [2] %v \n\n [3] Exits: ", self.Title, self.Description)
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
		str = str + "None"
	} else {
		str = str + strings.Join(exitList, " ")
	}

	str = str + "\n"

	return str
}

func (self *Room) HasExit(dir ExitDirection) bool {
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

func (self *Room) SetExitEnabled(dir ExitDirection, enabled bool) {
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
}

func (self *Room) AddItem(item Item) {
	self.Items = append(self.Items, item.GetId())
}

func (self *Room) RemoveItem(item Item) {
	for i, myItemId := range self.Items {
		if myItemId == item.GetId() {
			// TODO: Potential memory leak. See http://code.google.com/p/go-wiki/wiki/SliceTricks
			self.Items = append(self.Items[:i], self.Items[i+1:]...)
			return
		}
	}
}

// vim: nocindent
