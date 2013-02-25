package session

import (
	"kmud/database"
	"kmud/model"
	"kmud/utils"
	"strings"
)

func (session *Session) handleAction(action string, args []string) {
	switch action {

	case "stop":
		session.stop()

	case "?":
		session.help()

	case "ls":
		session.ls()

	case "inventory":
		fallthrough
	case "inv":
		fallthrough
	case "i":
		session.inventory()

	case "l":
		fallthrough
	case "look":
		session.look(args)

	case "a":
		fallthrough
	case "attack":
		session.attack(args)

	case "disconnect":
		session.disconnect()

	case "talk":
		session.talk(args)

	case "drop":
		session.drop(args)

	case "take":
		fallthrough
	case "t":
		fallthrough
	case "get":
		fallthrough
	case "g":
		session.pickup(args)

	default:
		direction := database.StringToDirection(action)

		if direction != database.DirectionNone {
			if session.room.HasExit(direction) {
				newRoom, err := model.MoveCharacter(session.player, direction)
				if err == nil {
					session.room = newRoom
					session.printRoom()
				} else {
					session.printError(err.Error())
				}

			} else {
				session.printError("You can't go that way")
			}
		} else {
			session.printError("You can't do that")
		}
	}
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

func (session *Session) attack(args []string) {
	charList := model.M.CharactersIn(session.room)
	index := utils.BestMatch(args[0], database.CharacterNames(charList))

	if index == -1 {
		session.printError("Not found")
	} else if index == -2 {
		session.printError("Which one do you mean?")
	} else {
		defender := charList[index]
		if defender.GetId() == session.player.GetId() {
			session.printError("You can't attack yourself")
		} else {
			model.StartFight(session.player, defender)
		}
	}
}

func (session *Session) disconnect() {
	session.printLine("Take luck!")
	panic("User quit")
}

func (session *Session) talk(args []string) {
	if len(args) != 1 {
		session.printError("Usage: talk <NPC name>")
		return
	}

	npcList := model.M.NpcsIn(session.room)
	index := utils.BestMatch(args[0], database.CharacterNames(npcList))

	if index == -1 {
		session.printError("Not found")
	} else if index == -2 {
		session.printError("Which one do you mean?")
	} else {
		npc := npcList[index]
		session.printLine(npc.PrettyConversation(session.user.GetColorMode()))
	}
}

func (session *Session) drop(args []string) {
	dropUsage := func() {
		session.printError("Usage: drop <item name>")
	}

	if len(args) != 1 {
		dropUsage()
		return
	}

	characterItems := model.M.GetItems(session.player.GetItemIds())
	index := utils.BestMatch(args[0], database.ItemNames(characterItems))

	if index == -1 {
		session.printError("Not found")
	} else if index == -2 {
		session.printError("Which one do you mean?")
	} else {
		item := characterItems[index]
		session.player.RemoveItem(item)
		session.room.AddItem(item)
		session.printLine("Dropped %s", item.PrettyName())
	}
}

func (session *Session) pickup(args []string) {
	takeUsage := func() {
		session.printError("Usage: take <item name>")
	}

	if len(args) != 1 {
		takeUsage()
		return
	}

	itemsInRoom := model.M.GetItems(session.room.GetItemIds())
	index := utils.BestMatch(args[0], database.ItemNames(itemsInRoom))

	if index == -2 {
		session.printError("Which one do you mean?")
	} else if index == -1 {
		session.printError("Item %s not found", args[0])
	} else {
		item := itemsInRoom[index]
		session.player.AddItem(item)
		session.room.RemoveItem(item)
		session.printLine("Picked up %s", item.PrettyName())
	}
}

func (session *Session) inventory() {
	itemIds := session.player.GetItemIds()

	if len(itemIds) == 0 {
		session.printLine("You aren't carrying anything")
	} else {
		var itemNames []string
		for _, item := range model.M.GetItems(itemIds) {
			itemNames = append(itemNames, item.PrettyName())
		}
		session.printLine("You are carrying: %s", strings.Join(itemNames, ", "))
	}

	session.printLine("Cash: %v", session.player.GetCash())
}

func (session *Session) help() {
	session.printLine("HELP!")
}

func (session *Session) ls() {
	session.printLine("Where do you think you are?!")
}

func (session *Session) stop() {
	model.StopFight(session.player)
}

// vim: nocindent
