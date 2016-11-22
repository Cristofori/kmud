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
						s.WriteLinef("Looking at: %s", char.GetName())
						s.WriteLinef("    Health: %v/%v", char.GetHitPoints(), char.GetHealth())
					} else {
						itemList := model.ItemsIn(s.GetRoom().GetId())
						index = utils.BestMatch(arg, itemList.Names())

						if index == -1 {
							s.WriteLine("Nothing to see")
						} else if index == -2 {
							s.printError("Which one do you mean?")
						} else {
							item := itemList[index]
							s.WriteLinef("Looking at: %s", item.GetName())
							contents := model.ItemsIn(item.GetId())
							if len(contents) > 0 {
								s.WriteLinef("Contents: %s", strings.Join(contents.Names(), ", "))
							} else {
								s.WriteLine("(empty)")
							}
						}
					}
				} else {
					if s.GetRoom().HasExit(dir) {
						loc := s.GetRoom().NextLocation(dir)
						roomToSee := model.GetRoomByLocation(loc, s.GetRoom().GetZoneId())
						if roomToSee != nil {
							s.printRoom(roomToSee)
						} else {
							s.WriteLine("Nothing to see")
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
							menu.AddActionI(i, skill.GetName(), func() bool {
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
					menu.AddActionI(i, skill.GetName(), func() bool {
						s.WriteLine("Skill: %v", sk.GetName())
						s.WriteLine("  Damage: %v", sk.GetPower())
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
				s.WriteLine(npc.PrettyConversation())
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

			characterItems := model.ItemsIn(s.pc.GetId())
			index := utils.BestMatch(arg, characterItems.Names())

			if index == -1 {
				s.printError("Not found")
			} else if index == -2 {
				s.printError("Which one do you mean?")
			} else {
				item := characterItems[index]
				if item.SetContainerId(s.GetRoom().GetId(), s.pc.GetId()) {
					s.WriteLine("Dropped %s", item.GetName())
				} else {
					s.printError("Not found")
				}
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

			itemsInRoom := model.ItemsIn(s.GetRoom().GetId())
			index := utils.BestMatch(arg, itemsInRoom.Names())

			if index == -2 {
				s.printError("Which one do you mean?")
			} else if index == -1 {
				s.printError("Not found")
			} else {
				item := itemsInRoom[index]
				if item.SetContainerId(s.pc.GetId(), s.GetRoom().GetId()) {
					s.WriteLine("Picked up %s", item.GetName())
				} else {
					s.printError("Not found")
				}
			}
		},
	},
	"i":   aAlias("inventory"),
	"inv": aAlias("inventory"),
	"inventory": {
		exec: func(s *Session, arg string) {
			items := model.ItemsIn(s.pc.GetId())

			if len(items) == 0 {
				s.WriteLinef("You aren't carrying anything")
			} else {
				names := make([]string, len(items))
				for i, item := range items {
					template := model.GetTemplate(item.GetTemplateId())
					names[i] = fmt.Sprintf("%s (%v)", item.GetName(), template.GetWeight())
				}
				s.WriteLinef("You are carrying: %s", strings.Join(names, ", "))
			}

			s.WriteLinef("Cash: %v", s.pc.GetCash())
			s.WriteLinef("Weight: %v/%v", model.CharacterWeight(s.pc), s.pc.GetCapacity())
		},
	},
	"help": {
		exec: func(s *Session, arg string) {
			s.WriteLine("HELP!")
		},
	},
	"ls": {
		exec: func(s *Session, arg string) {
			s.WriteLine("Where do you think you are?!")
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

			items := model.ItemsIn(store.GetId())
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
					// TODO - Transferring the item and money should be guarded by some kind of global session transaction
					if item.SetContainerId(s.pc.GetId(), store.GetId()) {
						s.pc.RemoveCash(item.GetValue())
						store.AddCash(item.GetValue())
						s.WriteLineColor(types.ColorGreen, "Bought %s", item.GetName())
					} else {
						s.printError("Transaction failed")
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

			items := model.ItemsIn(s.pc.GetId())
			index := utils.BestMatch(arg, items.Names())

			if index == -1 {
				s.printError("Not found")
			} else if index == -2 {
				s.printError("Which one do you mean?")
			} else {
				item := items[index]

				confirmed := s.getConfirmation(fmt.Sprintf("Sell %s for %v? ", item.GetName(), item.GetValue()))

				if confirmed {
					// TODO - Transferring the item and money should be guarded by some kind of global session transaction
					if item.SetContainerId(store.GetId(), s.pc.GetId()) {
						store.RemoveCash(item.GetValue())
						s.pc.AddCash(item.GetValue())
						s.WriteLineColor(types.ColorGreen, "Sold %s", item.GetName())
					} else {
						s.printError("Transaction failed")
					}
				} else {
					s.printError("Sale canceled")
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

			s.WriteLine("\r\nStore cash: %v", store.GetCash())

			items := model.ItemsIn(store.GetId())
			if len(items) == 0 {
				s.WriteLine("This store is empty")
			}

			for i, item := range items {
				s.WriteLine("[%v] %s - %v", i+1, item.GetName(), item.GetValue())
			}

			s.WriteLine("")
		},
	},
	"o": aAlias("open"),
	"open": {
		exec: func(s *Session, arg string) {
			items := model.ItemsIn(s.GetRoom().GetId())
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
							if model.CountItemsIn(s.pc.GetId()) == 0 {
								s.printError("You have nothing to deposit")
							} else {
								utils.ExecMenu(fmt.Sprintf("Deposit into %s", container.GetName()), s, func(menu *utils.Menu) {
									for i, item := range model.ItemsIn(s.pc.GetId()) {
										locItem := item
										menu.AddActionI(i, item.GetName(), func() bool {
											if locItem.SetContainerId(container.GetId(), s.pc.GetId()) {
												s.WriteLine("Item deposited")
											} else {
												s.printError("Failed to deposit item")
											}
											return model.CountItemsIn(s.pc.GetId()) > 0
										})
									}
								})
							}

							return true
						})

						for i, item := range model.ItemsIn(container.GetId()) {
							locItem := item
							menu.AddActionI(i, item.GetName(), func() bool {
								if locItem.SetContainerId(s.pc.GetId(), container.GetId()) {
									s.WriteLine("Took %s from %s", locItem.GetName(), container.GetName())
								} else {
									s.printError("Failed to take item")
								}
								return true
							})
						}
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
