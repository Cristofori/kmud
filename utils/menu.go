package utils

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"net"
	"regexp"
	"strconv"
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

func (self *Menu) AddActionData(key int, text string, data bson.ObjectId) {
	keyStr := strconv.Itoa(key)
	self.Actions = append(self.Actions, action{key: keyStr, text: text, data: data})
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

func (self *Menu) Exec(conn net.Conn, cm ColorMode) (string, bson.ObjectId) {
	for {
		self.Print(conn, cm)
		input := GetUserInput(conn, Colorize(cm, ColorWhite, self.Prompt))

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

func (self *Menu) Print(conn net.Conn, cm ColorMode) {
	border := Colorize(cm, ColorWhite, "-=-=-")
	title := Colorize(cm, ColorBlue, self.Title)
	WriteLine(conn, fmt.Sprintf("%s %s %s", border, title, border))

	for _, action := range self.Actions {
		regex := regexp.MustCompile("^\\[([^\\]]*)\\](.*)")
		matches := regex.FindStringSubmatch(action.text)

		actionText := action.text

		if len(matches) == 3 {
			actionText = Colorize(cm, ColorDarkBlue, "[") +
				Colorize(cm, ColorBlue, matches[1]) +
				Colorize(cm, ColorDarkBlue, "]") +
				Colorize(cm, ColorWhite, matches[2])
		}

		WriteLine(conn, "  "+actionText)
	}
}

// vim: nocindent
