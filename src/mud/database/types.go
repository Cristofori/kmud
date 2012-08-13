package database

import (
	"strings"
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

func (self *Room) ToString() string {
	str := "\n" + self.Title + "\n\n" + self.Description + "\n"

	var exitList []string
    if len(self.Exits) > 0 {
        str = str + "\n"
    }

	for _, exit := range self.Exits {
		exitList = append(exitList, exit.Text)
	}

	str = str + strings.Join(exitList, ", ")
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
