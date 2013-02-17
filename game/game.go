package game

import (
	"fmt"
	"io"
	"kmud/database"
	"kmud/model"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
	"time"
)

type userInputMode int

const (
	CleanUserInput userInputMode = iota
	RawUserInput   userInputMode = iota
)

func toggleExitMenu(cm utils.ColorMode, room *database.Room) utils.Menu {
	onOrOff := func(direction database.ExitDirection) string {
		text := "Off"
		if room.HasExit(direction) {
			text = "On"
		}
		return utils.Colorize(cm, utils.ColorBlue, text)
	}

	menu := utils.NewMenu("Edit Exits")

	menu.AddAction("n", "[N]orth: "+onOrOff(database.DirectionNorth))
	menu.AddAction("ne", "[NE]North East: "+onOrOff(database.DirectionNorthEast))
	menu.AddAction("e", "[E]ast: "+onOrOff(database.DirectionEast))
	menu.AddAction("se", "[SE]South East: "+onOrOff(database.DirectionSouthEast))
	menu.AddAction("s", "[S]outh: "+onOrOff(database.DirectionSouth))
	menu.AddAction("sw", "[SW]South West: "+onOrOff(database.DirectionSouthWest))
	menu.AddAction("w", "[W]est: "+onOrOff(database.DirectionWest))
	menu.AddAction("nw", "[NW]North West: "+onOrOff(database.DirectionNorthWest))
	menu.AddAction("u", "[U]p: "+onOrOff(database.DirectionUp))
	menu.AddAction("d", "[D]own: "+onOrOff(database.DirectionDown))

	return menu
}

func npcMenu(room *database.Room) utils.Menu {
	npcs := model.M.NpcsIn(room)

	menu := utils.NewMenu("NPCs")

	menu.AddAction("n", "[N]ew")

	for i, npc := range npcs {
		index := i + 1
		actionText := fmt.Sprintf("[%v]%v", index, npc.PrettyName())
		menu.AddActionData(index, actionText, npc.GetId())
	}

	return menu
}

func specificNpcMenu(npcId bson.ObjectId) utils.Menu {
	npc := model.M.GetCharacter(npcId)
	menu := utils.NewMenu(npc.PrettyName())
	menu.AddAction("r", "[R]ename")
	menu.AddAction("d", "[D]elete")
	menu.AddAction("c", "[C]onversation")
	return menu
}

func Exec(conn io.ReadWriter, currentUser *database.User, currentPlayer *database.Character) {
	currentRoom := model.M.GetRoom(currentPlayer.GetRoomId())
	currentZone := model.M.GetZone(currentRoom.GetZoneId())

	printString := func(data string) {
		io.WriteString(conn, data)
	}

	printLineColor := func(color utils.Color, line string, a ...interface{}) {
		utils.WriteLine(conn, utils.Colorize(currentUser.GetColorMode(), color, fmt.Sprintf(line, a...)))
	}

	printLine := func(line string, a ...interface{}) {
		printLineColor(utils.ColorWhite, line, a...)
	}

	printError := func(err string, a ...interface{}) {
		printLineColor(utils.ColorRed, err, a...)
	}

	printRoom := func() {
		playerList := model.M.PlayersIn(currentRoom, currentPlayer)
		npcList := model.M.NpcsIn(currentRoom)
		printLine(currentRoom.ToString(database.ReadMode, currentUser.GetColorMode(),
			playerList, npcList, model.M.GetItems(currentRoom.GetItemIds())))
	}

	printRoomEditor := func() {
		printLine(currentRoom.ToString(database.EditMode, currentUser.GetColorMode(), nil, nil, nil))
	}

	clearLine := func() {
		utils.ClearLine(conn)
	}

	prompt := func() string {
		return "> "
	}

	processEvent := func(event model.Event) string {
		message := event.ToString(currentPlayer)

		switch event.Type() {
		case model.RoomUpdateEventType:
			roomEvent := event.(model.RoomUpdateEvent)
			if roomEvent.Room.GetId() == currentRoom.GetId() {
				currentRoom = roomEvent.Room
			}
		}

		return message
	}

	eventChannel := model.Register(currentPlayer)
	defer model.Unregister(eventChannel)

	userInputChannel := make(chan string)
	promptChannel := make(chan string)

	inputModeChannel := make(chan userInputMode)
	panicChannel := make(chan interface{})

	/**
	 * Allows us to retrieve user input in a way that doesn't block the
	 * event loop by using channels and a separate Go routine to grab
	 * either the next user input or the next event.
	 */
	getUserInput := func(inputMode userInputMode, prompt string) string {
		inputModeChannel <- inputMode
		promptChannel <- prompt

		for {
			select {
			case input := <-userInputChannel:
				return input
			case event := <-eventChannel:
				message := processEvent(event)
				if message != "" {
					clearLine()
					printLine(message)
					printString(prompt)
				}
			case quitMessage := <-panicChannel:
				panic(quitMessage)
			}
		}
		panic("Unexpected code path")
	}

	// Same behavior as menu.Exec(), except that it uses getUserInput
	// which doesn't block the event loop while waiting for input
	execMenu := func(menu utils.Menu) (string, bson.ObjectId) {
		choice := ""
		var data bson.ObjectId

		for {
			menu.Print(conn, currentUser.GetColorMode())
			choice = getUserInput(CleanUserInput, menu.GetPrompt())
			if menu.HasAction(choice) || choice == "" {
				data = menu.GetData(choice)
				break
			}
		}
		return choice, data
	}

	processAction := func(action string, args []string) {
		switch action {
		case "l":
			fallthrough
		case "look":
			if len(args) == 0 {
				printRoom()
			} else if len(args) == 1 {
				arg := database.StringToDirection(args[0])

				if arg == database.DirectionNone {
					charList := model.M.CharactersIn(currentRoom)
					index := utils.BestMatch(args[0], database.CharacterNames(charList))

					if index == -2 {
						printError("Which one do you mean?")
					} else if index != -1 {
						printLine("Looking at: %s", charList[index].PrettyName())
					} else {
						itemList := model.M.ItemsIn(currentRoom)
						index = utils.BestMatch(args[0], database.ItemNames(itemList))

						if index == -1 {
							printLine("Nothing to see")
						} else if index == -2 {
							printError("Which one do you mean?")
						} else {
							printLine("Looking at: %s", itemList[index].PrettyName())
						}
					}
				} else {
					loc := currentRoom.NextLocation(arg)
					roomToSee := model.M.GetRoomByLocation(loc, currentZone)
					if roomToSee != nil {
						printLine(roomToSee.ToString(database.ReadMode, currentUser.GetColorMode(),
							model.M.PlayersIn(roomToSee, nil), model.M.NpcsIn(roomToSee), nil))
					} else {
						printLine("Nothing to see")
					}
				}
			}

		case "ls":
			printLine("Where do you think you are?!")

		case "a":
			fallthrough
		case "attack":
			charList := model.M.CharactersIn(currentRoom)
			index := utils.BestMatch(args[0], database.CharacterNames(charList))

			if index == -1 {
				printError("Not found")
			} else if index == -2 {
				printError("Which one do you mean?")
			} else {
				defender := charList[index]
				if defender.GetId() == currentPlayer.GetId() {
					printError("You can't attack yourself")
				} else {
					model.StartFight(currentPlayer, defender)
				}
			}

		case "stop":
			model.StopFight(currentPlayer)
			return

		case "inventory":
			fallthrough
		case "inv":
			fallthrough
		case "i":
			itemIds := currentPlayer.GetItemIds()

			if len(itemIds) == 0 {
				printLine("You aren't carrying anything")
			} else {
				var itemNames []string
				for _, item := range model.M.GetItems(itemIds) {
					itemNames = append(itemNames, item.PrettyName())
				}
				printLine("You are carrying: %s", strings.Join(itemNames, ", "))
			}

			printLine("Cash: %v", currentPlayer.GetCash())

		case "take":
			fallthrough
		case "t":
			fallthrough
		case "get":
			fallthrough
		case "g":
			takeUsage := func() {
				printError("Usage: take <item name>")
			}

			if len(args) != 1 {
				takeUsage()
				return
			}

			itemsInRoom := model.M.GetItems(currentRoom.GetItemIds())
			itemName := strings.ToLower(args[0])
			for _, item := range itemsInRoom {
				if strings.ToLower(item.PrettyName()) == itemName {
					currentPlayer.AddItem(item)
					currentRoom.RemoveItem(item)
					return
				}
			}

			printError("Item %s not found", args[0])

		case "drop":
			dropUsage := func() {
				printError("Usage: drop <item name>")
			}

			if len(args) != 1 {
				dropUsage()
				return
			}

			characterItems := model.M.GetItems(currentPlayer.GetItemIds())

			itemName := strings.ToLower(args[0])
			for _, item := range characterItems {
				if strings.ToLower(item.PrettyName()) == itemName {
					currentPlayer.RemoveItem(item)
					currentRoom.AddItem(item)
					return
				}
			}

			printError("You are not carrying a %s", args[0])

		case "talk":
			if len(args) != 1 {
				printError("Usage: talk <NPC name>")
				return
			}

			npcList := model.M.NpcsIn(currentRoom)
			index := utils.BestMatch(args[0], database.CharacterNames(npcList))

			if index == -1 {
				printError("Not found")
			} else if index == -2 {
				printError("Which one do you mean?")
			} else {
				npc := npcList[index]
				printLine(npc.PrettyConversation(currentUser.GetColorMode()))
			}

		case "":
			fallthrough
		case "logout":
			return

		case "quit":
			fallthrough
		case "exit":
			printLine("Take luck!")
			panic("User quit")

		default:
			direction := database.StringToDirection(action)

			if direction != database.DirectionNone {
				if currentRoom.HasExit(direction) {
					newRoom, err := model.MoveCharacter(currentPlayer, direction)
					if err == nil {
						currentRoom = newRoom
						printRoom()
					} else {
						printError(err.Error())
					}

				} else {
					printError("You can't go that way")
				}
			} else {
				printError("You can't do that")
			}
		}
	}

	processCommand := func(command string, args []string) {
		switch command {
		case "?":
			fallthrough
		case "help":
		case "edit":
			printRoomEditor()

			for {
				input := getUserInput(CleanUserInput, "Select a section to edit: ")

				switch input {
				case "":
					printRoom()
					return

				case "1":
					input = getUserInput(RawUserInput, "Enter new title: ")

					if input != "" {
						currentRoom.SetTitle(input)
					}
					printRoomEditor()

				case "2":
					input = getUserInput(RawUserInput, "Enter new description: ")

					if input != "" {
						currentRoom.SetDescription(input)
					}
					printRoomEditor()

				case "3":
					for {
						menu := toggleExitMenu(currentUser.GetColorMode(), currentRoom)

						choice, _ := execMenu(menu)

						if choice == "" {
							break
						}

						direction := database.StringToDirection(choice)
						if direction != database.DirectionNone {
							enable := !currentRoom.HasExit(direction)
							currentRoom.SetExitEnabled(direction, enable)
						}
					}

					printRoomEditor()

				default:
					printLine("Invalid selection")
				}
			}

			// Quick room/exit creation
		case "/n":
			currentRoom.SetExitEnabled(database.DirectionNorth, true)
			processAction("n", []string{})
			currentRoom.SetExitEnabled(database.DirectionSouth, true)

		case "/e":
			currentRoom.SetExitEnabled(database.DirectionEast, true)
			processAction("e", []string{})
			currentRoom.SetExitEnabled(database.DirectionWest, true)

		case "/s":
			currentRoom.SetExitEnabled(database.DirectionSouth, true)
			processAction("s", []string{})
			currentRoom.SetExitEnabled(database.DirectionNorth, true)

		case "/w":
			currentRoom.SetExitEnabled(database.DirectionWest, true)
			processAction("w", []string{})
			currentRoom.SetExitEnabled(database.DirectionEast, true)

		case "/u":
			currentRoom.SetExitEnabled(database.DirectionUp, true)
			processAction("u", []string{})
			currentRoom.SetExitEnabled(database.DirectionDown, true)

		case "/d":
			currentRoom.SetExitEnabled(database.DirectionDown, true)
			processAction("d", []string{})
			currentRoom.SetExitEnabled(database.DirectionUp, true)

		case "/ne":
			currentRoom.SetExitEnabled(database.DirectionNorthEast, true)
			processAction("ne", []string{})
			currentRoom.SetExitEnabled(database.DirectionSouthWest, true)

		case "/nw":
			currentRoom.SetExitEnabled(database.DirectionNorthWest, true)
			processAction("nw", []string{})
			currentRoom.SetExitEnabled(database.DirectionSouthEast, true)

		case "/se":
			currentRoom.SetExitEnabled(database.DirectionSouthEast, true)
			processAction("se", []string{})
			currentRoom.SetExitEnabled(database.DirectionNorthWest, true)

		case "/sw":
			currentRoom.SetExitEnabled(database.DirectionSouthWest, true)
			processAction("sw", []string{})
			currentRoom.SetExitEnabled(database.DirectionNorthEast, true)

		case "loc":
			fallthrough
		case "location":
			printLine("%v", currentRoom.GetLocation())

		case "map":
			mapUsage := func() {
				printError("Usage: /map [<radius>|all|load <name>]")
			}

			startX := 0
			startY := 0
			startZ := 0
			endX := 0
			endY := 0
			endZ := 0

			if len(args) == 0 {
				args = append(args, "10")
			}

			if len(args) == 1 {
				radius, err := strconv.Atoi(args[0])

				if err == nil && radius > 0 {
					startX = currentRoom.GetLocation().X - radius
					startY = currentRoom.GetLocation().Y - radius
					startZ = currentRoom.GetLocation().Z
					endX = startX + (radius * 2)
					endY = startY + (radius * 2)
					endZ = currentRoom.GetLocation().Z
				} else if args[0] == "all" {
					topLeft, bottomRight := model.ZoneCorners(currentZone.GetId())

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
			} else {
				mapUsage()
				return
			}

			width := endX - startX + 1
			height := endY - startY + 1
			depth := endZ - startZ + 1

			builder := newMapBuilder(width, height, depth)
			builder.setUserRoom(currentRoom)

			for z := startZ; z <= endZ; z += 1 {
				for y := startY; y <= endY; y += 1 {
					for x := startX; x <= endX; x += 1 {
						loc := database.Coordinate{x, y, z}
						room := model.M.GetRoomByLocation(loc, currentZone)

						if room != nil {
							// Translate to 0-based coordinates and double the coordinate
							// space to leave room for the exit lines
							builder.addRoom(room, (x-startX)*2, (y-startY)*2, z-startZ)
						}
					}
				}
			}

			printLine(utils.TrimEmptyRows(builder.toString(currentUser.GetColorMode())))

		case "zone":
			if len(args) == 0 {
				if currentZone.GetId() == "" {
					printLine("Currently in the null zone")
				} else {
					printLine("Current zone: " + utils.Colorize(currentUser.GetColorMode(), utils.ColorBlue, currentZone.GetName()))
				}
			} else if len(args) == 1 {
				if args[0] == "list" {
					printLineColor(utils.ColorBlue, "Zones")
					printLineColor(utils.ColorBlue, "-----")
					for _, zone := range model.M.GetZones() {
						printLine(zone.GetName())
					}
				} else {
					printError("Usage: /zone [list|rename <name>|new <name>]")
				}
			} else if len(args) == 2 {
				if args[0] == "rename" {
					zone := model.M.GetZoneByName(args[0])

					if zone != nil {
						printError("A zone with that name already exists")
						return
					}

					if currentZone.GetId() == "" {
						currentZone = model.M.CreateZone(args[1])
						model.MoveRoomsToZone("", currentZone.GetId())
					} else {
						currentZone.SetName(args[1])
					}
				} else if args[0] == "new" {
					zone := model.M.GetZoneByName(args[0])

					if zone != nil {
						printError("A zone with that name already exists")
						return
					}

					newZone := model.M.CreateZone(args[1])
					newRoom := model.M.CreateRoom(newZone)

					model.MoveCharacterToRoom(currentPlayer, newRoom)

					currentZone = newZone
					currentRoom = newRoom

					printRoom()
				}
			}

		case "broadcast":
			fallthrough
		case "b":
			if len(args) == 0 {
				printError("Nothing to say")
			} else {
				model.BroadcastMessage(currentPlayer, strings.Join(args, " "))
			}

		case "say":
			fallthrough
		case "s":
			if len(args) == 0 {
				printError("Nothing to say")
			} else {
				model.Say(currentPlayer, strings.Join(args, " "))
			}

		case "me":
			model.Emote(currentPlayer, strings.Join(args, " "))

		case "whisper":
			fallthrough
		case "tell":
			fallthrough
		case "w":
			if len(args) < 2 {
				printError("Usage: /whisper <player> <message>")
				return
			}

			name := string(args[0])
			targetChar := model.M.GetCharacterByName(name)

			if targetChar == nil {
				printError("Player '%s' not found", name)
				return
			}

			if !targetChar.IsOnline() {
				printError("Player '%s' is not online", targetChar.PrettyName())
				return
			}

			message := strings.Join(args[1:], " ")
			model.Tell(currentPlayer, targetChar, message)

		case "teleport":
			fallthrough
		case "tel":
			telUsage := func() {
				printError("Usage: /teleport [<zone>|<X> <Y> <Z>]")
			}

			x := 0
			y := 0
			z := 0

			newZone := currentZone

			if len(args) == 1 {
				newZone = model.M.GetZoneByName(args[0])

				if newZone == nil {
					printError("Zone not found")
					return
				}

				if newZone.GetId() == currentRoom.GetZoneId() {
					printLine("You're already in that zone")
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

			newRoom, err := model.MoveCharacterToLocation(currentPlayer, newZone, database.Coordinate{X: x, Y: y, Z: z})

			if err == nil {
				currentRoom = newRoom
				currentZone = newZone
				printRoom()
			} else {
				printError(err.Error())
			}

		case "who":
			chars := model.M.GetOnlineCharacters()

			printLine("")
			printLine("Online Players")
			printLine("--------------")

			for _, char := range chars {
				printLine(char.PrettyName())
			}
			printLine("")

		case "colors":
			printLineColor(utils.ColorRed, "Red")
			printLineColor(utils.ColorDarkRed, "Dark Red")
			printLineColor(utils.ColorGreen, "Green")
			printLineColor(utils.ColorDarkGreen, "Dark Green")
			printLineColor(utils.ColorBlue, "Blue")
			printLineColor(utils.ColorDarkBlue, "Dark Blue")
			printLineColor(utils.ColorYellow, "Yellow")
			printLineColor(utils.ColorDarkYellow, "Dark Yellow")
			printLineColor(utils.ColorMagenta, "Magenta")
			printLineColor(utils.ColorDarkMagenta, "Dark Magenta")
			printLineColor(utils.ColorCyan, "Cyan")
			printLineColor(utils.ColorDarkCyan, "Dark Cyan")
			printLineColor(utils.ColorBlack, "Black")
			printLineColor(utils.ColorWhite, "White")
			printLineColor(utils.ColorGray, "Gray")

		case "colormode":
			fallthrough
		case "cm":
			if len(args) == 0 {
				message := "Current color mode is: "
				switch currentUser.GetColorMode() {
				case utils.ColorModeNone:
					message = message + "None"
				case utils.ColorModeLight:
					message = message + "Light"
				case utils.ColorModeDark:
					message = message + "Dark"
				}
				printLine(message)
			} else if len(args) == 1 {
				switch strings.ToLower(args[0]) {
				case "none":
					currentUser.SetColorMode(utils.ColorModeNone)
					printLine("Color mode set to: None")
				case "light":
					currentUser.SetColorMode(utils.ColorModeLight)
					printLine("Color mode set to: Light")
				case "dark":
					currentUser.SetColorMode(utils.ColorModeDark)
					printLine("Color mode set to: Dark")
				default:
					printLine("Valid color modes are: None, Light, Dark")
				}
			} else {
				printLine("Valid color modes are: None, Light, Dark")
			}

		case "destroyroom":
			if len(args) == 1 {
				direction := database.StringToDirection(args[0])

				if direction == database.DirectionNone {
					printError("Not a valid direction")
				} else {
					loc := currentRoom.NextLocation(direction)
					roomToDelete := model.M.GetRoomByLocation(loc, currentZone)
					if roomToDelete != nil {
						model.DeleteRoom(roomToDelete)
						printLine("Room destroyed")
					} else {
						printError("No room in that direction")
					}
				}
			} else {
				printError("Usage: /destroyroom <direction>")
			}

		case "npc":
			menu := npcMenu(currentRoom)
			choice, npcId := execMenu(menu)

			getName := func() string {
				name := ""
				for {
					name = getUserInput(CleanUserInput, "Desired NPC name: ")
					char := model.M.GetCharacterByName(name)

					if name == "" {
						return ""
					} else if char != nil {
						printError("That name is unavailable")
					} else if err := utils.ValidateName(name); err != nil {
						printError(err.Error())
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
				model.M.CreateNpc(name, currentRoom)
			} else if npcId != "" {
				specificMenu := specificNpcMenu(npcId)
				choice, _ := execMenu(specificMenu)

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

					printLine("Conversation: %s", conversation)
					newConversation := getUserInput(RawUserInput, "New conversation text: ")

					if newConversation != "" {
						npc.SetConversation(newConversation)
					}
				}
			}

		done:
			printRoom()

		case "create":
			createUsage := func() {
				printError("Usage: /create <item name>")
			}

			if len(args) != 1 {
				createUsage()
				return
			}

			item := model.M.CreateItem(args[0])
			currentRoom.AddItem(item)
			printLine("Item created")

		case "destroyitem":
			destroyUsage := func() {
				printError("Usage: /destroyitem <item name>")
			}

			if len(args) != 1 {
				destroyUsage()
				return
			}

			itemsInRoom := model.M.GetItems(currentRoom.GetItemIds())
			name := strings.ToLower(args[0])

			for _, item := range itemsInRoom {
				if strings.ToLower(item.PrettyName()) == name {
					currentRoom.RemoveItem(item)
					model.M.DeleteItem(item.GetId())
					printLine("Item destroyed")
					return
				}
			}

			printError("Item not found")

		case "cash":
			cashUsage := func() {
				printError("Usage: /cash give <amount>")
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

				currentPlayer.AddCash(amount)
				printLine("Received: %v monies", amount)
			} else {
				cashUsage()
				return
			}

		case "roomid":
			printLine("Room ID: %v", currentRoom.GetId())

		default:
			printError("Unrecognized command: %s", command)
		}
	}

	printLineColor(utils.ColorWhite, "Welcome, "+currentPlayer.PrettyName())
	printRoom()

	// Main routine in charge of actually reading input from the connection object,
	// also has built in throttling to limit how fast we are allowed to process
	// commands from the user. 
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicChannel <- r
			}
		}()

		lastTime := time.Now()

		delay := time.Duration(200) * time.Millisecond

		for {
			mode := <-inputModeChannel
			prompt := utils.Colorize(currentUser.GetColorMode(), utils.ColorWhite, <-promptChannel)
			input := ""

			switch mode {
			case CleanUserInput:
				input = utils.GetUserInput(conn, prompt)
			case RawUserInput:
				input = utils.GetRawUserInput(conn, prompt)
			default:
				panic("Unhandled case in switch statement (userInputMode)")
			}

			diff := time.Since(lastTime)

			if diff < delay {
				time.Sleep(delay - diff)
			}

			userInputChannel <- input
			lastTime = time.Now()
		}
	}()

	// Main loop
	for {
		input := getUserInput(RawUserInput, prompt())
		if input == "" {
			return
		}
		if strings.HasPrefix(input, "/") {
			processCommand(utils.Argify(input[1:]))
		} else {
			processAction(utils.Argify(input))
		}
	}
}

// vim: nocindent
