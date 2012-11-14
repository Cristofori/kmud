package game

import (
	"fmt"
	"io"
	"kmud/database"
	"kmud/engine"
	"kmud/utils"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type userInputMode int

const (
	CleanUserInput userInputMode = iota
	RawUserInput   userInputMode = iota
)

func getToggleExitMenu(room database.Room) utils.Menu {
	onOrOff := func(direction database.ExitDirection) string {

		if room.HasExit(direction) {
			return "On"
		}

		return "Off"
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
	room := engine.GetCharacterRoom(*character)

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
		charList := engine.CharactersIn(room, *character)
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
					roomToSee, found := engine.GetRoomByLocation(loc)
					if found {
						printLine(roomToSee.ToString(database.ReadMode, user.ColorMode, engine.CharactersIn(roomToSee, database.Character{})))
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
					var newRoom database.Room
					var err error
					newRoom, err = engine.MoveCharacter(character, direction)
					if err == nil {
						room = newRoom
						printRoom()
					} else {
						printLine(err.Error())
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
						engine.UpdateRoom(room)
					}
					printRoomEditor()

				case "2":
					input = getUserInput(RawUserInput, "Enter new description: ")

					if input != "" {
						room.Description = input
						engine.UpdateRoom(room)
					}
					printRoomEditor()

				case "3":
					for {
						menu := getToggleExitMenu(room)
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
							engine.UpdateRoom(room)
						}
					}

					printRoomEditor()

				default:
					printLine("Invalid selection")
				}
			}

		case "rebuild":
			input := getUserInput(CleanUserInput, "Are you sure (delete all rooms and starts from scratch)? ")
			if input[0] == 'y' || input == "yes" {
				engine.GenerateDefaultMap()
			}

			room = engine.GetCharacterRoom(*character)
			printRoom()

		case "loc":
			fallthrough
		case "location":
			printLine(fmt.Sprintf("%v", room.Location))

		case "map":
			mapUsage := func() {
				printError("Usage: /map [<radius>|save <map name>]")
			}

			name := ""

			startX := 0
			startY := 0
			endX := 0
			endY := 0

			showUserRoom := false

			if len(args) < 2 {
				radius := 0
				if len(args) == 0 {
					radius = 10
				} else if len(args) == 1 {
					var err error
					radius, err = strconv.Atoi(args[0])

					if err != nil || radius < 1 {
						mapUsage()
						return
					}
				}

				showUserRoom = true

				startX = room.Location.X - radius
				startY = room.Location.Y - radius
				endX = startX + (radius * 2)
				endY = startY + (radius * 2)
			} else if len(args) >= 2 {
				if args[0] == "save" {
					name = strings.Join(args[1:], "_")
					topLeft, bottomRight := engine.MapCorners()

					startX = topLeft.X
					startY = topLeft.Y
					endX = bottomRight.X
					endY = bottomRight.Y
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

			builder := newMapBuilder(width, height)

			if showUserRoom {
				builder.setUserRoom(room)
			}

			z := room.Location.Z

			for y := startY; y <= endY; y += 1 {
				for x := startX; x <= endX; x += 1 {
					loc := database.Coordinate{x, y, z}
					currentRoom, found := engine.GetRoomByLocation(loc)

					if found {
						// Translate to 0-based coordinates and double the coordinate
						// space to leave room for the exit lines
						builder.addRoom(currentRoom, (x-startX)*2, (y-startY)*2)
					}
				}
			}

			if name == "" {
				printString(builder.toString(user.ColorMode))
			} else {
				filename := name + ".map"
				file, err := os.Create(filename)

				if err != nil {
					printError(err.Error())
					return
				}
				defer file.Close()

				mapData := builder.toString(utils.ColorModeNone)
				_, err = file.WriteString(mapData)

				if err == nil {
					printLine("Map saved as: " + utils.Colorize(user.ColorMode, utils.ColorBlue, filename))
				} else {
					printError(err.Error())
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

		case "who":
			chars := engine.GetOnlineCharacters()

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
					engine.UpdateUser(*user)
					printLine("Color mode set to: None")
				case "light":
					user.ColorMode = utils.ColorModeLight
					engine.UpdateUser(*user)
					printLine("Color mode set to: Light")
				case "dark":
					user.ColorMode = utils.ColorModeDark
					engine.UpdateUser(*user)
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
					roomToDelete, found := engine.GetRoomByLocation(loc)
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
