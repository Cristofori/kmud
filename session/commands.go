package session

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Cristofori/kmud/combat"
	"github.com/Cristofori/kmud/engine"
	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type command struct {
	admin bool
	alias string
	exec  func(*command, *Session, string)
	usage string
}

func (self *command) Usage(s *Session) {
	s.WriteLine(fmt.Sprintf("Usage: %s", self.usage))
}

var commands map[string]*command

func findCommand(name string) *command {
	return commands[strings.ToLower(name)]
}

func init() {
	commands = map[string]*command{
		"help": {
			usage: "/help <command name>",
			exec: func(self *command, s *Session, arg string) {
				if arg == "" {
					s.WriteLine("List of commands:")
					var names []string
					for name, command := range commands {
						if command.alias == "" {
							names = append(names, name)
						}
					}
					sort.Strings(names)
					width, height := s.user.GetWindowSize()

					pages := utils.Paginate(names, width, height/2)

					for _, page := range pages {
						s.WriteLine(page)
					}
				} else {
					command, found := commands[arg]
					if found {
						command.Usage(s)
					} else {
						s.printError("Command not found")
					}
				}
			},
		},
		"store": {
			admin: true,
			exec: func(self *command, s *Session, arg string) {
				s.execMenu("Store", func(menu *utils.Menu) {
					store := model.StoreIn(s.pc.GetRoomId())

					if store != nil {
						menu.AddAction("r", "Rename", func() {
							name := s.getCleanUserInput("New name: ")
							if name != "" {
								store.SetName("Name")
							}
						})
						menu.AddAction("d", "Delete", func() {
							model.DeleteStore(store.GetId())
						})
					} else {
						menu.AddAction("n", "New Store", func() {
							name := s.getCleanUserInput("Store name: ")
							if name != "" {
								model.CreateStore(name, s.pc.GetRoomId())
							}
						})
					}
				})
			},
		},
		"loc": cAlias("location"),
		"location": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				s.WriteLine("%v", s.GetRoom().GetLocation())
			},
		},
		"room": {
			admin: true,
			exec: func(self *command, s *Session, arg string) {
				s.execMenu(
					"Room",
					func(menu *utils.Menu) {
						menu.AddAction("t", fmt.Sprintf("Title - %s", s.GetRoom().GetTitle()), func() {
							title := s.getRawUserInput("Enter new title: ")
							if title != "" {
								s.GetRoom().SetTitle(title)
							}
						})

						menu.AddAction("d", "Description", func() {
							description := s.getRawUserInput("Enter new description: ")
							if description != "" {
								s.GetRoom().SetDescription(description)
							}
						})

						menu.AddAction("e", "Exits", func() {
							toggleExitMenu(s)
						})

						areaId := s.GetRoom().GetAreaId()
						areaName := "(None)"
						if areaId != nil {
							area := model.GetArea(areaId)
							areaName = area.GetName()
						}

						menu.AddAction("a", fmt.Sprintf("Area - %s", areaName), func() {
							s.execMenu("Change Area", func(menu *utils.Menu) {
								menu.AddAction("n", "None", func() {
									s.GetRoom().SetAreaId(nil)
								})

								for i, area := range model.GetAreas(s.currentZone()) {
									actionText := area.GetName()
									if area.GetId() == s.GetRoom().GetAreaId() {
										actionText += "*"
									}

									a := area
									menu.AddActionI(i, actionText, func() {
										s.GetRoom().SetAreaId(a.GetId())
									})
								}
							})
						})
					})

				s.PrintRoom()
			},
		},
		"map": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				zoneRooms := model.GetRoomsInZone(s.currentZone().GetId())
				roomsByLocation := map[types.Coordinate]types.Room{}

				for _, room := range zoneRooms {
					roomsByLocation[room.GetLocation()] = room
				}

				width, height := s.user.GetWindowSize()
				height /= 2
				width /= 2

				// width and height need to be odd numbers so that we keep the current location centered
				// and we don't go off the edge of the available space
				width += (width % 2) - 1
				height += (height % 2) - 1

				builder := newMapBuilder(width, height, 1)
				builder.setUserRoom(s.GetRoom())
				center := s.GetRoom().GetLocation()

				startX := center.X - (width / 2)
				endX := center.X + (width / 2)
				startY := center.Y - (height / 2)
				endY := center.Y + (height / 2)

				for y := startY; y <= endY; y++ {
					for x := startX; x <= endX; x++ {
						loc := types.Coordinate{X: x, Y: y, Z: center.Z}
						room := roomsByLocation[loc]

						if room != nil {
							// Translate to 0-based coordinates
							builder.addRoom(room, x-startX, y-startY, 0)
						}
					}
				}

				s.WriteLine(utils.TrimEmptyRows(builder.toString()))
			},
		},
		"zone": {
			admin: true,
			usage: "/zone [list|rename <name>|new <name>|delete <name>]",
			exec: func(self *command, s *Session, arg string) {
				subcommand, arg := utils.Argify(arg)
				if subcommand == "" {
					s.WriteLine("Current zone: " + types.Colorize(types.ColorBlue, s.currentZone().GetName()))
				} else if subcommand == "list" {
					s.WriteLineColor(types.ColorBlue, "Zones")
					s.WriteLineColor(types.ColorBlue, "-----")
					for _, zone := range model.GetZones() {
						s.WriteLine(zone.GetName())
					}
				} else if arg != "" {
					if subcommand == "rename" {
						zone := model.GetZoneByName(arg)

						if zone != nil {
							s.printError("A zone with that name already exists")
							return
						}

						s.currentZone().SetName(arg)
					} else if subcommand == "new" {
						newZone, err := model.CreateZone(arg)

						if err != nil {
							s.printError(err.Error())
							return
						}

						newRoom, err := model.CreateRoom(newZone, types.Coordinate{X: 0, Y: 0, Z: 0})
						utils.HandleError(err)

						model.MoveCharacterToRoom(s.pc, newRoom)

						s.PrintRoom()
					} else if subcommand == "delete" {
						zone := model.GetZoneByName(arg)

						if zone != nil {
							if zone == s.currentZone() {
								s.printError("You can't delete the zone you are in")
							} else {
								model.DeleteZone(zone.GetId())
								s.WriteLine("Zone deleted")
							}
						} else {
							s.printError("Zone not found")
						}
					} else {
						self.Usage(s)
					}
				} else {
					self.Usage(s)
				}
			},
		},
		"b": cAlias("broadcast"),
		"broadcast": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				if arg == "" {
					s.printError("Nothing to say")
				} else {
					model.BroadcastMessage(s.pc, arg)
				}
			},
		},
		"s": cAlias("say"),
		"say": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				if arg == "" {
					s.printError("Nothing to say")
				} else {
					model.Say(s.pc, arg)
				}
			},
		},
		"me": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				model.Emote(s.pc, arg)
			},
		},
		"w":    cAlias("whisper"),
		"tell": cAlias("whisper"),
		"whisper": {
			admin: false,
			usage: "/whisper <player> <message>",
			exec:  whisper,
		},
		"tp": cAlias("teleport"),
		"teleport": {
			admin: true,
			usage: "/teleport [<zone>|<X> <Y> <Z>]",
			exec: func(self *command, s *Session, arg string) {
				newZone := model.GetZoneByName(arg)
				var newRoom types.Room

				if newZone != nil {
					if newZone.GetId() == s.GetRoom().GetZoneId() {
						s.WriteLine("You're already in that zone")
					} else {
						zoneRooms := model.GetRoomsInZone(newZone.GetId())
						if len(zoneRooms) > 0 {
							newRoom = zoneRooms[0]
						}
					}
				} else {
					coords, err := utils.Atois(strings.Fields(arg))

					var x, y, z int

					if len(coords) != 3 || err != nil {
						s.printError("Zone not found: %s", arg)
						self.Usage(s)
					} else {
						x, y, z = coords[0], coords[1], coords[2]
					}

					newRoom = model.GetRoomByLocation(types.Coordinate{X: x, Y: y, Z: z}, s.currentZone().GetId())

					if newRoom == nil {
						s.printError("Invalid coordinates")
					}
				}

				if newRoom != nil {
					model.MoveCharacterToRoom(s.pc, newRoom)
					s.PrintRoom()
				}
			},
		},
		"who": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				chars := model.GetOnlinePlayerCharacters()

				s.WriteLine("")
				s.WriteLine("Online Players")
				s.WriteLine("--------------")

				for _, char := range chars {
					s.WriteLine(char.GetName())
				}
				s.WriteLine("")
			},
		},
		"colors": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				s.WriteLineColor(types.ColorNormal, "Normal")
				s.WriteLineColor(types.ColorRed, "Red")
				s.WriteLineColor(types.ColorDarkRed, "Dark Red")
				s.WriteLineColor(types.ColorGreen, "Green")
				s.WriteLineColor(types.ColorDarkGreen, "Dark Green")
				s.WriteLineColor(types.ColorBlue, "Blue")
				s.WriteLineColor(types.ColorDarkBlue, "Dark Blue")
				s.WriteLineColor(types.ColorYellow, "Yellow")
				s.WriteLineColor(types.ColorDarkYellow, "Dark Yellow")
				s.WriteLineColor(types.ColorMagenta, "Magenta")
				s.WriteLineColor(types.ColorDarkMagenta, "Dark Magenta")
				s.WriteLineColor(types.ColorCyan, "Cyan")
				s.WriteLineColor(types.ColorDarkCyan, "Dark Cyan")
				s.WriteLineColor(types.ColorBlack, "Black")
				s.WriteLineColor(types.ColorWhite, "White")
				s.WriteLineColor(types.ColorGray, "Gray")
			},
		},
		"cm": cAlias("colormode"),
		"colormode": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				if arg == "" {
					message := "Current color mode is: "
					switch s.user.GetColorMode() {
					case types.ColorModeNone:
						message = message + "None"
					case types.ColorModeLight:
						message = message + "Light"
					case types.ColorModeDark:
						message = message + "Dark"
					}
					s.WriteLine(message)
				} else {
					switch strings.ToLower(arg) {
					case "none":
						s.user.SetColorMode(types.ColorModeNone)
						s.WriteLine("Color mode set to: None")
					case "light":
						s.user.SetColorMode(types.ColorModeLight)
						s.WriteLine("Color mode set to: Light")
					case "dark":
						s.user.SetColorMode(types.ColorModeDark)
						s.WriteLine("Color mode set to: Dark")
					default:
						s.WriteLine("Valid color modes are: None, Light, Dark")
					}
				}
			},
		},
		"dr": cAlias("destroyroom"),
		"destroyroom": {
			admin: true,
			usage: "/destroyroom <direction>",
			exec: func(self *command, s *Session, arg string) {
				if arg == "" {
					self.Usage(s)
				} else {
					direction := types.StringToDirection(arg)

					if direction == types.DirectionNone {
						s.printError("Not a valid direction")
					} else {
						loc := s.GetRoom().NextLocation(direction)
						roomToDelete := model.GetRoomByLocation(loc, s.GetRoom().GetZoneId())
						if roomToDelete != nil {
							model.DeleteRoom(roomToDelete)
							s.WriteLine("Room destroyed")
						} else {
							s.printError("No room in that direction")
						}
					}
				}
			},
		},
		"npc": {
			admin: true,
			exec: func(self *command, s *Session, arg string) {
				s.execMenu("NPCs", func(menu *utils.Menu) {
					var npcs types.NPCList
					npcs = model.GetNpcs()

					menu.AddAction("n", "New", func() {
						name := s.getName("Desired NPC name: ", types.NpcType)
						if name != "" {
							model.CreateNpc(name, s.pc.GetRoomId(), nil)
						}
					})

					for i, npc := range npcs {
						n := npc
						menu.AddActionI(i, npc.GetName(), func() {
							s.specificNpcMenu(n)
						})
					}
				})

				s.PrintRoom()
			},
		},
		"items": {
			admin: true,
			usage: "Usage: /items",
			exec: func(self *command, s *Session, arg string) {
				s.execMenu("Items", func(menu *utils.Menu) {
					menu.AddAction("n", "New", func() {
						name := s.getRawUserInput("Item name: ")
						if name != "" {
							template := model.CreateTemplate(name)
							templateMenu(s, template)
						}
					})

					for i, template := range model.GetAllTemplates() {
						t := template
						menu.AddActionI(i, template.GetName(), func() {
							templateMenu(s, t)
						})
					}
				})
			},
		},
		"destroyitem": {
			admin: true,
			usage: "/destroyitem <item name>",
			exec: func(self *command, s *Session, arg string) {
				if arg == "" {
					self.Usage(s)
					return
				} else {
					itemsInRoom := model.ItemsIn(s.GetRoom().GetId())
					name := strings.ToLower(arg)

					for _, item := range itemsInRoom {
						if strings.ToLower(item.GetName()) == name {
							model.DeleteItem(item.GetId())
							s.WriteLine("Item destroyed")
							return
						}
					}

					s.printError("Item not found")
				}
			},
		},
		"roomid": {
			admin: true,
			exec: func(self *command, s *Session, arg string) {
				s.WriteLine("Room ID: %v", s.GetRoom().GetId())
			},
		},
		"cash": {
			admin: true,
			usage: "/cash give <amount>",
			exec: func(self *command, s *Session, arg string) {
				subcommand, arg := utils.Argify(arg)

				if subcommand == "give" {
					amount, err := utils.Atoir(arg, 1, math.MaxInt32)
					if err == nil {
						s.pc.AddCash(amount)
						s.WriteLine("Received: %v monies", amount)
					} else {
						s.printError(err.Error())
						self.Usage(s)
					}
				} else {
					self.Usage(s)
				}
			},
		},
		"ws": cAlias("windowsize"),
		"windowsize": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				width, height := s.user.GetWindowSize()

				header := fmt.Sprintf("Width: %v, Height: %v", width, height)

				topBar := header + " " + strings.Repeat("-", int(width)-2-len(header)) + "+"
				bottomBar := "+" + strings.Repeat("-", int(width)-2) + "+"
				outline := "|" + strings.Repeat(" ", int(width)-2) + "|"

				s.WriteLine(topBar)

				for i := 0; i < int(height)-3; i++ {
					s.WriteLine(outline)
				}

				s.WriteLine(bottomBar)
			},
		},
		"tt": cAlias("terminaltype"),
		"terminaltype": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				s.WriteLine("Terminal type: %s", s.user.GetTerminalType())
			},
		},
		"silent": {
			admin: false,
			usage: "/silent [on|off]",
			exec: func(self *command, s *Session, arg string) {
				if arg == "" {
					s.silentMode = !s.silentMode
				} else if arg == "on" {
					s.silentMode = true
				} else if arg == "off" {
					s.silentMode = false
				} else {
					self.Usage(s)
				}

				if s.silentMode {
					s.WriteLine("Silent mode on")
				} else {
					s.WriteLine("Silent mode off")
				}
			},
		},
		"r": cAlias("reply"),
		"reply": {
			admin: false,
			exec: func(self *command, s *Session, arg string) {
				targetChar := model.GetPlayerCharacter(s.replyId)

				if targetChar == nil {
					s.asyncMessage("No one to reply to")
				} else if arg == "" {
					prompt := "Reply to " + targetChar.GetName() + ": "
					input := s.getRawUserInput(prompt)

					if input != "" {
						newArg := fmt.Sprintf("%s %s", targetChar.GetName(), input)
						whisper(commands["whisper"], s, newArg)
					}
				} else {
					newArg := fmt.Sprintf("%s %s", targetChar.GetName(), arg)
					whisper(commands["whisper"], s, newArg)
				}
			},
		},
		"area": {
			admin: true,
			exec: func(self *command, s *Session, arg string) {
				s.execMenu("Areas", func(menu *utils.Menu) {
					menu.AddAction("n", "New", func() {
						name := s.getName("Area name: ", types.AreaType)
						if name != "" {
							model.CreateArea(name, s.currentZone())
						}
					})

					for i, area := range model.GetAreas(s.currentZone()) {
						a := area
						menu.AddActionI(i, area.GetName(), func() {
							s.specificAreaMenu(a)
						})
					}
				})
			},
		},
		"link": {
			admin: true,
			usage: "Usage: /link <name> [single|double*] to start, /link to finish, /link remove <name> [single|double*], /link rename <old name> <new name>, /link cancel",
			exec: func(self *command, s *Session, arg string) {
				args := strings.Split(arg, " ")
				StateName := "Linking"

				linkName, linking := s.states[StateName]

				if linking {
					if len(args) == 1 && args[0] == "cancel" {
						linkData.source = nil
						delete(s.states, StateName)
					} else if len(args) != 0 {
						self.Usage(s)
					} else {
						sourceRoom := model.GetRoom(linkData.source)

						sourceRoom.SetLink(linkName, s.pc.GetRoomId())
						if linkData.mode == LinkDouble {
							s.GetRoom().SetLink(linkName, linkData.source)
						}

						linkData.source = nil
						delete(s.states, StateName)

						s.PrintRoom()
					}
				} else {
					if len(args) == 0 {
						self.Usage(s)
						return
					}

					if args[0] == "remove" {
						mode := "double"
						if len(args) == 3 {
							if args[2] == "single" || args[2] == "double" {
								mode = args[2]
							} else {
								self.Usage(s)
								return
							}
						}

						if len(args) != 2 {
							self.Usage(s)
							return
						}

						linkNames := s.GetRoom().LinkNames()
						index := utils.BestMatch(args[1], linkNames)

						if index == -2 {
							s.printError("Which one do you mean?")
						} else if index == -1 {
							s.printError("Link not found")
						} else {
							linkName := linkNames[index]

							if mode == "double" {
								links := s.GetRoom().GetLinks()
								linkedRoom := model.GetRoom(links[linkName])
								linkedRoom.RemoveLink(linkName)
							}

							s.GetRoom().RemoveLink(linkName)
							s.PrintRoom()
						}
					} else if args[0] == "rename" {
						// TODO
					} else {
						if len(args) == 2 {
							if args[1] == LinkSingle || args[1] == LinkDouble {
								linkData.mode = args[1]
							} else {
								self.Usage(s)
								return
							}
						} else {
							linkData.mode = LinkDouble
						}

						// New link
						s.states[StateName] = utils.FormatName(args[0])
						linkData.source = s.pc.GetRoomId()
					}
				}
			},
		},
		"kill": {
			admin: true,
			usage: "/kill [npc name]",
			exec: func(self *command, s *Session, arg string) {
				if arg == "" {
					self.Usage(s)
				} else {
					npcs := model.NpcsIn(s.pc.GetRoomId())
					index := utils.BestMatch(arg, npcs.Characters().Names())

					if index == -1 {
						s.printError("Not found")
					} else if index == -2 {
						s.printError("Which one do you mean?")
					} else {
						npc := npcs[index]
						combat.Kill(npc)
						s.WriteLine("Killed %s", npc.GetName())
					}
				}
			},
		},
		"inspect": {
			admin: true,
			usage: "/inspect [name]",
			exec: func(self *command, s *Session, arg string) {
				if arg == "" {
					self.Usage(s)
				} else {
					characters := model.CharactersIn(s.pc.GetRoomId())
					index := utils.BestMatch(arg, characters.Names())

					if index == -1 {
						s.printError("Not found")
					} else if index == -2 {
						s.printError("Which one do you mean?")
					} else {
						char := characters[index]

						s.WriteLine(char.GetName())
						s.WriteLine("Health: %v/%v", char.GetHitPoints(), char.GetHealth())
					}
				}
			},
		},
		"skills": {
			admin: true,
			exec: func(self *command, s *Session, arg string) {
				s.execMenu("Skills", func(menu *utils.Menu) {
					menu.AddAction("n", "New", func() {
						name := s.getName("Skill name: ", types.SkillType)
						if name != "" {
							model.CreateSkill(name)
						}
					})

					for i, skill := range model.GetAllSkills() {
						sk := skill
						menu.AddActionI(i, skill.GetName(), func() {
							s.specificSkillMenu(sk)
						})
					}
				})
			},
		},
		"effects": {
			admin: true,
			exec: func(self *command, s *Session, arg string) {
				s.execMenu("Effects", func(menu *utils.Menu) {
					menu.AddAction("n", "New", func() {
						name := s.getName("Effect name: ", types.EffectType)
						if name != "" {
							model.CreateEffect(name)
						}
					})

					effects := model.GetAllEffects()
					for i, effect := range effects {
						e := effect
						menu.AddActionI(i, e.GetName(), func() {
							s.specificEffectMenu(e)
						})
					}
				})
			},
		},
		"testmenu": {
			admin: true,
			exec: func(self *command, s *Session, arg string) {
				s.execMenu("Menu Test", func(menu *utils.Menu) {
					for i := 0; i < 500; i++ {
						menu.AddActionI(i, fmt.Sprintf("Test Item %v", i), func() {
							menu.Exit()
						})
					}
				})
			},
		},
		"path": {
			admin: true,
			usage: "/path <coordinates>",
			exec: func(self *command, s *Session, arg string) {
				coords, err := utils.Atois(strings.Fields(arg))
				if err == nil {
					if len(coords) == 2 || len(coords) == 3 {
						x, y := coords[0], coords[1]

						z := s.GetRoom().GetLocation().Z
						if len(coords) == 3 {
							z = coords[2]
						}

						room := model.GetRoomByLocation(types.Coordinate{X: x, Y: y, Z: z}, s.GetRoom().GetZoneId())
						if room != nil {
							path := engine.FindPath(s.GetRoom(), room)
							/*
								s.WriteLine("Path:")
								for _, room := range path {
									s.WriteLine(fmt.Sprintf("%v", room.GetLocation()))
								}
							*/
							if len(path) > 0 {
								for _, room := range path {
									time.Sleep(200 * time.Millisecond)
									model.MoveCharacterToRoom(s.pc, room)
									s.PrintRoom()
									s.handleCommand("map", "")
								}
							} else {
								s.printError("No path found")
							}
						} else {
							s.printError("No room found at the given coordinates")
						}
					} else {
						self.Usage(s)
					}
				} else {
					self.Usage(s)
				}
			},
		},
		"time": {
			admin: false,
			usage: "/time",
			exec: func(self *command, s *Session, arg string) {
				s.WriteLine("%v", model.GetWorld().GetTime())
			},
		},
		"join": {
			admin: true,
			usage: "/join <player name>",
			exec: func(self *command, s *Session, arg string) {
				target := model.GetCharacterByName(arg)
				if target == nil {
					s.printError("Target not found")
				} else {
					model.MoveCharacterToRoom(s.pc, model.GetRoom(target.GetRoomId()))
					s.PrintRoom()
				}
			},
		},
		"bring": {
			admin: true,
			usage: "/bring <player name>",
			exec: func(self *command, s *Session, arg string) {
				// TODO - The target receives no indication that they've been moved
				target := model.GetCharacterByName(arg)
				if target == nil {
					s.printError("Target not found")
				} else {
					model.MoveCharacterToRoom(target, s.GetRoom())
				}
			},
		},
	}
}

func (s *Session) specificSkillMenu(skill types.Skill) {
	s.execMenu("", func(menu *utils.Menu) {
		menu.SetTitle(fmt.Sprintf("Skill - %s", skill.GetName()))
		menu.AddAction("n", "Name", func() {
			// TODO
		})

		menu.AddAction("e", "Effects", func() {
			s.execMenu("", func(menu *utils.Menu) {
				menu.SetTitle(fmt.Sprintf("%s - Effects", skill.GetName()))
				effects := model.GetAllEffects()
				for i, effect := range effects {
					e := effect
					menu.AddActionI(i, e.GetName(), func() {
						s.specificEffectMenu(e)
					})
				}
			})
		})
	})
}

func (s *Session) specificEffectMenu(effect types.Effect) {
	s.execMenu("", func(menu *utils.Menu) {
		menu.SetTitle(fmt.Sprintf("Effect - %s", effect.GetName()))
		menu.AddAction("r", "Rename", func() {
			name := s.getName("New name: ", types.EffectType)
			if name != "" {
				effect.SetName(name)
			}
		})
		menu.AddAction("t", fmt.Sprintf("Type - %v", effect.GetType()), func() {
			s.execMenu("New Effect: ", func(menu *utils.Menu) {
				menu.AddAction("h", "Hitpoint", func() {
					effect.SetType(types.HitpointEffect)
					menu.Exit()
				})
				menu.AddAction("s", "Stun", func() {
					effect.SetType(types.StunEffect)
					menu.Exit()
				})
			})
		})
		menu.AddAction("p", fmt.Sprintf("Power - %v", effect.GetPower()), func() {
			dmg, valid := s.getInt("New power: ", 0, 1000)
			if valid {
				effect.SetPower(dmg)
			}
		})
		menu.AddAction("c", fmt.Sprintf("Cost - %v", effect.GetCost()), func() {
			cost, valid := s.getInt("New cost: ", 0, 1000)
			if valid {
				effect.SetCost(cost)
			}
		})
		menu.AddAction("v", fmt.Sprintf("Variance - %v", effect.GetVariance()), func() {
			variance, valid := s.getInt("New variance: ", 0, 1000)
			if valid {
				effect.SetVariance(variance)
			}
		})
		menu.AddAction("s", fmt.Sprintf("Speed - %v", effect.GetSpeed()), func() {
			speed, valid := s.getInt("New speed: ", 0, 1000)
			if valid {
				effect.SetSpeed(speed)
			}
		})
		menu.AddAction("i", fmt.Sprintf("Time - %v", effect.GetTime()), func() {
			time, valid := s.getInt("New time: ", 0, 1000)
			if valid {
				effect.SetTime(time)
			}
		})
		menu.AddAction("d", "Delete", func() {
			model.DeleteSkill(effect.GetId())
			menu.Exit()
		})
	})
}

func (s *Session) specificNpcMenu(npc types.NPC) {
	s.execMenu(npc.GetName(), func(menu *utils.Menu) {
		menu.AddAction("r", "Rename", func() {
			name := s.getName("Desired NPC name: ", types.NpcType)
			if name != "" {
				npc.SetName(name)
			}
		})

		menu.AddAction("d", "Delete", func() {
			model.DeleteCharacter(npc.GetId())
			menu.Exit()
		})

		menu.AddAction("c", "Conversation", func() {
			conversation := npc.GetConversation()

			if conversation == "" {
				conversation = "<empty>"
			}

			s.WriteLine("Conversation: %s", conversation)
			newConversation := s.getRawUserInput("New conversation text: ")

			if newConversation != "" {
				npc.SetConversation(newConversation)
			}
		})

		roamingState := "Off"
		if npc.GetRoaming() {
			roamingState = "On"
		}

		menu.AddAction("o", fmt.Sprintf("Roaming - %s", roamingState), func() {
			npc.SetRoaming(!npc.GetRoaming())
		})
	})
}

func templateMenu(s *Session, template types.Template) {
	s.execMenu("", func(menu *utils.Menu) {
		menu.SetTitle(template.GetName())
		menu.AddAction("c", "Create", func() {
			item := model.CreateItem(template.GetId())
			item.SetContainerId(s.pc.GetId(), nil)
			s.WriteLine("Item created")
		})

		menu.AddAction("d", "Delete", func() {
			items := model.GetTemplateItems(template.GetId())
			if len(items) > 0 {
				if !s.getConfirmation(fmt.Sprintf("%v associated items will also be deleted, proceed? ", len(items))) {
					return
				}
			}
			model.DeleteTemplate(template.GetId())
			menu.Exit()
		})

		menu.AddAction("v", fmt.Sprintf("Value - %v", template.GetValue()), func() {
			value, valid := s.getInt("New value: ", 0, math.MaxInt32)
			if valid {
				template.SetValue(value)
			}
		})

		menu.AddAction("n", "Name", func() {
			name := s.getName("New name: ", types.TemplateType)
			if name != "" {
				template.SetName(name)
			}
		})

		menu.AddAction("w", fmt.Sprintf("Weight - %v", template.GetWeight()), func() {
			weight, valid := s.getInt("New weight: ", 0, math.MaxInt32)
			if valid {
				template.SetWeight(weight)
			}
		})

		menu.AddAction("a", fmt.Sprintf("Capacity - %v", template.GetCapacity()), func() {
			capacity, valid := s.getInt("New capacity: ", 0, math.MaxInt32)
			if valid {
				template.SetCapacity(capacity)
			}
		})
	})
}

func toggleExitMenu(s *Session) {
	onOrOff := func(direction types.Direction) string {
		text := "Off"
		if s.GetRoom().HasExit(direction) {
			text = "On"
		}
		return types.Colorize(types.ColorBlue, text)
	}

	toggleExit := func(direction types.Direction, menu *utils.Menu) func() {
		return func() {
			room := s.GetRoom()

			enable := !room.HasExit(direction)
			room.SetExitEnabled(direction, enable)

			loc := room.NextLocation(direction)
			otherRoom := model.GetRoomByLocation(loc, room.GetZoneId())
			if otherRoom != nil {
				otherRoom.SetExitEnabled(direction.Opposite(), enable)
			}
			menu.Exit()
		}
	}

	s.execMenu("Edit Exits", func(menu *utils.Menu) {
		menu.AddAction("n", "North: "+onOrOff(types.DirectionNorth), toggleExit(types.DirectionNorth, menu))
		menu.AddAction("ne", "North East: "+onOrOff(types.DirectionNorthEast), toggleExit(types.DirectionNorthEast, menu))
		menu.AddAction("e", "East: "+onOrOff(types.DirectionEast), toggleExit(types.DirectionEast, menu))
		menu.AddAction("se", "South East: "+onOrOff(types.DirectionSouthEast), toggleExit(types.DirectionSouthEast, menu))
		menu.AddAction("s", "South: "+onOrOff(types.DirectionSouth), toggleExit(types.DirectionSouth, menu))
		menu.AddAction("sw", "South West: "+onOrOff(types.DirectionSouthWest), toggleExit(types.DirectionSouthWest, menu))
		menu.AddAction("w", "West: "+onOrOff(types.DirectionWest), toggleExit(types.DirectionWest, menu))
		menu.AddAction("nw", "North West: "+onOrOff(types.DirectionNorthWest), toggleExit(types.DirectionNorthWest, menu))
		menu.AddAction("u", "Up: "+onOrOff(types.DirectionUp), toggleExit(types.DirectionUp, menu))
		menu.AddAction("d", "Down: "+onOrOff(types.DirectionDown), toggleExit(types.DirectionDown, menu))
	})
}

var linkData struct {
	source types.Id
	mode   string
}

const LinkSingle = "Single"
const LinkDouble = "Double"

func cAlias(name string) *command {
	return &command{alias: name}
}

func quickRoom(s *Session, command string) {
	dir := types.StringToDirection(command)

	if dir == types.DirectionNone {
		return
	}

	room := s.GetRoom()

	room.SetExitEnabled(dir, true)
	s.handleAction(command, "")
	s.GetRoom().SetExitEnabled(dir.Opposite(), true)
}

func whisper(self *command, s *Session, arg string) {
	name, message := utils.Argify(arg)
	if message == "" {
		self.Usage(s)
	} else {
		targetChar := model.GetPlayerCharacterByName(name)

		if targetChar != nil && targetChar.IsOnline() {
			model.Tell(s.pc, targetChar, message)
		} else {
			s.printError("Player '%s' not found", name)
		}
	}
}

func (s *Session) specificAreaMenu(area types.Area) {
	s.execMenu(area.GetName(), func(menu *utils.Menu) {
		menu.AddAction("r", "Rename", func() {
			name := s.getName("New name: ", types.AreaType)
			if name != "" {
				area.SetName(name)
			}
		})
		menu.AddAction("d", "Delete", func() {
			answer := s.getRawUserInput("Are you sure? ")

			if strings.ToLower(answer) == "y" {
				model.DeleteArea(area.GetId())
			}
			menu.Exit()
		})
		menu.AddAction("s", "Spawners", func() {
			s.spawnerMenu(area)
		})
	})
}

func (s *Session) spawnerMenu(area types.Area) {
	s.execMenu("Spawners", func(menu *utils.Menu) {
		for i, spawner := range model.GetAreaSpawners(area.GetId()) {
			sp := spawner
			menu.AddActionI(i, spawner.GetName(), func() {
				s.specificSpawnerMenu(sp)
			})
		}

		menu.AddAction("n", "New", func() {
			name := s.getName("Name of spawned NPC: ", types.SpawnerType)
			if name != "" {
				model.CreateSpawner(name, area.GetId())
			}
		})
	})
}

func (s *Session) specificSpawnerMenu(spawner types.Spawner) {
	s.execMenu(fmt.Sprintf("Spawner - %s", spawner.GetName()), func(menu *utils.Menu) {
		menu.AddAction("r", "Rename", func() {
			name := s.getName("New name: ", types.SpawnerType)
			if name != "" {
				spawner.SetName(name)
			}
		})

		menu.AddAction("c", fmt.Sprintf("Count - %v", spawner.GetCount()), func() {
			count, valid := s.getInt("New count: ", 0, 1000)
			if valid {
				spawner.SetCount(count)
			}
		})

		menu.AddAction("h", fmt.Sprintf("Health - %v", spawner.GetHealth()), func() {
			health, valid := s.getInt("New hitpoint count: ", 0, 1000)
			if valid {
				spawner.SetHealth(health)
			}
		})
	})
}

func pickEffect(s *Session) types.Effect {
	var chosenEffect types.Effect

	s.execMenu("Effects", func(menu *utils.Menu) {
		for i, effect := range model.GetAllEffects() {
			e := effect
			menu.AddActionI(i, effect.GetName(), func() {
				chosenEffect = e
				menu.Exit()
			})
		}
	})

	return chosenEffect
}

func (s *Session) getName(prompt string, objectType types.ObjectType) string {
	for {
		name := s.getCleanUserInput(prompt)
		if name == "" {
			return ""
		}

		id := model.FindObjectByName(name, objectType)
		if id == nil {
			return name
		} else if err := utils.ValidateName(name); err != nil {
			s.printError(err.Error())
		} else {
			s.printError("That name is unavailable")
		}
	}
}
