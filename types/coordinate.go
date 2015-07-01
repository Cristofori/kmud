package types

import "strings"

type Coordinate struct {
	X int
	Y int
	Z int
}

type Direction string

const (
	DirectionNorth     Direction = "North"
	DirectionNorthEast Direction = "NorthEast"
	DirectionEast      Direction = "East"
	DirectionSouthEast Direction = "SouthEast"
	DirectionSouth     Direction = "South"
	DirectionSouthWest Direction = "SouthWest"
	DirectionWest      Direction = "West"
	DirectionNorthWest Direction = "NorthWest"
	DirectionUp        Direction = "Up"
	DirectionDown      Direction = "Down"
	DirectionNone      Direction = "None"
)

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

func (dir Direction) ToString() string {
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
