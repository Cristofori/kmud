package session

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type commandHandler struct {
	session *Session
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

func (ch *commandHandler) handleCommand(command string, args []string) {
	if command[0] == '/' {
		ch.quickRoom(command[1:])
		return
	}

	found := utils.FindAndCallMethod(ch, command, args)

	if !found {
		ch.session.printError("Unrecognized command: %s", command)
	}
}

func (ch *commandHandler) quickRoom(command string) {
	dir := types.StringToDirection(command)

	if dir == types.DirectionNone {
		return
	}

	ch.session.room.SetExitEnabled(dir, true)
	ch.session.actioner.handleAction(command, []string{})
	ch.session.room.SetExitEnabled(dir.Opposite(), true)
}

func (ch *commandHandler) Loc(args []string) {
	ch.Location(args)
}

func (ch *commandHandler) Location(args []string) {
	ch.session.printLine("%v", ch.session.room.GetLocation())
}

func (ch *commandHandler) Room(args []string) {
	for {
		menu := utils.NewMenu("Room")

		menu.AddAction("t", fmt.Sprintf("Title - %s", ch.session.room.GetTitle()))
		menu.AddAction("d", "Description")
		menu.AddAction("e", "Exits")

		area := model.GetArea(ch.session.room.GetAreaId())
		name := "(None)"
		if area != nil {
			name = area.GetName()
		}
		menu.AddAction("a", fmt.Sprintf("Area - %s", name))

		choice, _ := ch.session.execMenu(menu)

		switch choice {
		case "":
			ch.session.printRoom()
			return

		case "t":
			title := ch.session.getUserInput(RawUserInput, "Enter new title: ")

			if title != "" {
				ch.session.room.SetTitle(title)
			}

		case "d":
			description := ch.session.getUserInput(RawUserInput, "Enter new description: ")

			if description != "" {
				ch.session.room.SetDescription(description)
			}

		case "e":
			for {
				menu := toggleExitMenu(ch.session.room)

				choice, _ := ch.session.execMenu(menu)

				if choice == "" {
					break
				}

				direction := types.StringToDirection(choice)
				if direction != types.DirectionNone {
					enable := !ch.session.room.HasExit(direction)
					ch.session.room.SetExitEnabled(direction, enable)

					// Disable the corresponding exit in the adjacent room if necessary
					loc := ch.session.room.NextLocation(direction)
					otherRoom := model.GetRoomByLocation(loc, ch.session.room.GetZoneId())
					if otherRoom != nil {
						otherRoom.SetExitEnabled(direction.Opposite(), enable)
					}
				}
			}
		case "a":
			menu := utils.NewMenu("Change Area")
			menu.AddAction("n", "None")
			for i, area := range model.GetAreas(ch.session.currentZone()) {
				index := i + 1
				actionText := area.GetName()
				if area.GetId() == ch.session.room.GetAreaId() {
					actionText += "*"
				}
				menu.AddActionData(index, actionText, area.GetId())
			}

			choice, areaId := ch.session.execMenu(menu)

			switch choice {
			case "n":
				ch.session.room.SetAreaId(nil)
			default:
				ch.session.room.SetAreaId(areaId)
			}
		}
	}
}

func (ch *commandHandler) Map(args []string) {
	zoneRooms := model.GetRoomsInZone(ch.session.currentZone())
	roomsByLocation := map[types.Coordinate]types.Room{}

	for _, room := range zoneRooms {
		roomsByLocation[room.GetLocation()] = room
	}

	width, height := ch.session.user.GetWindowSize()
	height /= 2
	width /= 2

	// width and height need to be odd numbers so that we keep the current location centered
	// and we don't go off the edge of the available space
	width += (width % 2) - 1
	height += (height % 2) - 1

	builder := newMapBuilder(width, height, 1)
	builder.setUserRoom(ch.session.room)
	center := ch.session.room.GetLocation()

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

	ch.session.printLine(utils.TrimEmptyRows(builder.toString()))
}

func (ch *commandHandler) Zone(args []string) {
	if len(args) == 0 {
		ch.session.printLine("Current zone: " + types.Colorize(types.ColorBlue, ch.session.currentZone().GetName()))
	} else if len(args) == 1 {
		if args[0] == "list" {
			ch.session.printLineColor(types.ColorBlue, "Zones")
			ch.session.printLineColor(types.ColorBlue, "-----")
			for _, zone := range model.GetZones() {
				ch.session.printLine(zone.GetName())
			}
		} else {
			ch.session.printError("Usage: /zone [list|rename <name>|new <name>]")
		}
	} else if len(args) == 2 {
		if args[0] == "rename" {
			zone := model.GetZoneByName(args[0])

			if zone != nil {
				ch.session.printError("A zone with that name already exists")
				return
			}

			ch.session.currentZone().SetName(args[1])
		} else if args[0] == "new" {
			newZone, err := model.CreateZone(args[1])

			if err != nil {
				ch.session.printError(err.Error())
				return
			}

			newRoom, err := model.CreateRoom(newZone, types.Coordinate{X: 0, Y: 0, Z: 0})
			utils.HandleError(err)

			model.MoveCharacterToRoom(ch.session.player, newRoom)

			ch.session.room = newRoom
			ch.session.printRoom()
		}
	}
}

func (ch *commandHandler) B(args []string) {
	ch.Broadcast(args)
}

func (ch *commandHandler) Broadcast(args []string) {
	if len(args) == 0 {
		ch.session.printError("Nothing to say")
	} else {
		model.BroadcastMessage(ch.session.player, strings.Join(args, " "))
	}
}

func (ch *commandHandler) S(args []string) {
	ch.Say(args)
}

func (ch *commandHandler) Say(args []string) {
	if len(args) == 0 {
		ch.session.printError("Nothing to say")
	} else {
		model.Say(ch.session.player, strings.Join(args, " "))
	}
}

func (ch *commandHandler) Me(args []string) {
	model.Emote(ch.session.player, strings.Join(args, " "))
}

func (ch *commandHandler) W(args []string) {
	ch.Whisper(args)
}

func (ch *commandHandler) Tell(args []string) {
	ch.Whisper(args)
}

func (ch *commandHandler) Whisper(args []string) {
	if len(args) < 2 {
		ch.session.printError("Usage: /whisper <player> <message>")
		return
	}

	name := string(args[0])
	targetChar := model.GetPlayerCharacterByName(name)

	if targetChar == nil || !targetChar.IsOnline() {
		ch.session.printError("Player '%s' not found", name)
		return
	}

	message := strings.Join(args[1:], " ")
	model.Tell(ch.session.player, targetChar, message)
}

func (ch *commandHandler) Tp(args []string) {
	ch.Teleport(args)
}

func (ch *commandHandler) Teleport(args []string) {
	telUsage := func() {
		ch.session.printError("Usage: /teleport [<zone>|<X> <Y> <Z>]")
	}

	x := 0
	y := 0
	z := 0

	newZone := ch.session.currentZone()

	if len(args) == 1 {
		newZone = model.GetZoneByName(args[0])

		if newZone == nil {
			ch.session.printError("Zone not found")
			return
		}

		if newZone.GetId() == ch.session.room.GetZoneId() {
			ch.session.printLine("You're already in that zone")
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

	newRoom, err := model.MoveCharacterToLocation(ch.session.player, newZone, types.Coordinate{X: x, Y: y, Z: z})

	if err == nil {
		ch.session.room = newRoom
		ch.session.printRoom()
	} else {
		ch.session.printError(err.Error())
	}
}

func (ch *commandHandler) Who(args []string) {
	chars := model.GetOnlinePlayerCharacters()

	ch.session.printLine("")
	ch.session.printLine("Online Players")
	ch.session.printLine("--------------")

	for _, char := range chars {
		ch.session.printLine(char.GetName())
	}
	ch.session.printLine("")
}

func (ch *commandHandler) Colors(args []string) {
	ch.session.printLineColor(types.ColorNormal, "Normal")
	ch.session.printLineColor(types.ColorRed, "Red")
	ch.session.printLineColor(types.ColorDarkRed, "Dark Red")
	ch.session.printLineColor(types.ColorGreen, "Green")
	ch.session.printLineColor(types.ColorDarkGreen, "Dark Green")
	ch.session.printLineColor(types.ColorBlue, "Blue")
	ch.session.printLineColor(types.ColorDarkBlue, "Dark Blue")
	ch.session.printLineColor(types.ColorYellow, "Yellow")
	ch.session.printLineColor(types.ColorDarkYellow, "Dark Yellow")
	ch.session.printLineColor(types.ColorMagenta, "Magenta")
	ch.session.printLineColor(types.ColorDarkMagenta, "Dark Magenta")
	ch.session.printLineColor(types.ColorCyan, "Cyan")
	ch.session.printLineColor(types.ColorDarkCyan, "Dark Cyan")
	ch.session.printLineColor(types.ColorBlack, "Black")
	ch.session.printLineColor(types.ColorWhite, "White")
	ch.session.printLineColor(types.ColorGray, "Gray")
}

func (ch *commandHandler) CM(args []string) {
	ch.ColorMode(args)
}

func (ch *commandHandler) ColorMode(args []string) {
	if len(args) == 0 {
		message := "Current color mode is: "
		switch ch.session.user.GetColorMode() {
		case types.ColorModeNone:
			message = message + "None"
		case types.ColorModeLight:
			message = message + "Light"
		case types.ColorModeDark:
			message = message + "Dark"
		}
		ch.session.printLine(message)
	} else if len(args) == 1 {
		switch strings.ToLower(args[0]) {
		case "none":
			ch.session.user.SetColorMode(types.ColorModeNone)
			ch.session.printLine("Color mode set to: None")
		case "light":
			ch.session.user.SetColorMode(types.ColorModeLight)
			ch.session.printLine("Color mode set to: Light")
		case "dark":
			ch.session.user.SetColorMode(types.ColorModeDark)
			ch.session.printLine("Color mode set to: Dark")
		default:
			ch.session.printLine("Valid color modes are: None, Light, Dark")
		}
	} else {
		ch.session.printLine("Valid color modes are: None, Light, Dark")
	}
}

func (ch *commandHandler) DR(args []string) {
	ch.DestroyRoom(args)
}

func (ch *commandHandler) DestroyRoom(args []string) {
	if len(args) == 1 {
		direction := types.StringToDirection(args[0])

		if direction == types.DirectionNone {
			ch.session.printError("Not a valid direction")
		} else {
			loc := ch.session.room.NextLocation(direction)
			roomToDelete := model.GetRoomByLocation(loc, ch.session.room.GetZoneId())
			if roomToDelete != nil {
				model.DeleteRoom(roomToDelete)
				ch.session.printLine("Room destroyed")
			} else {
				ch.session.printError("No room in that direction")
			}
		}
	} else {
		ch.session.printError("Usage: /destroyroom <direction>")
	}
}

func getNpcName(ch *commandHandler) string {
	name := ""
	for {
		name = ch.session.getUserInput(CleanUserInput, "Desired NPC name: ")
		char := model.GetNpcByName(name)

		if name == "" {
			return ""
		} else if char != nil {
			ch.session.printError("That name is unavailable")
		} else if err := utils.ValidateName(name); err != nil {
			ch.session.printError(err.Error())
		} else {
			break
		}
	}
	return name
}

func (ch *commandHandler) Npc(args []string) {
	for {
		choice, npcId := ch.session.execMenu(npcMenu(nil))
		if choice == "" {
			break
		} else if choice == "n" {
			name := getNpcName(ch)
			if name != "" {
				model.CreateNpc(name, ch.session.room.GetId(), nil)
			}
		} else if npcId != nil {
			for {
				specificMenu := specificNpcMenu(npcId)
				choice, _ := ch.session.execMenu(specificMenu)
				npc := model.GetNpc(npcId)

				if choice == "d" {
					model.DeleteCharacter(npcId)
				} else if choice == "r" {
					name := getNpcName(ch)
					if name != "" {
						npc.SetName(name)
					}
				} else if choice == "c" {
					conversation := npc.GetConversation()

					if conversation == "" {
						conversation = "<empty>"
					}

					ch.session.printLine("Conversation: %s", conversation)
					newConversation := ch.session.getUserInput(RawUserInput, "New conversation text: ")

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

	ch.session.printRoom()
}

func (ch *commandHandler) Create(args []string) {
	createUsage := func() {
		ch.session.printError("Usage: /create <item name>")
	}

	if len(args) != 1 {
		createUsage()
		return
	}

	item := model.CreateItem(args[0])
	ch.session.room.AddItem(item.GetId())
	ch.session.printLine("Item created")
}

func (ch *commandHandler) DestroyItem(args []string) {
	destroyUsage := func() {
		ch.session.printError("Usage: /destroyitem <item name>")
	}

	if len(args) != 1 {
		destroyUsage()
		return
	}

	itemsInRoom := model.GetItems(ch.session.room.GetItems())
	name := strings.ToLower(args[0])

	for _, item := range itemsInRoom {
		if strings.ToLower(item.GetName()) == name {
			ch.session.room.RemoveItem(item.GetId())
			model.DeleteItem(item.GetId())
			ch.session.printLine("Item destroyed")
			return
		}
	}

	ch.session.printError("Item not found")
}

func (ch *commandHandler) RoomID(args []string) {
	ch.session.printLine("Room ID: %v", ch.session.room.GetId())
}

func (ch *commandHandler) Cash(args []string) {
	cashUsage := func() {
		ch.session.printError("Usage: /cash give <amount>")
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

		ch.session.player.AddCash(amount)
		ch.session.printLine("Received: %v monies", amount)
	} else {
		cashUsage()
		return
	}
}

func (ch *commandHandler) WS(args []string) { // WindowSize
	width, height := ch.session.user.GetWindowSize()

	header := fmt.Sprintf("Width: %v, Height: %v", width, height)

	topBar := header + " " + strings.Repeat("-", int(width)-2-len(header)) + "+"
	bottomBar := "+" + strings.Repeat("-", int(width)-2) + "+"
	outline := "|" + strings.Repeat(" ", int(width)-2) + "|"

	ch.session.printLine(topBar)

	for i := 0; i < int(height)-3; i++ {
		ch.session.printLine(outline)
	}

	ch.session.printLine(bottomBar)
}

func (ch *commandHandler) TT(args []string) { // TerminalType
	ch.session.printLine("Terminal type: %s", ch.session.user.GetTerminalType())
}

func (ch *commandHandler) Silent(args []string) {
	usage := func() {
		ch.session.printError("Usage: /silent [on|off]")
	}

	if len(args) != 1 {
		usage()
	} else if args[0] == "on" {
		ch.session.silentMode = true
		ch.session.printLine("Silent mode ON")
	} else if args[0] == "off" {
		ch.session.silentMode = false
		ch.session.printLine("Silent mode OFF")
	} else {
		usage()
	}
}

func (ch *commandHandler) R(args []string) { // Reply
	targetChar := model.GetPlayerCharacter(ch.session.replyId)

	if targetChar == nil {
		ch.session.asyncMessage("No one to reply to")
	} else if len(args) > 0 {
		newArgs := make([]string, 1)
		newArgs[0] = targetChar.GetName()
		newArgs = append(newArgs, args...)
		ch.Whisper(newArgs)
	} else {
		prompt := "Reply to " + targetChar.GetName() + ": "
		input := ch.session.getUserInput(RawUserInput, prompt)

		if input != "" {
			newArgs := make([]string, 1)
			newArgs[0] = targetChar.GetName()
			newArgs = append(newArgs, input)
			ch.Whisper(newArgs)
		}
	}
}

func (ch *commandHandler) Prop(args []string) {
	props := ch.session.room.GetProperties()

	keyVals := []string{}

	for key, value := range props {
		keyVals = append(keyVals, fmt.Sprintf("%s=%s", key, value))
	}

	for _, line := range keyVals {
		ch.session.printLine(line)
	}
}

func (ch *commandHandler) SetProp(args []string) {
	if len(args) != 2 {
		ch.session.printError("Usage: /setprop <key> <value>")
		return
	}

	ch.session.room.SetProperty(args[0], args[1])
}

func (ch *commandHandler) DelProp(args []string) {
	if len(args) != 1 {
		ch.session.printError("Usage: /delprop <key>")
	}

	ch.session.room.RemoveProperty(args[0])
}

func (ch *commandHandler) Area(args []string) {
	for {
		menu := utils.NewMenu("Areas")

		menu.AddAction("n", "New")

		for i, area := range model.GetAreas(ch.session.currentZone()) {
			menu.AddActionData(i+1, area.GetName(), area.GetId())
		}

		choice, areaId := ch.session.execMenu(menu)

		switch choice {
		case "":
			return
		case "n":
			name := ch.session.getRawUserInput("Area name: ")

			if name != "" {
				model.CreateArea(name, ch.session.currentZone())
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

					choice, _ = ch.session.execMenu(areaMenu)

					switch choice {
					case "":
						break AreaMenu
					case "r":
						newName := ch.session.getRawUserInput("New name: ")

						if newName != "" {
							area.SetName(newName)
						}
					case "d":
						answer := ch.session.getRawUserInput("Are you sure? ")

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

							choice, spawnerId := ch.session.execMenu(menu)

							switch choice {
							case "":
								break SpawnerMenu
							case "n":
								name := ch.session.getRawUserInput("Name of spawned NPC: ")

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

									choice, _ := ch.session.execMenu(menu)

									switch choice {
									case "":
										break SingleSpawnerMenu
									case "r":
										newName := ch.session.getRawUserInput("New name: ")
										if newName != "" {
											spawner.SetName(newName)
										}
									case "c":
										input := ch.session.getRawUserInput("New count: ")
										if input != "" {
											newCount, err := strconv.ParseInt(input, 10, 0)

											if err != nil || newCount < 0 {
												ch.session.printError("Invalid value")
											} else {
												spawner.SetCount(int(newCount))
											}
										}
									case "h":
										input := ch.session.getRawUserInput("New hitpoint count: ")
										if input != "" {
											newCount, err := strconv.ParseInt(input, 10, 0)

											if err != nil || newCount <= 0 {
												ch.session.printError("Invalid value")
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
				ch.session.printError("That area doesn't exist")
			}
		}
	}
}

var linkData struct {
	source types.Id
	mode   string
}

const LinkSingle = "Single"
const LinkDouble = "Double"

func (ch *commandHandler) Link(args []string) {
	StateName := "Linking"

	usage := func() {
		ch.session.printError("Usage: /link <name> [single|double*] to start, /link to finish, /link remove <name> [single|double*], /link rename <old name> <new name>, /link cancel")
	}

	linkName, linking := ch.session.states[StateName]

	if linking {
		if len(args) == 1 && args[0] == "cancel" {
			linkData.source = nil
			delete(ch.session.states, StateName)
		} else if len(args) != 0 {
			usage()
		} else {
			sourceRoom := model.GetRoom(linkData.source)

			sourceRoom.SetLink(linkName, ch.session.room.GetId())
			if linkData.mode == LinkDouble {
				ch.session.room.SetLink(linkName, linkData.source)
			}

			linkData.source = nil
			delete(ch.session.states, StateName)

			ch.session.printRoom()
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

			linkNames := ch.session.room.LinkNames()
			index := utils.BestMatch(args[1], linkNames)

			if index == -2 {
				ch.session.printError("Which one do you mean?")
			} else if index == -1 {
				ch.session.printError("Link not found")
			} else {
				linkName := linkNames[index]

				if mode == "double" {
					links := ch.session.room.GetLinks()
					linkedRoom := model.GetRoom(links[linkName])
					linkedRoom.RemoveLink(linkName)
				}

				ch.session.room.RemoveLink(linkName)
				ch.session.printRoom()
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
			ch.session.states[StateName] = utils.FormatName(args[0])
			linkData.source = ch.session.room.GetId()
		}
	}
}

func (ch *commandHandler) Kill(args []string) {
	if len(args) != 1 {
		ch.session.printError("Usage: /kill [npc name]")
		return
	}

	npcs := model.NpcsIn(ch.session.room.GetId())
	index := utils.BestMatch(args[0], npcs.Characters().Names())

	if index == -1 {
		ch.session.printError("Not found")
	} else if index == -2 {
		ch.session.printError("Which one do you mean?")
	} else {
		npc := npcs[index]
		npc.SetHitPoints(0)
		ch.session.printLine("Killed %s", npc.GetName())
	}
}

func (ch *commandHandler) Inspect(args []string) {
	if len(args) != 1 {
		ch.session.printError("Usage: /inspect [name]")
		return
	}

	characters := model.CharactersIn(ch.session.room.GetId())
	index := utils.BestMatch(args[0], characters.Names())

	if index == -1 {
		ch.session.printError("Not found")
	} else if index == -2 {
		ch.session.printError("Which one do you mean?")
	} else {
		char := characters[index]

		ch.session.printLine(char.GetName())
		ch.session.printLine("Health: %v/%v", char.GetHitPoints(), char.GetHealth())
	}
}

// vim: nocindent
