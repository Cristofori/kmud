package database

import (
    "strings"
)

type Exit struct {
    Text string
    DestRoomId string
    Shortcut string
}

type Room struct {
    Title string
    Description string
    Exits []Exit
}

func (self *Room) ToString() string {
    str := self.Title + "\n\n" + self.Description + "\n\n"

    var exitList []string
    for _, exit := range self.Exits {
        exitList = append(exitList, exit.Text)
    }

    str = str + strings.Join(exitList, ", ")
    return str
}

// vim: nocindent
