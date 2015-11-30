package utils

import (
	"fmt"
	"strings"

	"github.com/Cristofori/kmud/types"
)

type Menu struct {
	actions     []action
	title       string
	prompt      string
	exitHandler func()
}

func ExecMenu(title string, comm types.Communicable, build func(*Menu)) {
	for {
		var menu Menu
		menu.title = title
		menu.prompt = "> "
		build(&menu)

		menu.Print(comm)

		input := comm.GetInput(types.Colorize(types.ColorWhite, menu.prompt))

		if input == "" {
			if menu.exitHandler != nil {
				menu.exitHandler()
			}
			return
		}

		action := menu.getAction(input)

		if action.handler != nil {
			if !action.handler() {
				return
			}
		} else if input != "?" && input != "help" {
			comm.WriteLine(types.Colorize(types.ColorRed, "Invalid selection"))
		}
	}
}

type action struct {
	key     string
	text    string
	data    types.Id
	handler func() bool
}

func (self *Menu) AddAction(key string, text string, handler func() bool) {
	self.actions = append(self.actions,
		action{key: strings.ToLower(key),
			text:    text,
			handler: handler,
		})
}

func (self *Menu) OnExit(handler func()) {
	self.exitHandler = handler
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

func (self *Menu) Print(comm types.Communicable) {
	border := types.Colorize(types.ColorWhite, "-=-=-")
	title := types.Colorize(types.ColorBlue, self.title)
	comm.WriteLine(fmt.Sprintf("%s %s %s", border, title, border))

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

		comm.WriteLine(fmt.Sprintf("  %s", actionText))
	}
}
