package game

import (
	"fmt"
	"io"
	"kmud/database"
	"kmud/model"
	"kmud/utils"
	"strconv"
	"strings"
	"time"
)

type userInputMode int

const (
	CleanUserInput userInputMode = iota
	RawUserInput   userInputMode = iota
)

func getToggleExitMenu(cm utils.ColorMode, room database.Room) utils.Menu {
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

func getNpcMenu(room database.Room) utils.Menu {
	npcs := model.M.NpcsIn(room.Id)

	menu := utils.NewMenu("NPCs")

	menu.AddAction("n", "[N]ew")

	for i, npc := range npcs {
		index := i + 1
		actionText := fmt.Sprintf("[%v]%v", index, npc.PrettyName())
		menu.AddActionData(index, actionText, npc.Id)
	}

	return menu
}

func Exec(conn io.ReadWriter, user *database.User, character *database.Character) {
	currentRoom := model.M.GetRoom(character.RoomId)
	currentZone := model.M.GetZone(currentRoom.ZoneId)

	printString := func(data string) {
		io.WriteString(conn, data)
	}

	printLineColor := func(color utils.Color, line string, a ...interface{}) {
		utils.WriteLine(conn, utils.Colorize(user.ColorMode, color, fmt.Sprintf(line, a...)))
	}

	printLine := func(line string, a ...interface{}) {
		printLineColor(utils.ColorWhite, line, a...)
	}

	printError := func(err string, a ...interface{}) {
		printLineColor(utils.ColorRed, err, a...)
	}

	printRoom := func() {
		charList := model.M.CharactersIn(currentRoom.Id, character.Id)
		printLine(currentRoom.ToString(database.ReadMode, user.ColorMode, charList))
	}

	printRoomEditor := func() {
		printLine(currentRoom.ToString(database.EditMode, user.ColorMode, nil))
	}

	prompt := func() string {
		return "> "
	}

	processEvent := func(event model.Event) string {
		message := event.ToString(*character)

		switch event.Type() {
		case model.RoomUpdateEventType:
			roomEvent := event.(model.RoomUpdateEvent)
			if roomEvent.Room.Id == currentRoom.Id {
				currentRoom = roomEvent.Room
			}
		}

		return message
	}

	eventChannel := model.Register(character)
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
					printLine("\n" + message)
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
	execMenu := func(menu utils.Menu) string {
		choice := ""
		for {
			menu.Print(conn, user.ColorMode)
			choice = getUserInput(CleanUserInput, menu.Prompt)
			if menu.HasAction(choice) || choice == "" {
				break
			}
		}
		return choice
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
					printLine("Nothing to see")
				} else {
					loc := currentRoom.Location.Next(arg)
					roomToSee, found := model.M.GetRoomByLocation(loc, currentZone.Id)
					if found {
						printLine(roomToSee.ToString(database.ReadMode, user.ColorMode, model.M.CharactersIn(roomToSee.Id, "")))
					} else {
						printLine("Nothing to see")
					}
				}
			}

		case "ls":
			printLine("Where do you think you are?!")

		case "i":
			printLine("You aren't carrying anything")

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
					newRoom, err := model.MoveCharacter(character, direction)
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
				input := getUserInput(CleanUserInput, "Select a section to edit> ")

				switch input {
				case "":
					printRoom()
					return

				case "1":
					input = getUserInput(RawUserInput, "Enter new title: ")

					if input != "" {
						currentRoom.Title = input
						model.M.UpdateRoom(currentRoom)
					}
					printRoomEditor()

				case "2":
					input = getUserInput(RawUserInput, "Enter new description: ")

					if input != "" {
						currentRoom.Description = input
						model.M.UpdateRoom(currentRoom)
					}
					printRoomEditor()

				case "3":
					for {
						menu := getToggleExitMenu(user.ColorMode, currentRoom)

						choice := execMenu(menu)

						if choice == "" {
							break
						}

						direction := database.StringToDirection(choice)
						if direction != database.DirectionNone {
							enable := !currentRoom.HasExit(direction)
							currentRoom.SetExitEnabled(direction, enable)
							model.M.UpdateRoom(currentRoom)
						}
					}

					printRoomEditor()

				default:
					printLine("Invalid selection")
				}
			}

		case "loc":
			fallthrough
		case "location":
			printLine("%v", currentRoom.Location)

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
					startX = currentRoom.Location.X - radius
					startY = currentRoom.Location.Y - radius
					startZ = currentRoom.Location.Z
					endX = startX + (radius * 2)
					endY = startY + (radius * 2)
					endZ = currentRoom.Location.Z
				} else if args[0] == "all" {
					topLeft, bottomRight := model.ZoneCorners(currentZone.Id)

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
						currentRoom, found := model.M.GetRoomByLocation(loc, currentZone.Id)

						if found {
							// Translate to 0-based coordinates and double the coordinate
							// space to leave room for the exit lines
							builder.addRoom(currentRoom, (x-startX)*2, (y-startY)*2, z-startZ)
						}
					}
				}
			}

			printLine(utils.TrimEmptyRows(builder.toString(user.ColorMode)))

		case "maps":
			printLine(strings.Join(maps(), "\n"))

		case "zone":
			if len(args) == 0 {
				if currentZone.Id == "" {
					printLine("Currently in the null zone")
				} else {
					printLine("Current zone: " + utils.Colorize(user.ColorMode, utils.ColorBlue, currentZone.Name))
				}
			} else if len(args) == 1 {
				if args[0] == "list" {
					for _, zone := range model.M.GetZones() {
						printLine(zone.Name)
					}
				} else {
					printError("Usage: /zone [list|rename <name>|new <name>]")
				}
			} else if len(args) == 2 {
				if args[0] == "rename" {
					_, found := model.M.GetZoneByName(args[0])

					if found {
						printError("A zone with that name already exists")
						return
					}

					if currentZone.Id == "" {
						currentZone = database.NewZone(args[1])
						model.M.UpdateZone(currentZone)
						model.MoveRoomsToZone("", currentZone.Id)
					} else {
						currentZone.Name = args[1]
						model.M.UpdateZone(currentZone)
					}
				} else if args[0] == "new" {
					_, found := model.M.GetZoneByName(args[0])

					if found {
						printError("A zone with that name already exists")
						return
					}

					newZone := database.NewZone(args[1])
					model.M.UpdateZone(newZone)

					newRoom := database.NewRoom(newZone.Id)
					model.M.UpdateRoom(newRoom)

					model.MoveCharacterToRoom(character, newRoom)

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
				model.BroadcastMessage(*character, strings.Join(args, " "))
			}

		case "say":
			fallthrough
		case "s":
			if len(args) == 0 {
				printError("Nothing to say")
			} else {
				model.Say(*character, strings.Join(args, " "))
			}

		case "me":
			model.Emote(*character, strings.Join(args, " "))

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
			targetChar, found := model.M.GetCharacterByName(name)

			if !found {
				printError("Player '%s' not found", name)
				return
			}

			if !targetChar.Online() {
				printError("Player '%s' is not online", targetChar.PrettyName())
				return
			}

			message := strings.Join(args[1:], " ")
			model.Tell(*character, targetChar, message)

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
				var found bool
				newZone, found = model.M.GetZoneByName(args[0])

				if !found {
					printError("Zone not found")
					return
				}

				if newZone.Id == currentRoom.ZoneId {
					printLine("You're already in that zone")
					return
				}

				zoneRooms := model.M.GetRoomsInZone(newZone.Id)

				if len(zoneRooms) > 0 {
					r := zoneRooms[0]
					x = r.Location.X
					y = r.Location.Y
					z = r.Location.Z
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

			newRoom, err := model.MoveCharacterToLocation(character, newZone.Id, database.Coordinate{X: x, Y: y, Z: z})

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
				switch user.ColorMode {
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
					user.ColorMode = utils.ColorModeNone
					model.M.UpdateUser(*user)
					printLine("Color mode set to: None")
				case "light":
					user.ColorMode = utils.ColorModeLight
					model.M.UpdateUser(*user)
					printLine("Color mode set to: Light")
				case "dark":
					user.ColorMode = utils.ColorModeDark
					model.M.UpdateUser(*user)
					printLine("Color mode set to: Dark")
				default:
					printLine("Valid color modes are: None, Light, Dark")
				}
			} else {
				printLine("Valid color modes are: None, Light, Dark")
			}

		case "delete":
			fallthrough
		case "del":
			if len(args) == 1 {
				direction := database.StringToDirection(args[0])

				if direction == database.DirectionNone {
					printError("Not a valid direction")
				} else {
					loc := currentRoom.Location.Next(direction)
					roomToDelete, found := model.M.GetRoomByLocation(loc, currentZone.Id)
					if found {
						model.DeleteRoom(roomToDelete)
					} else {
						printError("No room in that direction")
					}
				}
			} else {
				printError("Usage: /delete <direction>")
			}

		case "npc":
			menu := getNpcMenu(currentRoom)
			choice := execMenu(menu)

			if choice == "" {
				goto done
			}

			if choice == "n" {
				name := ""
				description := ""

				for {
					name = getUserInput(CleanUserInput, "Desired NPC name: ")
					_, found := model.M.GetCharacterByName(name)

					if name == "" {
						goto done
					} else if found {
						printError("That name is unavailable")
					} else if err := utils.ValidateName(name); err != nil {
						printError(err.Error())
					} else {
						break
					}
				}

				description = getUserInput(RawUserInput, "NPC description: ")

				if description == "" {
					goto done
				}
			}

		done:
			printRoom()

		default:
			printError("Unrecognized command: %s", command)
		}
	}

	printLineColor(utils.ColorWhite, "Welcome, "+character.PrettyName())
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
			prompt := utils.Colorize(user.ColorMode, utils.ColorWhite, <-promptChannel)
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
