package utils

import (
	"fmt"
	"github.com/Cristofori/kmud/types"
	"io"
	"strconv"
	"strings"
)

type Menu struct {
	actions []action
	title   string
	prompt  string
}

func NewMenu(text string) *Menu {
	var menu Menu
	menu.title = text
	menu.prompt = "> "
	return &menu
}

type action struct {
	key  string
	text string
	data types.Id
}

func (self *Menu) AddAction(key string, text string) {
	self.addAction(key, text, nil)
}

func (self *Menu) AddActionData(key int, text string, data types.Id) {
	keyStr := strconv.Itoa(key)
	self.addAction(keyStr, text, data)
}

func (self *Menu) addAction(key string, text string, data types.Id) {
	self.actions = append(self.actions, action{key: strings.ToLower(key), text: text, data: data})
}

func (self *Menu) GetData(choice string) types.Id {
	for _, action := range self.actions {
		if action.key == choice {
			return action.data
		}
	}

	return nil
}

func (self *Menu) GetPrompt() string {
	return self.prompt
}

func (self *Menu) getAction(key string) action {
	key = strings.ToLower(key)

	for _, action := range self.actions {
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

func (self *Menu) Exec(conn io.ReadWriter, cm types.ColorMode) (string, types.Id) {
	for {
		self.Print(conn, cm)
		input := GetUserInput(conn, types.Colorize(types.ColorWhite, self.prompt), cm)

		if input == "" {
			return "", nil
		}

		action := self.getAction(input)
		if action.key != "" {
			return action.key, action.data
		}
	}
}

func (self *Menu) Print(conn io.Writer, cm types.ColorMode) {
	border := types.Colorize(types.ColorWhite, "-=-=-")
	title := types.Colorize(types.ColorBlue, self.title)
	WriteLine(conn, fmt.Sprintf("%s %s %s", border, title, border), cm)

	for _, action := range self.actions {
		index := strings.Index(strings.ToLower(action.text), action.key)

		actionText := ""

		if index == -1 {
			actionText = fmt.Sprintf("%s[%s%s%s]%s%s",
				types.ColorDarkBlue,
				types.ColorBlue,
				strings.ToUpper(action.key),
				types.ColorDarkBlue,
				types.ColorWhite,
				action.text)
		} else {
			keyLength := len(action.key)
			actionText = fmt.Sprintf("%s%s[%s%s%s]%s%s",
				action.text[:index],
				types.ColorDarkBlue,
				types.ColorBlue,
				action.text[index:index+keyLength],
				types.ColorDarkBlue,
				types.ColorWhite,
				action.text[index+keyLength:])
		}

		WriteLine(conn, fmt.Sprintf("  %s", actionText), cm)
	}
}

// vim: nocindent
