package session

import (
	"fmt"
	"kmud/database"
	"kmud/model"
	"kmud/utils"
	"strconv"
	"strings"
)

func (session *Session) handleCommand(command string, args []string) {
	if command[0] == '/' {
		session.quickRoom(command[1:])
		return
	}

	switch command {
	case "edit":
		session.edit(args)

	case "loc":
		fallthrough
	case "location":
		session.printLine("%v", session.room.GetLocation())

	case "map":
		session.printMap(args)

	case "zone":
		session.zoneCommand(args)

	case "broadcast":
		fallthrough
	case "b":
		session.broadcast(args)

	case "say":
		fallthrough
	case "s":
		session.say(args)

	case "me":
		session.emote(args)

	case "whisper":
		fallthrough
	case "tell":
		fallthrough
	case "w":
		session.whisper(args)

	case "teleport":
		fallthrough
	case "tel":
		session.teleport(args)

	case "who":
		session.who()

	case "colors":
		session.colors()

	case "colormode":
		fallthrough
	case "cm":
		session.colorMode(args)

	case "destroyroom":
		session.destroyRoom(args)

	case "npc":
		session.npc(args)

	case "create":
		session.createItem(args)

	case "destroyitem":
		session.destroyItem(args)

	case "cash":
		session.cash(args)

	case "roomid":
		session.printLine("Room ID: %v", session.room.GetId())

	case "ws":
		session.windowSize()

	case "tt":
		session.terminalType()

	default:
		session.printError("Unrecognized command: %s", command)
	}
}

func (session *Session) quickRoom(command string) {
	dir := database.StringToDirection(command)

	if dir == database.DirectionNone {
		return
	}

	session.room.SetExitEnabled(dir, true)
	session.handleAction(command, []string{})
	session.room.SetExitEnabled(dir.Opposite(), true)
}

func (session *Session) edit(args []string) {
	session.printRoomEditor()

	for {
		input := session.getUserInput(CleanUserInput, "Select a section to edit: ")

		switch input {
		case "":
			session.printRoom()
			return

		case "1":
			input = session.getUserInput(RawUserInput, "Enter new title: ")

			if input != "" {
				session.room.SetTitle(input)
			}
			session.printRoomEditor()

		case "2":
			input = session.getUserInput(RawUserInput, "Enter new description: ")

			if input != "" {
				session.room.SetDescription(input)
			}
			session.printRoomEditor()

		case "3":
			for {
				menu := toggleExitMenu(session.user.GetColorMode(), session.room)

				choice, _ := session.execMenu(menu)

				if choice == "" {
					break
				}

				direction := database.StringToDirection(choice)
				if direction != database.DirectionNone {
					enable := !session.room.HasExit(direction)
					session.room.SetExitEnabled(direction, enable)

					// Disable the corresponding exit in the adjacent room if necessary
					loc := session.room.NextLocation(direction)
					otherRoom := model.M.GetRoomByLocation(loc, session.zone)
					if otherRoom != nil {
						otherRoom.SetExitEnabled(direction.Opposite(), enable)
					}
				}
			}

			session.printRoomEditor()

		default:
			session.printLine("Invalid selection")
		}
	}
}

func (session *Session) printMap(args []string) {
	mapUsage := func() {
		session.printError("Usage: /map [all]")
	}

	startX := 0
	startY := 0
	startZ := 0
	endX := 0
	endY := 0
	endZ := 0

	if len(args) == 0 {
		width, height := session.user.WindowSize()

		loc := session.room.GetLocation()

		startX = loc.X - (width / 4)
		startY = loc.Y - (height / 4)
		startZ = loc.Z

		endX = loc.X + (width / 4)
		endY = loc.Y + (height / 4)
		endZ = loc.Z
	} else if args[0] == "all" {
		topLeft, bottomRight := model.ZoneCorners(session.zone.GetId())

		startX = topLeft.X
		startY = topLeft.Y
		startZ = topLeft.Z
		endX = bottomRight.X
		endY = bottomRight.Y
		endZ = bottomRight.Z
	} else {
		mapUsage()
		return
	}

	width := endX - startX + 1
	height := endY - startY + 1
	depth := endZ - startZ + 1

	builder := newMapBuilder(width, height, depth)
	builder.setUserRoom(session.room)

	for z := startZ; z <= endZ; z++ {
		for y := startY; y <= endY; y++ {
			for x := startX; x <= endX; x++ {
				loc := database.Coordinate{x, y, z}
				room := model.M.GetRoomByLocation(loc, session.zone)

				if room != nil {
					// Translate to 0-based coordinates and double the coordinate
					// space to leave room for the exit lines
					builder.addRoom(room, (x-startX)*2, (y-startY)*2, z-startZ)
				}
			}
		}
	}

	session.printLine(utils.TrimEmptyRows(builder.toString(session.user.GetColorMode())))
}

func (session *Session) zoneCommand(args []string) {
	if len(args) == 0 {
		if session.zone.GetId() == "" {
			session.printLine("Currently in the null zone")
		} else {
			session.printLine("Current zone: " + utils.Colorize(session.user.GetColorMode(), utils.ColorBlue, session.zone.GetName()))
		}
	} else if len(args) == 1 {
		if args[0] == "list" {
			session.printLineColor(utils.ColorBlue, "Zones")
			session.printLineColor(utils.ColorBlue, "-----")
			for _, zone := range model.M.GetZones() {
				session.printLine(zone.GetName())
			}
		} else {
			session.printError("Usage: /zone [list|rename <name>|new <name>]")
		}
	} else if len(args) == 2 {
		if args[0] == "rename" {
			zone := model.M.GetZoneByName(args[0])

			if zone != nil {
				session.printError("A zone with that name already exists")
				return
			}

			if session.zone.GetId() == "" {
				session.zone = model.M.CreateZone(args[1])
				model.MoveRoomsToZone("", session.zone.GetId())
			} else {
				session.zone.SetName(args[1])
			}
		} else if args[0] == "new" {
			zone := model.M.GetZoneByName(args[0])

			if zone != nil {
				session.printError("A zone with that name already exists")
				return
			}

			newZone := model.M.CreateZone(args[1])
			newRoom := model.M.CreateRoom(newZone)

			model.MoveCharacterToRoom(session.player, newRoom)

			session.zone = newZone
			session.room = newRoom

			session.printRoom()
		}
	}
}

func (session *Session) broadcast(args []string) {
	if len(args) == 0 {
		session.printError("Nothing to say")
	} else {
		model.BroadcastMessage(session.player, strings.Join(args, " "))
	}
}

func (session *Session) say(args []string) {
	if len(args) == 0 {
		session.printError("Nothing to say")
	} else {
		model.Say(session.player, strings.Join(args, " "))
	}
}

func (session *Session) emote(args []string) {
	model.Emote(session.player, strings.Join(args, " "))
}

func (session *Session) whisper(args []string) {
	if len(args) < 2 {
		session.printError("Usage: /whisper <player> <message>")
		return
	}

	name := string(args[0])
	targetChar := model.M.GetCharacterByName(name)

	if targetChar == nil {
		session.printError("Player '%s' not found", name)
		return
	}

	if !targetChar.IsOnline() {
		session.printError("Player '%s' is not online", targetChar.PrettyName())
		return
	}

	message := strings.Join(args[1:], " ")
	model.Tell(session.player, targetChar, message)
}

func (session *Session) teleport(args []string) {
	telUsage := func() {
		session.printError("Usage: /teleport [<zone>|<X> <Y> <Z>]")
	}

	x := 0
	y := 0
	z := 0

	newZone := session.zone

	if len(args) == 1 {
		newZone = model.M.GetZoneByName(args[0])

		if newZone == nil {
			session.printError("Zone not found")
			return
		}

		if newZone.GetId() == session.room.GetZoneId() {
			session.printLine("You're already in that zone")
			return
		}

		zoneRooms := model.M.GetRoomsInZone(newZone)

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

	newRoom, err := model.MoveCharacterToLocation(session.player, newZone, database.Coordinate{X: x, Y: y, Z: z})

	if err == nil {
		session.room = newRoom
		session.zone = newZone
		session.printRoom()
	} else {
		session.printError(err.Error())
	}
}

func (session *Session) who() {
	chars := model.M.GetOnlineCharacters()

	session.printLine("")
	session.printLine("Online Players")
	session.printLine("--------------")

	for _, char := range chars {
		session.printLine(char.PrettyName())
	}
	session.printLine("")
}

func (session *Session) colors() {
	session.printLineColor(utils.ColorRed, "Red")
	session.printLineColor(utils.ColorDarkRed, "Dark Red")
	session.printLineColor(utils.ColorGreen, "Green")
	session.printLineColor(utils.ColorDarkGreen, "Dark Green")
	session.printLineColor(utils.ColorBlue, "Blue")
	session.printLineColor(utils.ColorDarkBlue, "Dark Blue")
	session.printLineColor(utils.ColorYellow, "Yellow")
	session.printLineColor(utils.ColorDarkYellow, "Dark Yellow")
	session.printLineColor(utils.ColorMagenta, "Magenta")
	session.printLineColor(utils.ColorDarkMagenta, "Dark Magenta")
	session.printLineColor(utils.ColorCyan, "Cyan")
	session.printLineColor(utils.ColorDarkCyan, "Dark Cyan")
	session.printLineColor(utils.ColorBlack, "Black")
	session.printLineColor(utils.ColorWhite, "White")
	session.printLineColor(utils.ColorGray, "Gray")
}

func (session *Session) colorMode(args []string) {
	if len(args) == 0 {
		message := "Current color mode is: "
		switch session.user.GetColorMode() {
		case utils.ColorModeNone:
			message = message + "None"
		case utils.ColorModeLight:
			message = message + "Light"
		case utils.ColorModeDark:
			message = message + "Dark"
		}
		session.printLine(message)
	} else if len(args) == 1 {
		switch strings.ToLower(args[0]) {
		case "none":
			session.user.SetColorMode(utils.ColorModeNone)
			session.printLine("Color mode set to: None")
		case "light":
			session.user.SetColorMode(utils.ColorModeLight)
			session.printLine("Color mode set to: Light")
		case "dark":
			session.user.SetColorMode(utils.ColorModeDark)
			session.printLine("Color mode set to: Dark")
		default:
			session.printLine("Valid color modes are: None, Light, Dark")
		}
	} else {
		session.printLine("Valid color modes are: None, Light, Dark")
	}
}

func (session *Session) destroyRoom(args []string) {
	if len(args) == 1 {
		direction := database.StringToDirection(args[0])

		if direction == database.DirectionNone {
			session.printError("Not a valid direction")
		} else {
			loc := session.room.NextLocation(direction)
			roomToDelete := model.M.GetRoomByLocation(loc, session.zone)
			if roomToDelete != nil {
				model.M.DeleteRoom(roomToDelete)
				session.printLine("Room destroyed")
			} else {
				session.printError("No room in that direction")
			}
		}
	} else {
		session.printError("Usage: /destroyroom <direction>")
	}
}

func (session *Session) npc(args []string) {
	menu := npcMenu(session.room)
	choice, npcId := session.execMenu(menu)

	getName := func() string {
		name := ""
		for {
			name = session.getUserInput(CleanUserInput, "Desired NPC name: ")
			char := model.M.GetCharacterByName(name)

			if name == "" {
				return ""
			} else if char != nil {
				session.printError("That name is unavailable")
			} else if err := utils.ValidateName(name); err != nil {
				session.printError(err.Error())
			} else {
				break
			}
		}
		return name
	}

	if choice == "" {
		goto done
	}

	if choice == "n" {
		name := getName()
		if name == "" {
			goto done
		}
		model.M.CreateNpc(name, session.room)
	} else if npcId != "" {
		specificMenu := specificNpcMenu(npcId)
		choice, _ := session.execMenu(specificMenu)

		switch choice {
		case "d":
			model.M.DeleteCharacter(npcId)
		case "r":
			name := getName()
			if name == "" {
				break
			}
			npc := model.M.GetCharacter(npcId)
			npc.SetName(name)
		case "c":
			npc := model.M.GetCharacter(npcId)
			conversation := npc.GetConversation()

			if conversation == "" {
				conversation = "<empty>"
			}

			session.printLine("Conversation: %s", conversation)
			newConversation := session.getUserInput(RawUserInput, "New conversation text: ")

			if newConversation != "" {
				npc.SetConversation(newConversation)
			}
		}
	}

done:
	session.printRoom()
}

func (session *Session) createItem(args []string) {
	createUsage := func() {
		session.printError("Usage: /create <item name>")
	}

	if len(args) != 1 {
		createUsage()
		return
	}

	item := model.M.CreateItem(args[0])
	session.room.AddItem(item)
	session.printLine("Item created")
}

func (session *Session) destroyItem(args []string) {
	destroyUsage := func() {
		session.printError("Usage: /destroyitem <item name>")
	}

	if len(args) != 1 {
		destroyUsage()
		return
	}

	itemsInRoom := model.M.GetItems(session.room.GetItemIds())
	name := strings.ToLower(args[0])

	for _, item := range itemsInRoom {
		if strings.ToLower(item.PrettyName()) == name {
			session.room.RemoveItem(item)
			model.M.DeleteItem(item.GetId())
			session.printLine("Item destroyed")
			return
		}
	}

	session.printError("Item not found")
}

func (session *Session) cash(args []string) {
	cashUsage := func() {
		session.printError("Usage: /cash give <amount>")
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

		session.player.AddCash(amount)
		session.printLine("Received: %v monies", amount)
	} else {
		cashUsage()
		return
	}
}

func (session *Session) windowSize() {
	width, height := session.user.WindowSize()

	header := fmt.Sprintf("Width: %v, Height: %v", width, height)

	topBar := header + " " + strings.Repeat("-", int(width)-2-len(header)) + "+"
	bottomBar := "+" + strings.Repeat("-", int(width)-2) + "+"
	outline := "|" + strings.Repeat(" ", int(width)-2) + "|"

	session.printLine(topBar)

	for i := 0; i < int(height)-3; i++ {
		session.printLine(outline)
	}

	session.printLine(bottomBar)
}

func (session *Session) terminalType() {
	session.printLine("Terminal type: %s", session.user.TerminalType())
}

// vim: nocindent
