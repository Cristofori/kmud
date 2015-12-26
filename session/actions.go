package session

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Cristofori/kmud/combat"
	"github.com/Cristofori/kmud/events"
	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type action struct {
	alias string
	exec  func(*Session, string)
}

var actions = map[string]action{
	"l": aAlias("look"),
	"look": {
		exec: func(s *Session, arg string) {
			if arg == "" {
				s.PrintRoom()
			} else {
				dir := types.StringToDirection(arg)

				if dir == types.DirectionNone {
					charList := model.CharactersIn(s.pc.GetRoomId())
					index := utils.BestMatch(arg, charList.Names())

					if index == -2 {
						s.printError("Which one do you mean?")
					} else if index != -1 {
						char := charList[index]
						s.printLine("Looking at: %s", char.GetName())
						s.printLine("    Health: %v/%v", char.GetHitPoints(), char.GetHealth())
					} else {
						itemList := model.ItemsIn(s.GetRoom())
						index = utils.BestMatch(arg, itemList.Names())

						if index == -1 {
							s.printLine("Nothing to see")
						} else if index == -2 {
							s.printError("Which one do you mean?")
						} else {
							s.printLine("Looking at: %s", itemList[index].GetName())
						}
					}
				} else {
					if s.GetRoom().HasExit(dir) {
						loc := s.GetRoom().NextLocation(dir)
						roomToSee := model.GetRoomByLocation(loc, s.GetRoom().GetZoneId())
						if roomToSee != nil {
							s.printRoom(roomToSee)
						} else {
							s.printLine("Nothing to see")
						}
					} else {
						s.printError("You can't look in that direction")
					}
				}
			}
		},
	},
	"a": aAlias("attack"),
	"attack": {
		exec: func(s *Session, arg string) {
			charList := model.CharactersIn(s.pc.GetRoomId())
			index := utils.BestMatch(arg, charList.Names())

			if index == -1 {
				s.printError("Not found")
			} else if index == -2 {
				s.printError("Which one do you mean?")
			} else {
				defender := charList[index]
				if defender.GetId() == s.pc.GetId() {
					s.printError("You can't attack yourself")
				} else {
					combat.StartFight(s.pc, nil, defender)
				}
			}
		},
	},
	"c": aAlias("cast"),
	"cast": {
		exec: func(s *Session, arg string) {
			usage := func() {
				s.printError("Usage: cast <spell> [target]")
			}

			spell, targetName := utils.Argify(arg)

			if spell == "" {
				usage()
				return
			}

			var skill types.Skill
			skills := model.GetSkills(s.pc.GetSkills())
			index := utils.BestMatch(spell, skills.Names())

			if index == -1 {
				s.printError("Skill not found")
			} else if index == -2 {
				s.printError("Which skill do you mean?")
			} else {
				skill = skills[index]
			}

			if skill != nil {
				var target types.Character

				if targetName == "" {
					target = s.pc
				} else {
					charList := model.CharactersIn(s.pc.GetRoomId())
					index := utils.BestMatch(targetName, charList.Names())

					if index == -1 {
						s.printError("Target not found")
					} else if index == -2 {
						s.printError("Which target do you mean?")
					} else {
						target = charList[index]
					}
				}

				if target != nil {
					s.WriteLineColor(types.ColorRed, "Casting %s on %s", skill.GetName(), target.GetName())
					combat.StartFight(s.pc, skill, target)
				}
			}
		},
	},
	"sb": aAlias("skillbook"),
	"skillbook": {
		exec: func(s *Session, arg string) {
			utils.ExecMenu("Skill Book", s, func(menu *utils.Menu) {
				menu.AddAction("a", "Add", func() bool {
					utils.ExecMenu("Select a skill to add", s, func(menu *utils.Menu) {
						for i, skill := range model.GetAllSkills() {
							sk := skill
							menu.AddAction(strconv.Itoa(i+1), skill.GetName(), func() bool {
								s.pc.AddSkill(sk.GetId())
								return true
							})
						}
					})
					return true
				})

				skills := model.GetSkills(s.pc.GetSkills())
				for i, skill := range skills {
					sk := skill
					menu.AddAction(strconv.Itoa(i+1), skill.GetName(), func() bool {
						s.printLine("Skill: %v", sk.GetName())
						s.printLine("  Damage: %v", sk.GetPower())
						return true
					})
				}
			})
		},
	},
	"talk": {
		exec: func(s *Session, arg string) {
			if arg == "" {
				s.printError("Usage: talk <NPC name>")
				return
			}

			npcList := model.NpcsIn(s.pc.GetRoomId())
			index := utils.BestMatch(arg, npcList.Characters().Names())

			if index == -1 {
				s.printError("Not found")
			} else if index == -2 {
				s.printError("Which one do you mean?")
			} else {
				npc := npcList[index]
				s.printLine(npc.PrettyConversation())
			}
		},
	},
	"drop": {
		exec: func(s *Session, arg string) {
			dropUsage := func() {
				s.printError("Usage: drop <item name>")
			}

			if arg == "" {
				dropUsage()
				return
			}

			characterItems := model.GetItems(s.pc.GetItems())
			index := utils.BestMatch(arg, characterItems.Names())

			if index == -1 {
				s.printError("Not found")
			} else if index == -2 {
				s.printError("Which one do you mean?")
			} else {
				item := characterItems[index]
				s.pc.RemoveItem(item.GetId())
				s.GetRoom().AddItem(item.GetId())
				s.printLine("Dropped %s", item.GetName())
			}
		},
	},
	"take": aAlias("get"),
	"t":    aAlias("get"),
	"g":    aAlias("g"),
	"get": {
		exec: func(s *Session, arg string) {
			takeUsage := func() {
				s.printError("Usage: take <item name>")
			}

			if arg == "" {
				takeUsage()
				return
			}

			itemsInRoom := model.GetItems(s.GetRoom().GetItems())
			index := utils.BestMatch(arg, itemsInRoom.Names())

			if index == -2 {
				s.printError("Which one do you mean?")
			} else if index == -1 {
				s.printError("Item %s not found", arg)
			} else {
				item := itemsInRoom[index]
				s.pc.AddItem(item.GetId())
				s.GetRoom().RemoveItem(item.GetId())
				s.printLine("Picked up %s", item.GetName())
			}
		},
	},
	"i":   aAlias("inventory"),
	"inv": aAlias("inventory"),
	"inventory": {
		exec: func(s *Session, arg string) {
			itemIds := s.pc.GetItems()

			if len(itemIds) == 0 {
				s.printLine("You aren't carrying anything")
			} else {
				items := model.GetItems(itemIds)
				s.printLine("You are carrying: %s", strings.Join(items.Names(), ", "))
			}

			s.printLine("Cash: %v", s.pc.GetCash())
		},
	},
	"help": {
		exec: func(s *Session, arg string) {
			s.printLine("HELP!")
		},
	},
	"ls": {
		exec: func(s *Session, arg string) {
			s.printLine("Where do you think you are?!")
		},
	},
	"stop": {
		exec: func(s *Session, arg string) {
			combat.StopFight(s.pc)
		},
	},
	"go": {
		exec: func(s *Session, arg string) {
			if arg == "" {
				s.printError("Usage: go <name>")
				return
			}

			links := s.GetRoom().GetLinks()
			linkNames := s.GetRoom().LinkNames()
			index := utils.BestMatch(arg, linkNames)

			if index == -2 {
				s.printError("Which one do you mean?")
			} else if index == -1 {
				s.printError("Exit %s not found", arg)
			} else {
				destId := links[linkNames[index]]
				newRoom := model.GetRoom(destId)
				model.MoveCharacterToRoom(s.pc, newRoom)
				s.PrintRoom()
			}
		},
	},
	"lock": {
		exec: func(s *Session, arg string) {
			if arg == "" {
				s.printError("Usage: lock <direction>")
			}
			handleLock(s, arg, true)
		},
	},
	"unlock": {
		exec: func(s *Session, arg string) {
			if arg == "" {
				s.printError("Usage: unlock <direction>")
			}
			handleLock(s, arg, false)
		},
	},
	"buy": {
		exec: func(s *Session, arg string) {
			store := model.StoreIn(s.pc.GetRoomId())
			if store == nil {
				s.printError("There is no store here")
				return
			}

			if arg == "" {
				s.printError("Usage: buy <item name>")
			}

			items := model.GetItems(store.GetItems())
			index, err := strconv.Atoi(arg)
			var item types.Item

			if err == nil {
				index--
				if index < len(items) && index >= 0 {
					item = items[index]
				} else {
					s.printError("Invalid selection")
				}
			} else {
				index := utils.BestMatch(arg, items.Names())

				if index == -1 {
					s.printError("Not found")
				} else if index == -2 {
					s.printError("Which one do you mean?")
				} else {
					item = items[index]
				}
			}

			if item != nil {
				confirmed := s.getConfirmation(fmt.Sprintf("Buy %s for %v? ", item.GetName(), item.GetValue()))

				if confirmed {
					if store.RemoveItem(item.GetId()) {
						s.pc.AddItem(item.GetId())
						s.pc.RemoveCash(item.GetValue())
						store.AddCash(item.GetValue())
						s.printLine(types.Colorize(types.ColorGreen, "Bought %s"), item.GetName())
					} else {
						s.printError("That item is no longer available")
					}
				} else {
					s.printError("Purchase canceled")
				}
			}
		},
	},
	"sell": {
		exec: func(s *Session, arg string) {
			store := model.StoreIn(s.pc.GetRoomId())
			if store == nil {
				s.printError("There is no store here")
				return
			}

			if arg == "" {
				s.printError("Usage: sell <item name>")
			}

			items := model.GetItems(s.pc.GetItems())
			index := utils.BestMatch(arg, items.Names())

			if index == -1 {
				s.printError("Not found")
			} else if index == -2 {
				s.printError("Which one do you mean?")
			} else {
				item := items[index]

				confirmed := s.getConfirmation(fmt.Sprintf("Sell %s for %v? ", item.GetName(), item.GetValue()))

				if confirmed {
					if s.pc.RemoveItem(item.GetId()) {
						store.AddItem(item.GetId())
						store.RemoveCash(item.GetValue())
						s.pc.AddCash(item.GetValue())
						s.printLine(types.Colorize(types.ColorGreen, "Sold %s"), item.GetName())
					} else {
						s.printError("Transaction failed")
					}
				}
			}
		},
	},
	"store": {
		exec: func(s *Session, arg string) {
			store := model.StoreIn(s.pc.GetRoomId())
			if store == nil {
				s.printError("There is no store here")
				return
			}

			s.printLine("\r\nStore cash: %v", store.GetCash())

			itemIds := store.GetItems()
			if len(itemIds) == 0 {
				s.printLine("This store is empty")
			}

			for i, item := range model.GetItems(itemIds) {
				s.printLine("[%v] %s - %v", i+1, item.GetName(), item.GetValue())
			}

			s.printLine("")
		},
	},
	"o": aAlias("open"),
	"open": {
		exec: func(s *Session, arg string) {
			items := model.ItemsIn(s.GetRoom())
			containers := types.ItemList{}

			for _, item := range items {
				if item.GetCapacity() > 0 {
					containers = append(containers, item)
				}
			}

			if len(containers) == 0 {
				s.printError("There's nothing here to open")
			} else {
				index := utils.BestMatch(arg, containers.Names())

				if index == -2 {
					s.printError("Which one do you mean?")
				} else if index != -1 {
					container := containers[index]

					utils.ExecMenu(container.GetName(), s, func(menu *utils.Menu) {
						menu.AddAction("d", "Deposit", func() bool {
							if len(s.pc.GetItems()) == 0 {
								s.printError("You have nothing to deposit")
							} else {
								utils.ExecMenu(fmt.Sprintf("Deposit into %s", container.GetName()), s, func(menu *utils.Menu) {
									for i, item := range model.GetItems(s.pc.GetItems()) {
										locItem := item
										menu.AddAction(strconv.Itoa(i+1), item.GetName(), func() bool {
											if s.pc.RemoveItem(locItem.GetId()) {
												container.AddItem(locItem.GetId())
											}
											return len(s.pc.GetItems()) > 0
										})
									}
								})
							}

							return true
						})

						menu.AddAction("w", "Withdraw", func() bool {
							if len(container.GetItems()) == 0 {
								s.printError("There is nothing to withdraw")
							} else {
								utils.ExecMenu(fmt.Sprintf("Withdraw from %s", container.GetName()), s, func(menu *utils.Menu) {
									for i, item := range model.GetItems(container.GetItems()) {
										locItem := item
										menu.AddAction(strconv.Itoa(i+1), item.GetName(), func() bool {
											if container.RemoveItem(locItem.GetId()) {
												s.pc.AddItem(locItem.GetId())
											}
											return len(container.GetItems()) > 0
										})
									}
								})
							}

							return true
						})
					})
				}
			}
		},
	},
}

func handleLock(s *Session, arg string, locked bool) {
	dir := types.StringToDirection(arg)

	if dir == types.DirectionNone {
		s.printError("Invalid direction")
	} else {
		s.GetRoom().SetLocked(dir, locked)

		events.Broadcast(events.LockEvent{
			RoomId: s.pc.GetRoomId(),
			Exit:   dir,
			Locked: locked,
		})

		// Lock on both sides
		location := s.GetRoom().NextLocation(dir)
		otherRoom := model.GetRoomByLocation(location, s.GetRoom().GetZoneId())

		if otherRoom != nil {
			otherRoom.SetLocked(dir.Opposite(), locked)

			events.Broadcast(events.LockEvent{
				RoomId: otherRoom.GetId(),
				Exit:   dir.Opposite(),
				Locked: locked,
			})
		}
	}
}

func aAlias(name string) action {
	return action{alias: name}
}
