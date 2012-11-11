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
		message := ""

		switch event.Type() {
		case engine.MessageEventType:
			messageEvent := event.(engine.MessageEvent)
			message = utils.Colorize(user.ColorMode, utils.ColorBlue, messageEvent.Character.PrettyName()+": ") +
				utils.Colorize(user.ColorMode, utils.ColorWhite, messageEvent.Message)
		case engine.EnterEventType:
			enterEvent := event.(engine.EnterEvent)
			if enterEvent.RoomId == room.Id && enterEvent.Character.Id != character.Id {
				message = event.ToString()
			}
		case engine.LeaveEventType:
			moveEvent := event.(engine.LeaveEvent)
			if moveEvent.RoomId == room.Id && moveEvent.Character.Id != character.Id {
				message = event.ToString()
			}
		case engine.RoomUpdateEventType:
			roomEvent := event.(engine.RoomUpdateEvent)
			if roomEvent.Room.Id == room.Id {
				message = event.ToString()
				room = roomEvent.Room
			}
		case engine.LoginEventType:
			loginEvent := event.(engine.LoginEvent)
			if loginEvent.Character.Id != character.Id {
				message = event.ToString()
			}
		case engine.LogoutEventType:
			message = event.ToString()

		default:
			panic(fmt.Sprintf("Unhandled event: %v", event))
		}

		return message
	}

	eventChannel := engine.Register(*character)
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
							menu.Print(conn)
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
			width := 20 // Should be even
			if len(args) > 0 {
				var err error
				width, err = strconv.Atoi(args[0])

				if err != nil {
					printError("Invalid number given")
					return
				}
			}

			builder := newMapBuilder(width, width)

			startX := room.Location.X - (width / 2) + 1
			startY := room.Location.Y - (width / 2) + 1
			endX := startX + width
			endY := startY + width

			z := room.Location.Z

			for y := startY; y < endY; y += 1 {
				for x := startX; x < endX; x += 1 {
					loc := database.Coordinate{x, y, z}
					currentRoom, found := engine.GetRoomByLocation(loc)

					if found {
						// Translate to 0-based coordinates and double the coordinate
						// space to leave room for the exit lines
						builder.AddRoom(currentRoom, (x-startX)*2, (y-startY)*2)
					}
				}
			}

			printString(builder.toString(user.ColorMode))

		case "message":
			if len(args) == 0 {
				printLine("Nothing to say")
			} else {
				engine.BroadcastMessage(*character, strings.Join(args, " "))
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
				switch args[0] {
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
						room.SetExitEnabled(direction, false)
						engine.UpdateRoom(room)
					} else {
						printError("No room in that direction")
					}
				}
			} else {
				printError("Usage: /delete <direction>")
			}

		default:
			printError("Unrecognized command")
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
			prompt := <-promptChannel
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
		input := getUserInput(CleanUserInput, prompt())
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
