package utils

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"net"
)

type action struct {
	key  string
	text string
	data bson.ObjectId
}

type Menu struct {
	Actions []action
	Title   string
}

func NewMenu(text string) Menu {
	var menu Menu
	// menu.Actions = map[string]string{}
	menu.Title = text
	return menu
}

func (self *Menu) AddAction(key string, text string) {
	self.Actions = append(self.Actions, action{key: key, text: text})
}

func (self *Menu) AddActionData(key string, text string, data bson.ObjectId) {
	self.Actions = append(self.Actions, action{key: key, text: text, data: data})
}

func (self *Menu) Exec(conn net.Conn) (string, bson.ObjectId) {

	border := "-=-=-"
	for {
		WriteLine(conn, fmt.Sprintf("%s %s %s", border, self.Title, border))

		for _, action := range self.Actions {
			WriteLine(conn, fmt.Sprintf("  %s", action.text))
		}

		input := GetUserInput(conn, "> ")

		if input == "" {
			return "", ""
		}

		for _, action := range self.Actions {
			if action.key == input {
				return input, action.data
			}
		}
	}

	panic("Unexpected code path")
	return "", ""
}

// vim: nocindent
