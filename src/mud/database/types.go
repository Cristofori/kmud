package database

import (
	"fmt"
	"strings"
)

type Exit struct {
	Id         string
	Text       string
	DestRoomId string
	Direction  ExitDirection
}

type Room struct {
	Id          string
	Title       string
	Description string
	Exits       []Exit
}

type ExitDirection int

const (
    None ExitDirection = iota
    North ExitDirection = iota
    East ExitDirection = iota
    South ExitDirection = iota
    West ExitDirection = iota
    Up ExitDirection = iota
    Down ExitDirection = iota
)

type PrintMode int

const (
	ReadMode PrintMode = iota
	EditMode PrintMode = iota
)

func (self *Room) ToString(mode PrintMode) string {

	var str string
	if mode == ReadMode {
		str = fmt.Sprintf("\n >>> %v <<<\n\n %v \n\n Exits: ", self.Title, self.Description)
	} else {
		str = fmt.Sprintf("\n [1] %v \n\n [2] %v \n\n [3] Exits: ", self.Title, self.Description)
	}

	var exitList []string
	if len(self.Exits) > 0 {
		for _, exit := range self.Exits {
			exitList = append(exitList, exit.Text)
		}
		str = str + strings.Join(exitList, ", ")
	} else {
		str = str + "None"
	}
	str = str + "\n"

	return str
}

func (self *Room) GetExit(dir ExitDirection) Exit {
	for _, exit := range self.Exits {
		if exit.Direction == dir {
			return exit
		}
	}

	var exit Exit
    exit.Direction = None
	return exit
}

func (self *Room) ExitId(dir ExitDirection) string {
	return self.GetExit(dir).Id
}

func (self *Room) HasExit(dir ExitDirection) bool {
	return self.GetExit(dir).Id != ""
}

func StringToDirection( str string ) ExitDirection {
    dirStr := strings.ToLower(str)
    switch dirStr {
        case "n": return North
        case "s": return South
        case "e": return East
        case "w": return West
        case "u": return Up
        case "d": return Down
    }

    return None
}

// vim: nocindent
