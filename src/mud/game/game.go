package game

import (
	"fmt"
	"io"
	"mud/database"
	"mud/engine"
	"mud/utils"
	"net"
	"strings"
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
	menu.AddAction("ne", "[NE]North East"+onOrOff(database.DirectionNorthEast))
	menu.AddAction("e", "[E]ast: "+onOrOff(database.DirectionEast))
	menu.AddAction("se", "[SE]South East"+onOrOff(database.DirectionSouthEast))
	menu.AddAction("s", "[S]outh: "+onOrOff(database.DirectionSouth))
	menu.AddAction("sw", "[SW]South West"+onOrOff(database.DirectionSouthWest))
	menu.AddAction("w", "[W]est: "+onOrOff(database.DirectionWest))
	menu.AddAction("nw", "[NW]North West"+onOrOff(database.DirectionNorthWest))
	menu.AddAction("u", "[U]p: "+onOrOff(database.DirectionUp))
	menu.AddAction("d", "[D]own: "+onOrOff(database.DirectionDown))

	return menu
}

func Exec(conn net.Conn, character database.Character) {
	room := engine.GetCharacterRoom(character)

	printString := func(data string) {
		io.WriteString(conn, data)
	}

	printLine := func(line string) {
		utils.WriteLine(conn, line)
	}

	printRoom := func() {
		charList := engine.CharactersIn(room, character)
		printLine(room.ToString(database.ReadMode, charList))
	}

	printRoomEditor := func() {
		printLine(room.ToString(database.EditMode, nil))
	}

	prompt := func() string {
		return "> "
	}

	processEvent := func(event engine.Event) string {
		message := ""

		switch event.Type() {
		case engine.MessageEventType:
			message = event.ToString()
		case engine.EnterEventType:
			enterEvent := event.(engine.EnterEvent)
			if enterEvent.RoomId == room.Id && enterEvent.Character != character {
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
						printLine(roomToSee.ToString(database.ReadMode, engine.CharactersIn(roomToSee, database.Character{})))
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
					character, newRoom, err = engine.MoveCharacter(character, direction)
					if err == nil {
						room = newRoom
						printRoom()
					} else {
						printLine(err.Error())
					}

				} else {
					printLine("You can't go that way")
				}
			} else {
				printLine("You can't do that")
			}
		}
	}

	processCommand := func(command string, args []string) {
		switch command {
		case "?":
			fallthrough
		case "help":
		case "dig":
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

			room = engine.GetCharacterRoom(character)
			printRoom()

		case "loc":
			fallthrough
		case "location":
			printLine(fmt.Sprintf("%v", room.Location))

		case "map":
			width := 20 // Should be even

			startX := room.Location.X - (width / 2)
			startY := room.Location.Y - (width / 2)
			endX := startX + width
			endY := startY + width

			z := room.Location.Z

			for y := startY; y < endY; y += 1 {
				exitRow := ""
				printString("\n")
				for x := startX; x < endX; x += 1 {
					loc := database.Coordinate{x, y, z}

					currentRoom, currentFound := engine.GetRoomByLocation(loc)
					eastRoom, eastFound := engine.GetRoomByLocation(loc.Next(database.DirectionEast))
					southRoom, southFound := engine.GetRoomByLocation(loc.Next(database.DirectionSouth))

					if currentFound {
						if currentRoom == room {
							printString("O")
						} else {
							printString("#")
						}

						if eastFound {
							if currentRoom.HasExit(database.DirectionEast) {
								if eastRoom.HasExit(database.DirectionWest) {
									printString("-")
								} else {
									printString(">")
								}
							} else if eastRoom.HasExit(database.DirectionWest) {
								printString("<")
							}
						} else {
							printString(" ")
						}

						if southFound {
							if currentRoom.HasExit(database.DirectionSouth) {
								if southRoom.HasExit(database.DirectionNorth) {
									exitRow = exitRow + "|"
								} else {
									exitRow = exitRow + "v"
								}
							} else {
								exitRow = exitRow + "^"
							}
						} else {
							exitRow = exitRow + " "
						}
					} else {
						printString("  ")
						exitRow = exitRow + " "
					}
					exitRow = exitRow + " "
				}
				printString("\n" + exitRow)
			}
			printString("\n")

		case "message":
			if len(args) == 0 {
				printLine("Nothing to say")
			} else {
				engine.BroadcastMessage(character, strings.Join(args, " "))
			}

		default:
			printLine("Unrecognized command")
		}
	}

	printLine("Welcome, " + utils.FormatName(character.Name))
	printRoom()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicChannel <- r
			}
		}()

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

			userInputChannel <- input
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
