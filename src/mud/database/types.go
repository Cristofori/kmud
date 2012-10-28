package database

import (
	"container/list"
	"fmt"
	"labix.org/v2/mgo/bson"
	"mud/utils"
	"strings"
)

type Coordinate struct {
	X int
	Y int
	Z int
}

type User struct {
	Id           bson.ObjectId `bson:"_id,omitempty"`
	Name         string
	CharacterIds []bson.ObjectId `bson:"characterids,omitempty"`
	online       bool
}

func NewUser(name string) User {
	var user User
	user.Id = bson.NewObjectId()
	user.Name = name
	return user
}

type Character struct {
	Id     bson.ObjectId `bson:"_id"`
	Name   string
	RoomId bson.ObjectId `bson:"roomid"`
	online bool
}

func NewCharacter(name string) Character {
	var character Character
	character.Id = bson.NewObjectId()
	character.Name = name
	character.RoomId = ""
	character.online = false
	return character
}

type Room struct {
	Id            bson.ObjectId `bson:"_id"`
	Title         string
	Description   string
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

func NewRoom() Room {
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

	return room
}

type ExitDirection int

const (
	DirectionNone      ExitDirection = iota
	DirectionNorth     ExitDirection = iota
	DirectionNorthEast ExitDirection = iota
	DirectionEast      ExitDirection = iota
	DirectionSouthEast ExitDirection = iota
	DirectionSouth     ExitDirection = iota
	DirectionSouthWest ExitDirection = iota
	DirectionWest      ExitDirection = iota
	DirectionNorthWest ExitDirection = iota
	DirectionUp        ExitDirection = iota
	DirectionDown      ExitDirection = iota
)

type PrintMode int

const (
	ReadMode PrintMode = iota
	EditMode PrintMode = iota
)

func directionToExitString(direction ExitDirection) string {
	switch direction {
	case DirectionNorth:
		return "[N]orth"
	case DirectionNorthEast:
		return "[NE]North East"
	case DirectionEast:
		return "[E]ast"
	case DirectionSouthEast:
		return "[SE]South East"
	case DirectionSouth:
		return "[S]outh"
	case DirectionSouthWest:
		return "[SW]South West"
	case DirectionWest:
		return "[W]est"
	case DirectionNorthWest:
		return "[NW]North West"
	case DirectionUp:
		return "[U]p"
	case DirectionDown:
		return "[D]own"
	case DirectionNone:
		return "None"
	}

	panic("Unexpected code path")
}

func (self *Room) ToString(mode PrintMode, chars *list.List) string {

	var str string
	if mode == ReadMode {
		str = fmt.Sprintf("\n >>> %v <<< (%v %v %v)\n\n %v",
			utils.Colorize(utils.ColorModeNormal, utils.ColorBlue, self.Title),
			self.Location.X,
			self.Location.Y,
			self.Location.Z,
			self.Description)

		if chars.Len() > 0 {
			str = str + "\n\n Also here: "

			var names []string
			for e := chars.Front(); e != nil; e = e.Next() {
				char := e.Value.(Character)
				names = append(names, char.PrettyName())
			}
			str = str + strings.Join(names, ", ")
		}

		str = str + "\n\n Exits: "

	} else {
		str = fmt.Sprintf("\n [1] %v \n\n [2] %v \n\n [3] Exits: ", self.Title, self.Description)
	}

	var exitList []string

	appendIfExists := func(direction ExitDirection) {
		if self.HasExit(direction) {
			exitList = append(exitList, directionToExitString(direction))
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

func (self *Character) PrettyName() string {
	return utils.FormatName(self.Name)
}

func (self *Character) SetOnline(online bool) {
	self.online = online
}

func (self *Character) Online() bool {
	return self.online
}

func (self *User) PrettyName() string {
	return utils.FormatName(self.Name)
}

func (self *User) SetOnline(online bool) {
	self.online = online
}

func (self *User) Online() bool {
	return self.online
}

func (self *Coordinate) Next(direction ExitDirection) Coordinate {
	newCoord := *self
	switch direction {
	case DirectionNorth:
		newCoord.Y -= 1
	case DirectionNorthEast:
		newCoord.Y -= 1
		newCoord.X += 1
	case DirectionEast:
		newCoord.X += 1
	case DirectionSouthEast:
		newCoord.Y += 1
		newCoord.X += 1
	case DirectionSouth:
		newCoord.Y += 1
	case DirectionSouthWest:
		newCoord.Y += 1
		newCoord.X -= 1
	case DirectionWest:
		newCoord.X -= 1
	case DirectionNorthWest:
		newCoord.Y -= 1
		newCoord.X -= 1
	case DirectionUp:
		newCoord.Z -= 1
	case DirectionDown:
		newCoord.Z += 1
	}
	return newCoord
}

func StringToDirection(str string) ExitDirection {
	dirStr := strings.ToLower(str)
	switch dirStr {
	case "n":
		fallthrough
	case "north":
		return DirectionNorth
	case "ne":
		return DirectionNorthEast
	case "e":
		fallthrough
	case "east":
		return DirectionEast
	case "se":
		return DirectionSouthEast
	case "s":
		fallthrough
	case "south":
		return DirectionSouth
	case "sw":
		return DirectionSouthWest
	case "w":
		fallthrough
	case "west":
		return DirectionWest
	case "nw":
		return DirectionNorthWest
	case "u":
		fallthrough
	case "up":
		return DirectionUp
	case "d":
		fallthrough
	case "down":
		return DirectionDown
	}

	return DirectionNone
}

// vim: nocindent
