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
		input := GetUserInput(conn, Colorize(ColorWhite, self.prompt), cm)

		if input == "" {
			return "", ""
		}

		action := self.getAction(input)
		if action.key != "" {
			return action.key, action.data
		}
	}
}

func (self *Menu) Print(conn io.Writer, cm ColorMode) {
	border := Colorize(ColorWhite, "-=-=-")
	title := Colorize(ColorBlue, self.title)
	WriteLine(conn, fmt.Sprintf("%s %s %s", border, title, border), cm)

	for _, action := range self.actions {
		regex := regexp.MustCompile("\\[([^\\]]*)\\]")
		replace := func(str string) string {
			return fmt.Sprintf("%s[%s"+str[1:len(str)-1]+"%s]%s",
				ColorDarkBlue, ColorBlue, ColorDarkBlue, ColorWhite)
		}

		actionText := regex.ReplaceAllStringFunc(action.text, replace)

		WriteLine(conn, "  "+actionText, cm)
	}
}

// vim: nocindent
