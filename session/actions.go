package session

import (
	"kmud/database"
	"kmud/model"
	"kmud/utils"
)

var actionMap map[string]func([]string)

func registerAction(action string, actionFunc func([]string)) {
	if actionMap == nil {
		actionMap = map[string]func([]string){}
	}

	actionMap[action] = actionFunc
}

func registerActions(actions []string, actionFunc func([]string)) {
	for _, action := range actions {
		registerAction(action, actionFunc)
	}
}

func callAction(action string, args []string) bool {
	actionFunc, found := actionMap[action]

	if !found {
		return false
	}

	actionFunc(args)
	return true
}

func makeList(argList ...string) []string {
	list := make([]string, len(argList))
	for i, arg := range argList {
		list[i] = arg
	}
	return list
}

func (session *Session) look(args []string) {
	if len(args) == 0 {
		session.printRoom()
	} else if len(args) == 1 {
		arg := database.StringToDirection(args[0])

		if arg == database.DirectionNone {
			charList := model.M.CharactersIn(session.room)
			index := utils.BestMatch(args[0], database.CharacterNames(charList))

			if index == -2 {
				session.printError("Which one do you mean?")
			} else if index != -1 {
				session.printLine("Looking at: %s", charList[index].PrettyName())
			} else {
				itemList := model.M.ItemsIn(session.room)
				index = utils.BestMatch(args[0], database.ItemNames(itemList))

				if index == -1 {
					session.printLine("Nothing to see")
				} else if index == -2 {
					session.printError("Which one do you mean?")
				} else {
					session.printLine("Looking at: %s", itemList[index].PrettyName())
				}
			}
		} else {
			if session.room.HasExit(arg) {
				loc := session.room.NextLocation(arg)
				roomToSee := model.M.GetRoomByLocation(loc, session.zone)
				if roomToSee != nil {
					session.printLine(roomToSee.ToString(database.ReadMode, session.user.GetColorMode(),
						model.M.PlayersIn(roomToSee, nil), model.M.NpcsIn(roomToSee), nil))
				} else {
					session.printLine("Nothing to see")
				}
			} else {
				session.printError("You can't look in that direction")
			}
		}
	}
}

// vim: nocindent
