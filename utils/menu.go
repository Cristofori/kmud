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
	Prompt  string
}

func NewMenu(text string) Menu {
	var menu Menu
	menu.Title = text
	menu.Prompt = "> "
	return menu
}

func (self *Menu) AddAction(key string, text string) {
	self.Actions = append(self.Actions, action{key: key, text: text})
}

func (self *Menu) AddActionData(key string, text string, data bson.ObjectId) {
	self.Actions = append(self.Actions, action{key: key, text: text, data: data})
}

func (self *Menu) getAction(key string) action {
	for _, action := range self.Actions {
		if action.key == key {
			return action
		}
	}
	return action{}
}

func (self *Menu) HasAction(key string) bool {
	action := self.getAction(key)
	return action.key != ""
}

func (self *Menu) Exec(conn net.Conn) (string, bson.ObjectId) {
	for {
		self.Print(conn)
		input := GetUserInput(conn, self.Prompt)

		if input == "" {
			return "", ""
		}

		action := self.getAction(input)
		if action.key != "" {
			return action.key, action.data
		}
	}

	panic("Unexpected code path")
	return "", ""
}

func (self *Menu) Print(conn net.Conn) {
	border := "-=-=-"
	WriteLine(conn, fmt.Sprintf("%s %s %s", border, self.Title, border))

	for _, action := range self.Actions {
		WriteLine(conn, fmt.Sprintf("  %s", action.text))
	}
}

// vim: nocindent
