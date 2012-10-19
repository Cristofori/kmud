package database

import (
	"fmt"
	"strings"
)

type Room struct {
	Id          string
	Title       string
	Description string
	Exits       map[ExitDirection]bool
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
		str = fmt.Sprintf("\n >>> %v <<<\n\n %v \n\n Exits: ", self.Title, self.Description)
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
	return self.Exits[dir] == true
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
