package session

import (
	"fmt"
	"kmud/database"
	"kmud/engine"
	"kmud/model"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
)

type commandHandler struct {
	session *Session
}

func npcMenu(room *database.Room) *utils.Menu {
	var npcs []*database.Character

	if room != nil {
		npcs = model.M.NpcsIn(room)
	} else {
		npcs = model.M.GetAllNpcs()
	}

	menu := utils.NewMenu("NPCs")

	menu.AddAction("n", "[N]ew")

	for i, npc := range npcs {
		index := i + 1
		actionText := fmt.Sprintf("[%v]%v", index, npc.GetName())
		menu.AddActionData(index, actionText, npc.GetId())
	}

	return menu
}

func specificNpcMenu(npcId bson.ObjectId) *utils.Menu {
	npc := model.M.GetCharacter(npcId)
	menu := utils.NewMenu(npc.GetName())
	menu.AddAction("r", "[R]ename")
	menu.AddAction("d", "[D]elete")
	menu.AddAction("c", "[C]onversation")

	roamingState := "Off"
	if npc.GetProperty(engine.RoamingProperty) == "true" {
		roamingState = "On"
	}

	menu.AddAction("o", fmt.Sprintf("R[o]aming - %s", roamingState))
	return menu
}

func spawnMenu() *utils.Menu {
	menu := utils.NewMenu("Spawn")

	menu.AddAction("n", "[N]ew")

	templates := model.M.GetAllNpcTemplates()

	for i, template := range templates {
		index := i + 1
		actionText := fmt.Sprintf("[%v]%v", index, template.GetName())
		menu.AddActionData(index, actionText, template.GetId())
	}

	return menu
}

func specificSpawnMenu(templateId bson.ObjectId) *utils.Menu {
	template := model.M.GetCharacter(templateId)
	menu := utils.NewMenu(template.GetName())

	menu.AddAction("r", "[R]ename")
	menu.AddAction("d", "[D]elete")

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
	dir := database.StringToDirection(command)

	if dir == database.DirectionNone {
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
	menu := utils.NewMenu("Room")

	menu.AddAction("t", "[T]itle")
	menu.AddAction("d", "[D]escription")
	menu.AddAction("e", "[E]xits")

	for {
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

				direction := database.StringToDirection(choice)
				if direction != database.DirectionNone {
					enable := !ch.session.room.HasExit(direction)
					ch.session.room.SetExitEnabled(direction, enable)

					// Disable the corresponding exit in the adjacent room if necessary
					loc := ch.session.room.NextLocation(direction)
					otherRoom := model.M.GetRoomByLocation(loc, ch.session.zone)
					if otherRoom != nil {
						otherRoom.SetExitEnabled(direction.Opposite(), enable)
					}
				}
			}
		}
	}
}

func (ch *commandHandler) Map(args []string) {
	mapUsage := func() {
		ch.session.printError("Usage: /map [all]")
	}

	startX := 0
	startY := 0
	startZ := 0
	endX := 0
	endY := 0
	endZ := 0

	if len(args) == 0 {
		width, height := ch.session.user.WindowSize()

		loc := ch.session.room.GetLocation()

		startX = loc.X - (width / 4)
		startY = loc.Y - (height / 4)
		startZ = loc.Z

		endX = loc.X + (width / 4)
		endY = loc.Y + (height / 4)
		endZ = loc.Z
	} else if args[0] == "all" {
		topLeft, bottomRight := model.ZoneCorners(ch.session.zone)

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

	// Width and height need to be even numbers so that we don't go off
	// the edge of the screen in either direction
	width -= (width % 2)
	height -= (height % 2)

	depth := endZ - startZ + 1

	builder := newMapBuilder(width, height, depth)
	builder.setUserRoom(ch.session.room)

	for z := startZ; z <= endZ; z++ {
		for y := startY; y <= endY; y++ {
			for x := startX; x <= endX; x++ {
				loc := database.Coordinate{X: x, Y: y, Z: z}
				room := model.M.GetRoomByLocation(loc, ch.session.zone)

				if room != nil {
					// Translate to 0-based coordinates
					builder.addRoom(room, x-startX, y-startY, z-startZ)
				}
			}
		}
	}

	ch.session.printLine(utils.TrimEmptyRows(builder.toString()))
}

func (ch *commandHandler) Zone(args []string) {
	if len(args) == 0 {
		if ch.session.zone.GetId() == "" {
			ch.session.printLine("Currently in the null zone")
		} else {
			ch.session.printLine("Current zone: " + utils.Colorize(utils.ColorBlue, ch.session.zone.GetName()))
		}
	} else if len(args) == 1 {
		if args[0] == "list" {
			ch.session.printLineColor(utils.ColorBlue, "Zones")
			ch.session.printLineColor(utils.ColorBlue, "-----")
			for _, zone := range model.M.GetZones() {
				ch.session.printLine(zone.GetName())
			}
		} else {
			ch.session.printError("Usage: /zone [list|rename <name>|new <name>]")
		}
	} else if len(args) == 2 {
		if args[0] == "rename" {
			zone := model.M.GetZoneByName(args[0])

			if zone != nil {
				ch.session.printError("A zone with that name already exists")
				return
			}

			ch.session.zone.SetName(args[1])
		} else if args[0] == "new" {
			newZone, err := model.M.CreateZone(args[1])

			if err != nil {
				ch.session.printError(err.Error())
				return
			}

			newRoom, err := model.M.CreateRoom(newZone, database.Coordinate{X: 0, Y: 0, Z: 0})
			utils.PanicIfError(err)

			model.MoveCharacterToRoom(ch.session.player, newRoom)

			ch.session.zone = newZone
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
	targetChar := model.M.GetCharacterByName(name)

	if targetChar == nil {
		ch.session.printError("Player '%s' not found", name)
		return
	}

	if !targetChar.IsOnline() {
		ch.session.printError("Player '%s' is not online", targetChar.GetName())
		return
	}

	message := strings.Join(args[1:], " ")
	model.Tell(ch.session.player, targetChar, message)
}

func (ch *commandHandler) Tel(args []string) {
	ch.Teleport(args)
}

func (ch *commandHandler) Teleport(args []string) {
	telUsage := func() {
		ch.session.printError("Usage: /teleport [<zone>|<X> <Y> <Z>]")
	}

	x := 0
	y := 0
	z := 0

	newZone := ch.session.zone

	if len(args) == 1 {
		newZone = model.M.GetZoneByName(args[0])

		if newZone == nil {
			ch.session.printError("Zone not found")
			return
		}

		if newZone.GetId() == ch.session.room.GetZoneId() {
			ch.session.printLine("You're already in that zone")
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

	newRoom, err := model.MoveCharacterToLocation(ch.session.player, newZone, database.Coordinate{X: x, Y: y, Z: z})

	if err == nil {
		ch.session.room = newRoom
		ch.session.zone = newZone
		ch.session.printRoom()
	} else {
		ch.session.printError(err.Error())
	}
}

func (ch *commandHandler) Who(args []string) {
	chars := model.M.GetOnlineCharacters()

	ch.session.printLine("")
	ch.session.printLine("Online Players")
	ch.session.printLine("--------------")

	for _, char := range chars {
		ch.session.printLine(char.GetName())
	}
	ch.session.printLine("")
}

func (ch *commandHandler) Colors(args []string) {
	ch.session.printLineColor(utils.ColorRed, "Red")
	ch.session.printLineColor(utils.ColorDarkRed, "Dark Red")
	ch.session.printLineColor(utils.ColorGreen, "Green")
	ch.session.printLineColor(utils.ColorDarkGreen, "Dark Green")
	ch.session.printLineColor(utils.ColorBlue, "Blue")
	ch.session.printLineColor(utils.ColorDarkBlue, "Dark Blue")
	ch.session.printLineColor(utils.ColorYellow, "Yellow")
	ch.session.printLineColor(utils.ColorDarkYellow, "Dark Yellow")
	ch.session.printLineColor(utils.ColorMagenta, "Magenta")
	ch.session.printLineColor(utils.ColorDarkMagenta, "Dark Magenta")
	ch.session.printLineColor(utils.ColorCyan, "Cyan")
	ch.session.printLineColor(utils.ColorDarkCyan, "Dark Cyan")
	ch.session.printLineColor(utils.ColorBlack, "Black")
	ch.session.printLineColor(utils.ColorWhite, "White")
	ch.session.printLineColor(utils.ColorGray, "Gray")
}

func (ch *commandHandler) CM(args []string) {
	ch.ColorMode(args)
}

func (ch *commandHandler) ColorMode(args []string) {
	if len(args) == 0 {
		message := "Current color mode is: "
		switch ch.session.user.GetColorMode() {
		case utils.ColorModeNone:
			message = message + "None"
		case utils.ColorModeLight:
			message = message + "Light"
		case utils.ColorModeDark:
			message = message + "Dark"
		}
		ch.session.printLine(message)
	} else if len(args) == 1 {
		switch strings.ToLower(args[0]) {
		case "none":
			ch.session.user.SetColorMode(utils.ColorModeNone)
			ch.session.printLine("Color mode set to: None")
		case "light":
			ch.session.user.SetColorMode(utils.ColorModeLight)
			ch.session.printLine("Color mode set to: Light")
		case "dark":
			ch.session.user.SetColorMode(utils.ColorModeDark)
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
		direction := database.StringToDirection(args[0])

		if direction == database.DirectionNone {
			ch.session.printError("Not a valid direction")
		} else {
			loc := ch.session.room.NextLocation(direction)
			roomToDelete := model.M.GetRoomByLocation(loc, ch.session.zone)
			if roomToDelete != nil {
				model.M.DeleteRoom(roomToDelete)
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
		char := model.M.GetCharacterByName(name)

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
				model.M.CreateNpc(name, ch.session.room)
			}
		} else if npcId != "" {
			for {
				specificMenu := specificNpcMenu(npcId)
				choice, _ := ch.session.execMenu(specificMenu)
				npc := model.M.GetCharacter(npcId)

				if choice == "d" {
					model.M.DeleteCharacterId(npcId)
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
					currentVal := npc.GetProperty(engine.RoamingProperty)
					newVal := "true"
					if currentVal == "true" {
						newVal = "false"
					}

					npc.SetProperty(engine.RoamingProperty, newVal)
				} else if choice == "" {
					break
				}
			}
		}
	}

	ch.session.printRoom()
}

func (ch *commandHandler) Spawn(args []string) {
	for {
		menu := spawnMenu()
		choice, templateId := ch.session.execMenu(menu)

		if choice == "" {
			break
		} else if choice == "n" {
			name := getNpcName(ch)
			if name != "" {
				model.M.CreateNpcTemplate(name)
			}
		} else {
			for {
				specificMenu := specificSpawnMenu(templateId)
				choice, _ := ch.session.execMenu(specificMenu)

				if choice == "" {
					break
				} else if choice == "r" {
					newName := getNpcName(ch)
					if newName != "" {
						template := model.M.GetCharacter(templateId)
						template.SetName(newName)
					}
				} else if choice == "d" {
					model.M.DeleteCharacterId(templateId)
					break
				}
			}
		}
	}
}

func (ch *commandHandler) Create(args []string) {
	createUsage := func() {
		ch.session.printError("Usage: /create <item name>")
	}

	if len(args) != 1 {
		createUsage()
		return
	}

	item := model.M.CreateItem(args[0])
	ch.session.room.AddItem(item)
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

	itemsInRoom := model.M.GetItems(ch.session.room.GetItemIds())
	name := strings.ToLower(args[0])

	for _, item := range itemsInRoom {
		if strings.ToLower(item.GetName()) == name {
			ch.session.room.RemoveItem(item)
			model.M.DeleteItem(item)
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
	width, height := ch.session.user.WindowSize()

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
	ch.session.printLine("Terminal type: %s", ch.session.user.TerminalType())
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
	targetChar := model.M.GetCharacter(ch.session.replyId)

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

// vim: nocindent
