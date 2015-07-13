package session

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type command struct {
	admin bool
	alias string
	exec  func(*Session, []string)
}

func npcMenu(roomId types.Id) *utils.Menu {
	var npcs types.NPCList

	if roomId != nil {
		npcs = model.NpcsIn(roomId)
	} else {
		npcs = model.GetNpcs()
	}

	menu := utils.NewMenu("NPCs")

	menu.AddAction("n", "New")

	for i, npc := range npcs {
		index := i + 1
		menu.AddActionData(index, npc.GetName(), npc.GetId())
	}

	return menu
}

func specificNpcMenu(npcId types.Id) *utils.Menu {
	npc := model.GetNpc(npcId)
	menu := utils.NewMenu(npc.GetName())
	menu.AddAction("r", "Rename")
	menu.AddAction("d", "Delete")
	menu.AddAction("c", "Conversation")

	roamingState := "Off"
	if npc.GetRoaming() {
		roamingState = "On"
	}

	menu.AddAction("o", fmt.Sprintf("Roaming - %s", roamingState))
	return menu
}

func toggleExitMenu(room types.Room) *utils.Menu {
	onOrOff := func(direction types.Direction) string {
		text := "Off"
		if room.HasExit(direction) {
			text = "On"
		}
		return types.Colorize(types.ColorBlue, text)
	}

	menu := utils.NewMenu("Edit Exits")

	menu.AddAction("n", "North: "+onOrOff(types.DirectionNorth))
	menu.AddAction("ne", "North East: "+onOrOff(types.DirectionNorthEast))
	menu.AddAction("e", "East: "+onOrOff(types.DirectionEast))
	menu.AddAction("se", "South East: "+onOrOff(types.DirectionSouthEast))
	menu.AddAction("s", "South: "+onOrOff(types.DirectionSouth))
	menu.AddAction("sw", "South West: "+onOrOff(types.DirectionSouthWest))
	menu.AddAction("w", "West: "+onOrOff(types.DirectionWest))
	menu.AddAction("nw", "North West: "+onOrOff(types.DirectionNorthWest))
	menu.AddAction("u", "Up: "+onOrOff(types.DirectionUp))
	menu.AddAction("d", "Down: "+onOrOff(types.DirectionDown))

	return menu
}

var commands map[string]command

func (self *Session) handleCommand(command string, args []string) {
	if commands == nil {
		initCommands()
	}

	if command[0] == '/' && self.user.IsAdmin() {
		quickRoom(self, command[1:])
		return
	}

	handler, found := commands[command]

	if found {
		if handler.alias != "" {
			handler = commands[handler.alias]
		}
		handler.exec(self, args)
	} else {
		self.printError("Unrecognized command: %s", command)
	}
}

func initCommands() {
	commands = map[string]command{
		"loc": command{alias: "location"},
		"location": command{
			admin: false,
			exec: func(s *Session, args []string) {
				s.printLine("%v", s.room.GetLocation())
			},
		},
		"room": command{
			admin: true,
			exec: func(s *Session, args []string) {
				for {
					menu := utils.NewMenu("Room")

					menu.AddAction("t", fmt.Sprintf("Title - %s", s.room.GetTitle()))
					menu.AddAction("d", "Description")
					menu.AddAction("e", "Exits")

					area := model.GetArea(s.room.GetAreaId())
					name := "(None)"
					if area != nil {
						name = area.GetName()
					}
					menu.AddAction("a", fmt.Sprintf("Area - %s", name))

					choice, _ := s.execMenu(menu)

					switch choice {
					case "":
						s.printRoom()
						return

					case "t":
						title := s.getUserInput(RawUserInput, "Enter new title: ")

						if title != "" {
							s.room.SetTitle(title)
						}

					case "d":
						description := s.getUserInput(RawUserInput, "Enter new description: ")

						if description != "" {
							s.room.SetDescription(description)
						}

					case "e":
						for {
							menu := toggleExitMenu(s.room)

							choice, _ := s.execMenu(menu)

							if choice == "" {
								break
							}

							direction := types.StringToDirection(choice)
							if direction != types.DirectionNone {
								enable := !s.room.HasExit(direction)
								s.room.SetExitEnabled(direction, enable)

								// Disable the corresponding exit in the adjacent room if necessary
								loc := s.room.NextLocation(direction)
								otherRoom := model.GetRoomByLocation(loc, s.room.GetZoneId())
								if otherRoom != nil {
									otherRoom.SetExitEnabled(direction.Opposite(), enable)
								}
							}
						}
					case "a":
						menu := utils.NewMenu("Change Area")
						menu.AddAction("n", "None")
						for i, area := range model.GetAreas(s.currentZone()) {
							index := i + 1
							actionText := area.GetName()
							if area.GetId() == s.room.GetAreaId() {
								actionText += "*"
							}
							menu.AddActionData(index, actionText, area.GetId())
						}

						choice, areaId := s.execMenu(menu)

						switch choice {
						case "n":
							s.room.SetAreaId(nil)
						default:
							s.room.SetAreaId(areaId)
						}
					}
				}
			},
		},
		"map": command{
			admin: false,
			exec: func(s *Session, args []string) {
				zoneRooms := model.GetRoomsInZone(s.currentZone())
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
				builder.setUserRoom(s.room)
				center := s.room.GetLocation()

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

				s.printLine(utils.TrimEmptyRows(builder.toString()))
			},
		},
		"zone": command{
			admin: true,
			exec: func(s *Session, args []string) {
				if len(args) == 0 {
					s.printLine("Current zone: " + types.Colorize(types.ColorBlue, s.currentZone().GetName()))
				} else if len(args) == 1 {
					if args[0] == "list" {
						s.printLineColor(types.ColorBlue, "Zones")
						s.printLineColor(types.ColorBlue, "-----")
						for _, zone := range model.GetZones() {
							s.printLine(zone.GetName())
						}
					} else {
						s.printError("Usage: /zone [list|rename <name>|new <name>]")
					}
				} else if len(args) == 2 {
					if args[0] == "rename" {
						zone := model.GetZoneByName(args[0])

						if zone != nil {
							s.printError("A zone with that name already exists")
							return
						}

						s.currentZone().SetName(args[1])
					} else if args[0] == "new" {
						newZone, err := model.CreateZone(args[1])

						if err != nil {
							s.printError(err.Error())
							return
						}

						newRoom, err := model.CreateRoom(newZone, types.Coordinate{X: 0, Y: 0, Z: 0})
						utils.HandleError(err)

						model.MoveCharacterToRoom(s.player, newRoom)

						s.room = newRoom
						s.printRoom()
					}
				}
			},
		},
		"b": command{alias: "broadcast"},
		"broadcast": command{
			admin: false,
			exec: func(s *Session, args []string) {
				if len(args) == 0 {
					s.printError("Nothing to say")
				} else {
					model.BroadcastMessage(s.player, strings.Join(args, " "))
				}
			},
		},
		"s": alias("say"),
		"say": command{
			admin: false,
			exec: func(s *Session, args []string) {
				if len(args) == 0 {
					s.printError("Nothing to say")
				} else {
					model.Say(s.player, strings.Join(args, " "))
				}
			},
		},
		"me": command{
			admin: false,
			exec: func(s *Session, args []string) {
				model.Emote(s.player, strings.Join(args, " "))
			},
		},
		"w":    alias("whiser"),
		"tell": alias("whisper"),
		"whisper": command{
			admin: false,
			exec: func(s *Session, args []string) {
				if len(args) < 2 {
					s.printError("Usage: /whisper <player> <message>")
					return
				}

				name := string(args[0])
				targetChar := model.GetPlayerCharacterByName(name)

				if targetChar == nil || !targetChar.IsOnline() {
					s.printError("Player '%s' not found", name)
					return
				}

				message := strings.Join(args[1:], " ")
				model.Tell(s.player, targetChar, message)
			},
		},
		"tp": alias("teleport"),
		"teleport": command{
			admin: true,
			exec: func(s *Session, args []string) {
				telUsage := func() {
					s.printError("Usage: /teleport [<zone>|<X> <Y> <Z>]")
				}

				x := 0
				y := 0
				z := 0

				newZone := s.currentZone()

				if len(args) == 1 {
					newZone = model.GetZoneByName(args[0])

					if newZone == nil {
						s.printError("Zone not found")
						return
					}

					if newZone.GetId() == s.room.GetZoneId() {
						s.printLine("You're already in that zone")
						return
					}

					zoneRooms := model.GetRoomsInZone(newZone)

					if len(zoneRooms) > 0 {
						r := zoneRooms[0]
						x = r.GetLocation().X
						y = r.GetLocation().Y
						z = r.GetLocation().Z
					}
				} else if len(args) == 3 {
					var err error
					x, err = strconv.Atoi(args[0])

					if err != nil {
						telUsage()
						return
					}

					y, err = strconv.Atoi(args[1])

					if err != nil {
						telUsage()
						return
					}

					z, err = strconv.Atoi(args[2])

					if err != nil {
						telUsage()
						return
					}
				} else {
					telUsage()
					return
				}

				newRoom, err := model.MoveCharacterToLocation(s.player, newZone, types.Coordinate{X: x, Y: y, Z: z})

				if err == nil {
					s.room = newRoom
					s.printRoom()
				} else {
					s.printError(err.Error())
				}
			},
		},
		"who": command{
			admin: false,
			exec: func(s *Session, args []string) {
				chars := model.GetOnlinePlayerCharacters()

				s.printLine("")
				s.printLine("Online Players")
				s.printLine("--------------")

				for _, char := range chars {
					s.printLine(char.GetName())
				}
				s.printLine("")
			},
		},
		"colors": command{
			admin: false,
			exec: func(s *Session, args []string) {
				s.printLineColor(types.ColorNormal, "Normal")
				s.printLineColor(types.ColorRed, "Red")
				s.printLineColor(types.ColorDarkRed, "Dark Red")
				s.printLineColor(types.ColorGreen, "Green")
				s.printLineColor(types.ColorDarkGreen, "Dark Green")
				s.printLineColor(types.ColorBlue, "Blue")
				s.printLineColor(types.ColorDarkBlue, "Dark Blue")
				s.printLineColor(types.ColorYellow, "Yellow")
				s.printLineColor(types.ColorDarkYellow, "Dark Yellow")
				s.printLineColor(types.ColorMagenta, "Magenta")
				s.printLineColor(types.ColorDarkMagenta, "Dark Magenta")
				s.printLineColor(types.ColorCyan, "Cyan")
				s.printLineColor(types.ColorDarkCyan, "Dark Cyan")
				s.printLineColor(types.ColorBlack, "Black")
				s.printLineColor(types.ColorWhite, "White")
				s.printLineColor(types.ColorGray, "Gray")
			},
		},
		"cm": alias("colormode"),
		"colormode": command{
			admin: false,
			exec: func(s *Session, args []string) {
				if len(args) == 0 {
					message := "Current color mode is: "
					switch s.user.GetColorMode() {
					case types.ColorModeNone:
						message = message + "None"
					case types.ColorModeLight:
						message = message + "Light"
					case types.ColorModeDark:
						message = message + "Dark"
					}
					s.printLine(message)
				} else if len(args) == 1 {
					switch strings.ToLower(args[0]) {
					case "none":
						s.user.SetColorMode(types.ColorModeNone)
						s.printLine("Color mode set to: None")
					case "light":
						s.user.SetColorMode(types.ColorModeLight)
						s.printLine("Color mode set to: Light")
					case "dark":
						s.user.SetColorMode(types.ColorModeDark)
						s.printLine("Color mode set to: Dark")
					default:
						s.printLine("Valid color modes are: None, Light, Dark")
					}
				} else {
					s.printLine("Valid color modes are: None, Light, Dark")
				}
			},
		},
		"dr": alias("destroyroom"),
		"destroyroom": command{
			admin: true,
			exec: func(s *Session, args []string) {
				if len(args) == 1 {
					direction := types.StringToDirection(args[0])

					if direction == types.DirectionNone {
						s.printError("Not a valid direction")
					} else {
						loc := s.room.NextLocation(direction)
						roomToDelete := model.GetRoomByLocation(loc, s.room.GetZoneId())
						if roomToDelete != nil {
							model.DeleteRoom(roomToDelete)
							s.printLine("Room destroyed")
						} else {
							s.printError("No room in that direction")
						}
					}
				} else {
					s.printError("Usage: /destroyroom <direction>")
				}
			},
		},
		"npc": command{
			admin: true,
			exec: func(s *Session, args []string) {
				for {
					choice, npcId := s.execMenu(npcMenu(nil))
					if choice == "" {
						break
					} else if choice == "n" {
						name := getNpcName(s)
						if name != "" {
							model.CreateNpc(name, s.room.GetId(), nil)
						}
					} else if npcId != nil {
						for {
							specificMenu := specificNpcMenu(npcId)
							choice, _ := s.execMenu(specificMenu)
							npc := model.GetNpc(npcId)

							if choice == "d" {
								model.DeleteCharacter(npcId)
							} else if choice == "r" {
								name := getNpcName(s)
								if name != "" {
									npc.SetName(name)
								}
							} else if choice == "c" {
								conversation := npc.GetConversation()

								if conversation == "" {
									conversation = "<empty>"
								}

								s.printLine("Conversation: %s", conversation)
								newConversation := s.getUserInput(RawUserInput, "New conversation text: ")

								if newConversation != "" {
									npc.SetConversation(newConversation)
								}
							} else if choice == "o" {
								npc.SetRoaming(!npc.GetRoaming())
							} else if choice == "" {
								break
							}
						}
					}
				}

				s.printRoom()
			},
		},
		"create": command{
			admin: true,
			exec: func(s *Session, args []string) {
				createUsage := func() {
					s.printError("Usage: /create <item name>")
				}

				if len(args) != 1 {
					createUsage()
					return
				}

				item := model.CreateItem(args[0])
				s.room.AddItem(item.GetId())
				s.printLine("Item created")
			},
		},
		"destroyitem": command{
			admin: true,
			exec: func(s *Session, args []string) {
				destroyUsage := func() {
					s.printError("Usage: /destroyitem <item name>")
				}

				if len(args) != 1 {
					destroyUsage()
					return
				}

				itemsInRoom := model.GetItems(s.room.GetItems())
				name := strings.ToLower(args[0])

				for _, item := range itemsInRoom {
					if strings.ToLower(item.GetName()) == name {
						s.room.RemoveItem(item.GetId())
						model.DeleteItem(item.GetId())
						s.printLine("Item destroyed")
						return
					}
				}

				s.printError("Item not found")
			},
		},
		"roomid": command{
			admin: true,
			exec: func(s *Session, args []string) {
				s.printLine("Room ID: %v", s.room.GetId())
			},
		},
		"cash": command{
			admin: false,
			exec: func(s *Session, args []string) {
				cashUsage := func() {
					s.printError("Usage: /cash give <amount>")
				}

				if len(args) != 2 {
					cashUsage()
					return
				}

				if args[0] == "give" {
					amount, err := strconv.Atoi(args[1])

					if err != nil {
						cashUsage()
						return
					}

					s.player.AddCash(amount)
					s.printLine("Received: %v monies", amount)
				} else {
					cashUsage()
					return
				}
			},
		},
		"ws": alias("windowsize"),
		"windowsize": command{
			admin: false,
			exec: func(s *Session, args []string) {
				width, height := s.user.GetWindowSize()

				header := fmt.Sprintf("Width: %v, Height: %v", width, height)

				topBar := header + " " + strings.Repeat("-", int(width)-2-len(header)) + "+"
				bottomBar := "+" + strings.Repeat("-", int(width)-2) + "+"
				outline := "|" + strings.Repeat(" ", int(width)-2) + "|"

				s.printLine(topBar)

				for i := 0; i < int(height)-3; i++ {
					s.printLine(outline)
				}

				s.printLine(bottomBar)
			},
		},
		"tt": alias("terminaltype"),
		"terminaltype": command{
			admin: false,
			exec: func(s *Session, args []string) {
				s.printLine("Terminal type: %s", s.user.GetTerminalType())
			},
		},
		"silent": command{
			admin: false,
			exec: func(s *Session, args []string) {
				usage := func() {
					s.printError("Usage: /silent [on|off]")
				}

				if len(args) != 1 {
					usage()
				} else if args[0] == "on" {
					s.silentMode = true
					s.printLine("Silent mode ON")
				} else if args[0] == "off" {
					s.silentMode = false
					s.printLine("Silent mode OFF")
				} else {
					usage()
				}
			},
		},
		"r": alias("reply"),
		"reply": command{
			admin: false,
			exec: func(s *Session, args []string) {
				targetChar := model.GetPlayerCharacter(s.replyId)

				if targetChar == nil {
					s.asyncMessage("No one to reply to")
				} else if len(args) > 0 {
					newArgs := make([]string, 1)
					newArgs[0] = targetChar.GetName()
					newArgs = append(newArgs, args...)
					s.handleCommand("whisper", newArgs)
				} else {
					prompt := "Reply to " + targetChar.GetName() + ": "
					input := s.getUserInput(RawUserInput, prompt)

					if input != "" {
						newArgs := make([]string, 1)
						newArgs[0] = targetChar.GetName()
						newArgs = append(newArgs, input)
						s.handleCommand("whisper", newArgs)
					}
				}
			},
		},
		"area": command{
			admin: true,
			exec: func(s *Session, args []string) {
				for {
					menu := utils.NewMenu("Areas")

					menu.AddAction("n", "New")

					for i, area := range model.GetAreas(s.currentZone()) {
						menu.AddActionData(i+1, area.GetName(), area.GetId())
					}

					choice, areaId := s.execMenu(menu)

					switch choice {
					case "":
						return
					case "n":
						name := s.getRawUserInput("Area name: ")

						if name != "" {
							model.CreateArea(name, s.currentZone())
						}
					default:
						area := model.GetArea(areaId)

						if area != nil {
						AreaMenu:
							for {
								areaMenu := utils.NewMenu(area.GetName())
								areaMenu.AddAction("r", "Rename")
								areaMenu.AddAction("d", "Delete")
								areaMenu.AddAction("s", "Spawners")

								choice, _ = s.execMenu(areaMenu)

								switch choice {
								case "":
									break AreaMenu
								case "r":
									newName := s.getRawUserInput("New name: ")

									if newName != "" {
										area.SetName(newName)
									}
								case "d":
									answer := s.getRawUserInput("Are you sure? ")

									if strings.ToLower(answer) == "y" {
										model.DeleteArea(areaId)
									}
								case "s":
								SpawnerMenu:
									for {
										menu := utils.NewMenu("Spawners")

										for i, spawner := range model.GetAreaSpawners(areaId) {
											menu.AddActionData(i+1, spawner.GetName(), spawner.GetId())
										}

										menu.AddAction("n", "New")

										choice, spawnerId := s.execMenu(menu)

										switch choice {
										case "":
											break SpawnerMenu
										case "n":
											name := s.getRawUserInput("Name of spawned NPC: ")

											if name != "" {
												model.CreateSpawner(name, areaId)
											}
										default:
											spawner := model.GetSpawner(spawnerId)

										SingleSpawnerMenu:
											for {
												menu := utils.NewMenu(fmt.Sprintf("%s - %s", "Spawner", spawner.GetName()))

												menu.AddAction("r", "Rename")
												menu.AddAction("c", fmt.Sprintf("Count - %v", spawner.GetCount()))
												menu.AddAction("h", fmt.Sprintf("Hitpoints - %v", spawner.GetHitPoints()))

												choice, _ := s.execMenu(menu)

												switch choice {
												case "":
													break SingleSpawnerMenu
												case "r":
													newName := s.getRawUserInput("New name: ")
													if newName != "" {
														spawner.SetName(newName)
													}
												case "c":
													input := s.getRawUserInput("New count: ")
													if input != "" {
														newCount, err := strconv.ParseInt(input, 10, 0)

														if err != nil || newCount < 0 {
															s.printError("Invalid value")
														} else {
															spawner.SetCount(int(newCount))
														}
													}
												case "h":
													input := s.getRawUserInput("New hitpoint count: ")
													if input != "" {
														newCount, err := strconv.ParseInt(input, 10, 0)

														if err != nil || newCount <= 0 {
															s.printError("Invalid value")
														} else {
															spawner.SetHealth(int(newCount))
														}
													}
												}
											}
										}
									}
								}
							}
						} else {
							s.printError("That area doesn't exist")
						}
					}
				}
			},
		},
		"link": command{
			admin: true,
			exec: func(s *Session, args []string) {
				StateName := "Linking"

				usage := func() {
					s.printError("Usage: /link <name> [single|double*] to start, /link to finish, /link remove <name> [single|double*], /link rename <old name> <new name>, /link cancel")
				}

				linkName, linking := s.states[StateName]

				if linking {
					if len(args) == 1 && args[0] == "cancel" {
						linkData.source = nil
						delete(s.states, StateName)
					} else if len(args) != 0 {
						usage()
					} else {
						sourceRoom := model.GetRoom(linkData.source)

						sourceRoom.SetLink(linkName, s.room.GetId())
						if linkData.mode == LinkDouble {
							s.room.SetLink(linkName, linkData.source)
						}

						linkData.source = nil
						delete(s.states, StateName)

						s.printRoom()
					}
				} else {
					if len(args) == 0 {
						usage()
						return
					}

					if args[0] == "remove" {
						mode := "double"
						if len(args) == 3 {
							if args[2] == "single" || args[2] == "double" {
								mode = args[2]
							} else {
								usage()
								return
							}
						}

						if len(args) != 2 {
							usage()
							return
						}

						linkNames := s.room.LinkNames()
						index := utils.BestMatch(args[1], linkNames)

						if index == -2 {
							s.printError("Which one do you mean?")
						} else if index == -1 {
							s.printError("Link not found")
						} else {
							linkName := linkNames[index]

							if mode == "double" {
								links := s.room.GetLinks()
								linkedRoom := model.GetRoom(links[linkName])
								linkedRoom.RemoveLink(linkName)
							}

							s.room.RemoveLink(linkName)
							s.printRoom()
						}
					} else if args[0] == "rename" {
						// TODO
					} else {
						if len(args) == 2 {
							if args[1] == LinkSingle || args[1] == LinkDouble {
								linkData.mode = args[1]
							} else {
								usage()
								return
							}
						} else {
							linkData.mode = LinkDouble
						}

						// New link
						s.states[StateName] = utils.FormatName(args[0])
						linkData.source = s.room.GetId()
					}
				}
			},
		},
		"kill": command{
			admin: true,
			exec: func(s *Session, args []string) {
				if len(args) != 1 {
					s.printError("Usage: /kill [npc name]")
					return
				}

				npcs := model.NpcsIn(s.room.GetId())
				index := utils.BestMatch(args[0], npcs.Characters().Names())

				if index == -1 {
					s.printError("Not found")
				} else if index == -2 {
					s.printError("Which one do you mean?")
				} else {
					npc := npcs[index]
					npc.SetHitPoints(0)
					s.printLine("Killed %s", npc.GetName())
				}
			},
		},
		"inspect": command{
			admin: true,
			exec: func(s *Session, args []string) {
				if len(args) != 1 {
					s.printError("Usage: /inspect [name]")
					return
				}

				characters := model.CharactersIn(s.room.GetId())
				index := utils.BestMatch(args[0], characters.Names())

				if index == -1 {
					s.printError("Not found")
				} else if index == -2 {
					s.printError("Which one do you mean?")
				} else {
					char := characters[index]

					s.printLine(char.GetName())
					s.printLine("Health: %v/%v", char.GetHitPoints(), char.GetHealth())
				}
			},
		},
		"skills": command{
			admin: true,
			exec: func(s *Session, args []string) {
			Loop:
				for {
					menu := utils.NewMenu("Skills")

					menu.AddAction("n", "New")

					skills := model.GetSkills()

					for i, skill := range skills {
						menu.AddActionData(i+1, skill.GetName(), skill.GetId())
					}

					choice, skillId := s.execMenu(menu)

					switch choice {
					case "":
						break Loop
					case "n":
						name := s.getRawUserInput("Skill name: ")
						if name != "" {
							model.CreateSkill(name)
						}
					default:
						skill := model.GetSkill(skillId)
					SingleSkillMenu:
						for {
							menu := utils.NewMenu(fmt.Sprintf("Skill - %s", skill.GetName()))
							menu.AddAction("a", fmt.Sprintf("Damage - %v", skill.GetDamage()))
							menu.AddAction("d", "Delete")

							choice, _ := s.execMenu(menu)

							switch choice {
							case "a":
								input := s.getRawUserInput("New damage value: ")
								dmg, err := strconv.ParseInt(input, 10, 0)

								if err != nil || dmg < 0 {
									s.printError("Invalid value")
								} else {
									skill.SetDamage(int(dmg))
								}
							case "d":
								model.DeleteSkill(skillId)
								break SingleSkillMenu
							}
						}
					}
				}
			},
		},
	}
}

var linkData struct {
	source types.Id
	mode   string
}

const LinkSingle = "Single"
const LinkDouble = "Double"

func alias(name string) command {
	return command{alias: name}
}

func quickRoom(s *Session, command string) {
	dir := types.StringToDirection(command)

	if dir == types.DirectionNone {
		return
	}

	s.room.SetExitEnabled(dir, true)
	s.actioner.handleAction(command, []string{})
	s.room.SetExitEnabled(dir.Opposite(), true)
}

func getNpcName(s *Session) string {
	name := ""
	for {
		name = s.getUserInput(CleanUserInput, "Desired NPC name: ")
		char := model.GetNpcByName(name)

		if name == "" {
			return ""
		} else if char != nil {
			s.printError("That name is unavailable")
		} else if err := utils.ValidateName(name); err != nil {
			s.printError(err.Error())
		} else {
			break
		}
	}
	return name
}

// vim: nocindent
