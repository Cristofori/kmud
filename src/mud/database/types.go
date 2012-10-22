package database

import (
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
}

func newUser(name string) User {
	var user User
	user.Id = bson.NewObjectId()
	user.Name = name
	return user
}

type Character struct {
	Id     bson.ObjectId `bson:"_id"`
	Name   string
	RoomId bson.ObjectId `bson:"roomid"`
}

func newCharacter(name string) Character {
	var character Character
	character.Id = bson.NewObjectId()
	character.Name = name
	character.RoomId = ""
	return character
}

type Room struct {
	Id          bson.ObjectId `bson:"_id"`
	Title       string
	Description string
	Location    Coordinate
	ExitNorth   bool
	ExitEast    bool
	ExitSouth   bool
	ExitWest    bool
	ExitUp      bool
	ExitDown    bool
	Default     bool
}

func newRoom() Room {
	var room Room
	room.Id = bson.NewObjectId()
	room.Title = "The Void"
	room.Description = "You are floating in the blackness of space. Complete darkness surrounds " +
		"you in all directions. There is no escape, there is no hope, just the emptiness. " +
		"You are likely to be eaten by a grue."

	room.ExitNorth = false
	room.ExitEast = false
	room.ExitSouth = false
	room.ExitWest = false
	room.ExitUp = false
	room.ExitDown = false

	room.Location = Coordinate{0, 0, 0}

	room.Default = false

	return room
}

type ExitDirection int

const (
	DirectionNone  ExitDirection = iota
	DirectionNorth ExitDirection = iota
	DirectionEast  ExitDirection = iota
	DirectionSouth ExitDirection = iota
	DirectionWest  ExitDirection = iota
	DirectionUp    ExitDirection = iota
	DirectionDown  ExitDirection = iota
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
	case DirectionEast:
		return "[E]ast"
	case DirectionSouth:
		return "[S]outh"
	case DirectionWest:
		return "[W]est"
	case DirectionUp:
		return "[U]p"
	case DirectionDown:
		return "[D]own"
	case DirectionNone:
		return "None"
	}

	panic("Unexpected code path")
}

func (self *Room) ToString(mode PrintMode) string {

	var str string
	if mode == ReadMode {
		str = fmt.Sprintf("\n >>> %v <<< (%v %v %v)\n\n %v \n\n Exits: ",
			self.Title,
			self.Location.X,
			self.Location.Y,
			self.Location.Z,
			self.Description)
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
	appendIfExists(DirectionEast)
	appendIfExists(DirectionSouth)
	appendIfExists(DirectionWest)
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
	case DirectionEast:
		return self.ExitEast
	case DirectionSouth:
		return self.ExitSouth
	case DirectionWest:
		return self.ExitWest
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
	case DirectionEast:
		self.ExitEast = enabled
	case DirectionSouth:
		self.ExitSouth = enabled
	case DirectionWest:
		self.ExitWest = enabled
	case DirectionUp:
		self.ExitUp = enabled
	case DirectionDown:
		self.ExitDown = enabled
	}
}

func (self *Character) PrettyName() string {
	return utils.FormatName(self.Name)
}

func (self *User) PrettyName() string {
	return utils.FormatName(self.Name)
}

func StringToDirection(str string) ExitDirection {
	dirStr := strings.ToLower(str)
	switch dirStr {
	case "n":
		return DirectionNorth
	case "s":
		return DirectionSouth
	case "e":
		return DirectionEast
	case "w":
		return DirectionWest
	case "u":
		return DirectionUp
	case "d":
		return DirectionDown
	}

	return DirectionNone
}

// vim: nocindent
