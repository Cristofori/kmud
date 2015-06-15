package session

import (
	"strings"

	"github.com/Cristofori/kmud/database"
	"github.com/Cristofori/kmud/events"
	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/utils"
)

type actionHandler struct {
	session *Session
}

func (ah *actionHandler) handleAction(action string, args []string) {
	if len(args) == 0 {
		direction := database.StringToDirection(action)

		if direction != database.DirectionNone {
			if ah.session.room.HasExit(direction) {
				newRoom, err := model.MoveCharacter(ah.session.player, direction)
				if err == nil {
					ah.session.room = newRoom
					ah.session.printRoom()
				} else {
					ah.session.printError(err.Error())
				}

			} else {
				ah.session.printError("You can't go that way")
			}

			return
		}
	}

	found := utils.FindAndCallMethod(ah, action, args)

	if !found {
		ah.session.printError("You can't do that")
	}
}

func (ah *actionHandler) L(args []string) {
	ah.Look(args)
}

func (ah *actionHandler) Look(args []string) {
	if len(args) == 0 {
		ah.session.printRoom()
	} else if len(args) == 1 {
		arg := database.StringToDirection(args[0])

		if arg == database.DirectionNone {
			charList := model.PlayerCharactersIn(ah.session.room, nil)
			index := utils.BestMatch(args[0], charList.Names())

			if index == -2 {
				ah.session.printError("Which one do you mean?")
			} else if index != -1 {
				ah.session.printLine("Looking at: %s", charList[index].GetName())
			} else {
				itemList := model.ItemsIn(ah.session.room)
				index = utils.BestMatch(args[0], database.ItemNames(itemList))

				if index == -1 {
					ah.session.printLine("Nothing to see")
				} else if index == -2 {
					ah.session.printError("Which one do you mean?")
				} else {
					ah.session.printLine("Looking at: %s", itemList[index].GetName())
				}
			}
		} else {
			if ah.session.room.HasExit(arg) {
				loc := ah.session.room.NextLocation(arg)
				roomToSee := model.GetRoomByLocation(loc, ah.session.currentZone())
				if roomToSee != nil {
					area := model.GetArea(roomToSee.GetAreaId())
					ah.session.printLine(roomToSee.ToString(model.PlayerCharactersIn(roomToSee, nil),
						model.NpcsIn(roomToSee), nil, area))
				} else {
					ah.session.printLine("Nothing to see")
				}
			} else {
				ah.session.printError("You can't look in that direction")
			}
		}
	}
}

func (ah *actionHandler) A(args []string) {
	ah.Attack(args)
}

func (ah *actionHandler) Attack(args []string) {
	charList := model.CharactersIn(ah.session.room)
	index := utils.BestMatch(args[0], charList.Names())

	if index == -1 {
		ah.session.printError("Not found")
	} else if index == -2 {
		ah.session.printError("Which one do you mean?")
	} else {
		defender := charList[index]
		if defender.GetId() == ah.session.player.GetId() {
			ah.session.printError("You can't attack yourself")
		} else {
			events.StartFight(ah.session.player, defender)
		}
	}
}

func (ah *actionHandler) Disconnect(args []string) {
	ah.session.printLine("Take luck!")
	panic("User quit")
}

func (ah *actionHandler) Talk(args []string) {
	if len(args) != 1 {
		ah.session.printError("Usage: talk <NPC name>")
		return
	}

	npcList := model.NpcsIn(ah.session.room)
	index := utils.BestMatch(args[0], npcList.Names())

	if index == -1 {
		ah.session.printError("Not found")
	} else if index == -2 {
		ah.session.printError("Which one do you mean?")
	} else {
		npc := npcList[index]
		ah.session.printLine(npc.PrettyConversation())
	}
}

func (ah *actionHandler) Drop(args []string) {
	dropUsage := func() {
		ah.session.printError("Usage: drop <item name>")
	}

	if len(args) != 1 {
		dropUsage()
		return
	}

	characterItems := model.GetItems(ah.session.player.GetItemIds())
	index := utils.BestMatch(args[0], database.ItemNames(characterItems))

	if index == -1 {
		ah.session.printError("Not found")
	} else if index == -2 {
		ah.session.printError("Which one do you mean?")
	} else {
		item := characterItems[index]
		ah.session.player.RemoveItem(item)
		ah.session.room.AddItem(item)
		ah.session.printLine("Dropped %s", item.GetName())
	}
}

func (ah *actionHandler) Take(args []string) {
	ah.Pickup(args)
}

func (ah *actionHandler) T(args []string) {
	ah.Pickup(args)
}

func (ah *actionHandler) Get(args []string) {
	ah.Pickup(args)
}

func (ah *actionHandler) G(args []string) {
	ah.Pickup(args)
}

func (ah *actionHandler) Pickup(args []string) {
	takeUsage := func() {
		ah.session.printError("Usage: take <item name>")
	}

	if len(args) != 1 {
		takeUsage()
		return
	}

	itemsInRoom := model.GetItems(ah.session.room.GetItemIds())
	index := utils.BestMatch(args[0], database.ItemNames(itemsInRoom))

	if index == -2 {
		ah.session.printError("Which one do you mean?")
	} else if index == -1 {
		ah.session.printError("Item %s not found", args[0])
	} else {
		item := itemsInRoom[index]
		ah.session.player.AddItem(item)
		ah.session.room.RemoveItem(item)
		ah.session.printLine("Picked up %s", item.GetName())
	}
}

func (ah *actionHandler) I(args []string) {
	ah.Inventory(args)
}

func (ah *actionHandler) Inv(args []string) {
	ah.Inventory(args)
}

func (ah *actionHandler) Inventory(args []string) {
	itemIds := ah.session.player.GetItemIds()

	if len(itemIds) == 0 {
		ah.session.printLine("You aren't carrying anything")
	} else {
		var itemNames []string
		for _, item := range model.GetItems(itemIds) {
			itemNames = append(itemNames, item.GetName())
		}
		ah.session.printLine("You are carrying: %s", strings.Join(itemNames, ", "))
	}

	ah.session.printLine("Cash: %v", ah.session.player.GetCash())
}

func (ah *actionHandler) Help(args []string) {
	ah.session.printLine("HELP!")
}

func (ah *actionHandler) Ls(args []string) {
	ah.session.printLine("Where do you think you are?!")
}

func (ah *actionHandler) Stop(args []string) {
	events.StopFight(ah.session.player)
}

// vim: nocindent
