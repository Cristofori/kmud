package database

import (
	"strings"
    "fmt"
)

type Exit struct {
	Id         string
	Text       string
	DestRoomId string
	Shortcut   string
}

type Room struct {
	Title       string
	Description string
	Exits       []Exit
}

type PrintMode int
const (
    ReadMode PrintMode = iota
    EditMode PrintMode = iota
)

func (self *Room) ToString( mode PrintMode ) string {

    var str string
    if mode == ReadMode {
        str = fmt.Sprintf( "\n >>> %v <<<\n\n %v \n\n Exits: ", self.Title, self.Description )
    } else {
        str = fmt.Sprintf( "\n [1] %v \n\n [2] %v \n\n [3] Exits: ", self.Title, self.Description )
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

	return str
}

func (self *Room) GetExit(shortcut string) Exit {
	for _, exit := range self.Exits {
		if exit.Shortcut == shortcut {
			return exit
		}
	}

	var exit Exit
	return exit
}

func (self *Room) ExitId(shortcut string) string {
	return self.GetExit(shortcut).Id
}

func (self *Room) HasExit(shortcut string) bool {
	return self.GetExit(shortcut).Id != ""
}

// vim: nocindent
