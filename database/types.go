package database

import (
	"fmt"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
	"strings"
	"sync"
)

type Coordinate struct {
	X int
	Y int
	Z int
}

// All database types should meet this interface
type Identifiable interface {
	GetId() bson.ObjectId
	GetType() objectType
}

type Nameable interface {
	PrettyName() string
}

type objectType int

const (
	characterType objectType = iota
	roomType      objectType = iota
	userType      objectType = iota
	zoneType      objectType = iota
)

const (
	dbObjectName string = "name"
)

type DbObject struct {
	Id      bson.ObjectId `bson:"_id"`
	objType objectType
	Name    string `bson:",omitempty"`
	Fields  map[string]interface{}
	mutex   sync.Mutex
}

type ObjectField string

type Item struct {
	DbObject `bson:",inline"`
}

func NewItem(name string) Item {
	var item Item
	item.Id = bson.NewObjectId()
	item.Name = name

	return item
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

func directionToExitString(colorMode utils.ColorMode, direction ExitDirection) string {

	letterColor := utils.ColorBlue
	bracketColor := utils.ColorDarkBlue
	textColor := utils.ColorWhite

	colorize := func(letters string, text string) string {
		return fmt.Sprintf("%s%s%s%s",
			utils.Colorize(colorMode, bracketColor, "["),
			utils.Colorize(colorMode, letterColor, letters),
			utils.Colorize(colorMode, bracketColor, "]"),
			utils.Colorize(colorMode, textColor, text))
	}

	switch direction {
	case DirectionNorth:
		return colorize("N", "orth")
	case DirectionNorthEast:
		return colorize("NE", "North East")
	case DirectionEast:
		return colorize("E", "ast")
	case DirectionSouthEast:
		return colorize("SE", "South East")
	case DirectionSouth:
		return colorize("S", "outh")
	case DirectionSouthWest:
		return colorize("SW", "South West")
	case DirectionWest:
		return colorize("W", "est")
	case DirectionNorthWest:
		return colorize("NW", "North West")
	case DirectionUp:
		return colorize("U", "p")
	case DirectionDown:
		return colorize("D", "own")
	case DirectionNone:
		return utils.Colorize(colorMode, utils.ColorWhite, "None")
	}

	panic("Unexpected code path")
}

func (self *DbObject) initDbObject(objType objectType) {
	self.Id = bson.NewObjectId()
	self.objType = objType
	self.Fields = map[string]interface{}{}

	commitObject(getCollectionFromType(objType), *self)
}

func (self DbObject) GetId() bson.ObjectId {
	return self.Id
}

func (self DbObject) GetType() objectType {
	return self.objType
}

func (self DbObject) PrettyName() string {
	return utils.FormatName(self.Name)
}

func (self *DbObject) setField(key string, value interface{}) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.Fields[key] = value
	updateObject(*self, "fields."+key, value)
}

func (self *DbObject) getField(key string) interface{} {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	return self.Fields[key]
}

func (self *DbObject) hasField(key string) bool {
	_, found := self.Fields[key]
	return found
}

func (self *DbObject) SetName(name string) {
	self.Name = name
	updateObject(*self, dbObjectName, name)
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
