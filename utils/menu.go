package utils

import (
	"fmt"
	"io"
	"labix.org/v2/mgo/bson"
	"regexp"
	"strconv"
)

type menuAction struct {
	key  string
	text string
	data bson.ObjectId
}

type Menu struct {
	actions []menuAction
	title   string
	prompt  string
}

func NewMenu(text string) Menu {
	var menu Menu
	menu.title = text
	menu.prompt = "> "
	return menu
}

func (self *Menu) AddAction(key string, text string) {
	self.actions = append(self.actions, menuAction{key: key, text: text})
}

func (self *Menu) AddActionData(key int, text string, data bson.ObjectId) {
	keyStr := strconv.Itoa(key)
	self.actions = append(self.actions, menuAction{key: keyStr, text: text, data: data})
}

func (self *Menu) GetData(choice string) bson.ObjectId {
	for _, action := range self.actions {
		if action.key == choice {
			return action.data
		}
	}

	return ""
}

func (self *Menu) GetPrompt() string {
	return self.prompt
}

func (self *Menu) getAction(key string) menuAction {
	for _, action := range self.actions {
		if action.key == key {
			return action
		}
	}
	return menuAction{}
}

func (self *Menu) HasAction(key string) bool {
	action := self.getAction(key)
	return action.key != ""
}

func (self *Menu) Exec(conn io.ReadWriter, cm ColorMode) (string, bson.ObjectId) {
	for {
		self.Print(conn, cm)
		input := GetUserInput(conn, Colorize(cm, ColorWhite, self.prompt))

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

func (self *Menu) Print(conn io.Writer, cm ColorMode) {
	border := Colorize(cm, ColorWhite, "-=-=-")
	title := Colorize(cm, ColorBlue, self.title)
	WriteLine(conn, fmt.Sprintf("%s %s %s", border, title, border))

	for _, action := range self.actions {
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
