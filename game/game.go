package game

import (
	"fmt"
	"io"
	"kmud/database"
	"kmud/engine"
	"kmud/utils"
	"net"
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

func Exec(conn net.Conn, user *database.User, character *database.Character) {
	room := engine.M.GetRoom(character.RoomId)
	zone := engine.M.GetZone(room.ZoneId)

	printString := func(data string) {
		io.WriteString(conn, data)
	}

	printLineColor := func(color utils.Color, line string) {
		utils.WriteLine(conn, utils.Colorize(user.ColorMode, color, line))
	}

	printLine := func(line string) {
		printLineColor(utils.ColorWhite, line)
	}

	printError := func(line string) {
		printLineColor(utils.ColorRed, line)
	}

	printRoom := func() {
		charList := engine.M.CharactersIn(room, *character)
		printLine(room.ToString(database.ReadMode, user.ColorMode, charList))
	}

	printRoomEditor := func() {
		printLine(room.ToString(database.EditMode, user.ColorMode, nil))
	}

	prompt := func() string {
		return "> "
	}

	processEvent := func(event engine.Event) string {
		message := event.ToString(*character)

		switch event.Type() {
		case engine.RoomUpdateEventType:
			roomEvent := event.(engine.RoomUpdateEvent)
			if roomEvent.Room.Id == room.Id {
				room = roomEvent.Room
			}
		}

		return message
	}

	eventChannel := engine.Register(character)
	defer engine.Unregister(eventChannel)

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
			case event := <-*eventChannel:
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
					loc := room.Location.Next(arg)
					roomToSee, found := engine.M.GetRoomByLocation(loc, zone.Id)
					if found {
						printLine(roomToSee.ToString(database.ReadMode, user.ColorMode, engine.M.CharactersIn(roomToSee, database.Character{})))
					} else {
						printLine("Nothing to see")
					}
				}
			}

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
			conn.Close()
			panic("User quit")

		default:
			direction := database.StringToDirection(action)

			if direction != database.DirectionNone {
				if room.HasExit(direction) {
					newRoom, err := engine.MoveCharacter(character, direction)
					if err == nil {
						room = newRoom
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
						room.Title = input
						engine.M.UpdateRoom(room)
					}
					printRoomEditor()

				case "2":
					input = getUserInput(RawUserInput, "Enter new description: ")

					if input != "" {
						room.Description = input
						engine.M.UpdateRoom(room)
					}
					printRoomEditor()

				case "3":
					for {
						menu := getToggleExitMenu(user.ColorMode, room)
						choice := ""

						for {
							menu.Print(conn, user.ColorMode)
							choice = getUserInput(CleanUserInput, menu.Prompt)
							if menu.HasAction(choice) || choice == "" {
								break
							}
						}

						if choice == "" {
							break
						}

						direction := database.StringToDirection(choice)
						if direction != database.DirectionNone {
							enable := !room.HasExit(direction)
							room.SetExitEnabled(direction, enable)
							engine.M.UpdateRoom(room)
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
			printLine(fmt.Sprintf("%v", room.Location))

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
					startX = room.Location.X - radius
					startY = room.Location.Y - radius
					startZ = room.Location.Z
					endX = startX + (radius * 2)
					endY = startY + (radius * 2)
					endZ = room.Location.Z
				} else if args[0] == "all" {
					topLeft, bottomRight := engine.ZoneCorners(zone.Id)

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
			builder.setUserRoom(room)

			for z := startZ; z <= endZ; z += 1 {
				for y := startY; y <= endY; y += 1 {
					for x := startX; x <= endX; x += 1 {
						loc := database.Coordinate{x, y, z}
						currentRoom, found := engine.M.GetRoomByLocation(loc, zone.Id)

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
				if zone.Id == "" {
					printLine("Currently in the null zone")
				} else {
					printLine("Current zone: " + utils.Colorize(user.ColorMode, utils.ColorBlue, zone.Name))
				}
			} else if len(args) == 1 {
				if args[0] == "list" {
					for _, zone := range engine.M.GetZones() {
						printLine(zone.Name)
					}
				}
			} else if len(args) == 2 {
				if args[0] == "rename" {
					_, found := engine.M.GetZoneByName(args[0])

					if found {
						printError("A zone with that name already exists")
						return
					}

					if zone.Id == "" {
						zone = database.NewZone(args[1])
						engine.M.UpdateZone(zone)
						engine.MoveRoomsToZone("", zone.Id)
					} else {
						zone.Name = args[1]
						engine.M.UpdateZone(zone)
					}
				} else if args[0] == "new" {
					_, found := engine.M.GetZoneByName(args[0])

					if found {
						printError("A zone with that name already exists")
						return
					}

					newZone := database.NewZone(args[1])
					engine.M.UpdateZone(newZone)

					zone = newZone
					newRoom := database.NewRoom(newZone.Id)

					engine.M.UpdateRoom(newRoom)

					engine.MoveCharacterToRoom(character, newRoom)

					room = newRoom
					printRoom()
				}
			}

		case "message":
			fallthrough
		case "m":
			if len(args) == 0 {
				printError("Nothing to say")
			} else {
				engine.BroadcastMessage(*character, strings.Join(args, " "))
			}

		case "say":
			fallthrough
		case "s":
			if len(args) == 0 {
				printError("Nothing to say")
			} else {
				engine.Say(*character, strings.Join(args, " "))
			}

		case "teleport":
			fallthrough
		case "tel":
			telUsage := func() {
				printError("Usage: /teleport <X> <Y> <Z>")
			}

			if len(args) != 3 {
				telUsage()
				return
			}

			x, err := strconv.Atoi(args[0])

			if err != nil {
				telUsage()
				return
			}

			y, err := strconv.Atoi(args[1])

			if err != nil {
				telUsage()
				return
			}

			z, err := strconv.Atoi(args[2])

			if err != nil {
				telUsage()
				return
			}

			newRoom, err := engine.MoveCharacterToLocation(character, database.Coordinate{X: x, Y: y, Z: z})

			if err == nil {
				room = newRoom
				printRoom()
			} else {
				printError(err.Error())
			}

		case "who":
			chars := engine.M.GetOnlineCharacters()

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
					engine.M.UpdateUser(*user)
					printLine("Color mode set to: None")
				case "light":
					user.ColorMode = utils.ColorModeLight
					engine.M.UpdateUser(*user)
					printLine("Color mode set to: Light")
				case "dark":
					user.ColorMode = utils.ColorModeDark
					engine.M.UpdateUser(*user)
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
					loc := room.Location.Next(direction)
					roomToDelete, found := engine.M.GetRoomByLocation(loc, zone.Id)
					if found {
						engine.DeleteRoom(roomToDelete)
					} else {
						printError("No room in that direction")
					}
				}
			} else {
				printError("Usage: /delete <direction>")
			}

		default:
			printError("Unrecognized command: \"" + command + "\"")
		}
	}

	printLineColor(utils.ColorWhite, "Welcome, "+character.PrettyName())
	printRoom()

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
