package database

import (
	"fmt"
	"kmud/utils"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

type Direction int

const (
	DirectionNorth     Direction = iota
	DirectionNorthEast Direction = iota
	DirectionEast      Direction = iota
	DirectionSouthEast Direction = iota
	DirectionSouth     Direction = iota
	DirectionSouthWest Direction = iota
	DirectionWest      Direction = iota
	DirectionNorthWest Direction = iota
	DirectionUp        Direction = iota
	DirectionDown      Direction = iota
	DirectionNone      Direction = iota
)

type Identifiable interface {
	GetId() bson.ObjectId
	GetType() objectType
	ReadLock()
	ReadUnlock()
	Destroy()
	IsDestroyed() bool
}

type objectType int

const (
	CharType objectType = iota
	UserType objectType = iota
	ZoneType objectType = iota
	AreaType objectType = iota
	RoomType objectType = iota
	ItemType objectType = iota
)

type Coordinate struct {
	X int
	Y int
	Z int
}

func directionToExitString(direction Direction) string {
	letterColor := utils.ColorBlue
	bracketColor := utils.ColorDarkBlue
	textColor := utils.ColorWhite

	colorize := func(letters string, text string) string {
		return fmt.Sprintf("%s%s%s%s",
			utils.Colorize(bracketColor, "["),
			utils.Colorize(letterColor, letters),
			utils.Colorize(bracketColor, "]"),
			utils.Colorize(textColor, text))
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
		return utils.Colorize(utils.ColorWhite, "None")
	}

	panic("Unexpected code path")
}

func (self *Coordinate) Next(direction Direction) Coordinate {
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

func StringToDirection(str string) Direction {
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

func DirectionToString(dir Direction) string {
	switch dir {
	case DirectionNorth:
		return "North"
	case DirectionNorthEast:
		return "NorthEast"
	case DirectionEast:
		return "East"
	case DirectionSouthEast:
		return "SouthEast"
	case DirectionSouth:
		return "South"
	case DirectionSouthWest:
		return "SouthWest"
	case DirectionWest:
		return "West"
	case DirectionNorthWest:
		return "NorthWest"
	case DirectionUp:
		return "Up"
	case DirectionDown:
		return "Down"
	case DirectionNone:
		return "None"
	}

	panic("Unexpected code path")
}

func (self Direction) Opposite() Direction {
	switch self {
	case DirectionNorth:
		return DirectionSouth
	case DirectionNorthEast:
		return DirectionSouthWest
	case DirectionEast:
		return DirectionWest
	case DirectionSouthEast:
		return DirectionNorthWest
	case DirectionSouth:
		return DirectionNorth
	case DirectionSouthWest:
		return DirectionNorthEast
	case DirectionWest:
		return DirectionEast
	case DirectionNorthWest:
		return DirectionSouthEast
	case DirectionUp:
		return DirectionDown
	case DirectionDown:
		return DirectionUp
	}

	return DirectionNone
}

// vim: nocindent
