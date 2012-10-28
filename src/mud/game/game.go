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

func Exec(conn net.Conn, character database.Character) {
	room := engine.GetCharacterRoom(character)

	printString := func(data string) {
		io.WriteString(conn, data)
	}

	printLine := func(line string) {
		utils.WriteLine(conn, line)
	}

	printLineColor := func(color utils.Color, line string) {
		printLine(utils.Colorize(utils.ColorModeNormal, color, line))
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
						// Translate to 0-based coordinates
						builder.AddRoom(currentRoom, (x-startX)*2, (y-startY)*2)
					}
				}
			}

			printString(builder.ToString())

		case "message":
			if len(args) == 0 {
				printLine("Nothing to say")
			} else {
				engine.BroadcastMessage(character, strings.Join(args, " "))
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

		default:
			printLine("Unrecognized command")
		}
	}

	printString(utils.ColorNormal)
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

type mapBuilder struct {
	width  int
	height int
	data   [][]mapTile
}

type mapTile struct {
	char rune
}

func newMapBuilder(width int, height int) mapBuilder {
	var builder mapBuilder

	// Double the size to account for extra space to draw exits
	width *= 2
	height *= 2

	builder.data = make([][]mapTile, height)

	for y := 0; y < height; y += 1 {
		builder.data[y] = make([]mapTile, width)
	}

	builder.width = width
	builder.height = height

	for y := 0; y < height; y += 1 {
		for x := 0; x < width; x += 1 {
			builder.data[y][x].char = ' '
		}
	}

	return builder
}

func (self *mapBuilder) AddRoom(room database.Room, x int, y int) {
	addIfExists := func(dir database.ExitDirection, x int, y int) {
		if x < 0 || y < 0 {
			return
		}

		if room.HasExit(dir) {
			self.data[y][x].AddExit(dir)
		}
	}

	centerX := (self.width / 2) - 2
	centerY := (self.height / 2) - 2

	if x == centerX && y == centerY {
		self.data[y][x].char = 'O'
	} else {
		self.data[y][x].char = '#'
	}

	addIfExists(database.DirectionNorth, x, y-1)
	addIfExists(database.DirectionNorthEast, x+1, y-1)
	addIfExists(database.DirectionEast, x+1, y)
	addIfExists(database.DirectionSouthEast, x+1, y+1)
	addIfExists(database.DirectionSouth, x, y+1)
	addIfExists(database.DirectionSouthWest, x-1, y+1)
	addIfExists(database.DirectionWest, x-1, y)
	addIfExists(database.DirectionNorthWest, x-1, y-1)
}

func (self *mapBuilder) ToString() string {
	var rows []string

	for y := 0; y < self.height; y += 1 {
		row := ""
		for x := 0; x < self.width; x += 1 {
			char := self.data[y][x].char
			row = row + string(char)
		}
		rows = append(rows, row)
	}
	rows = trim(rows)

	str := ""
	for _, row := range rows {
		str = str + row + "\n"
	}
	return str
}

func trim(rows []string) []string {
	rowEmpty := func(row string) bool {
		for _, char := range row {
			if char != ' ' {
				return false
			}
		}
		return true
	}

	// Trim from the top
	for _, row := range rows {
		if !rowEmpty(row) {
			break
		}

		rows = rows[1:]
	}

	// Trim from the bottom
	for i := len(rows) - 1; i >= 0; i -= 1 {
		row := rows[i]
		if !rowEmpty(row) {
			break
		}
		rows = rows[:len(rows)-1]
	}

	return rows
}

func (self *mapTile) AddExit(dir database.ExitDirection) {

	combineChars := func(r1 rune, r2 rune, r3 rune) {
		if self.char == r1 {
			self.char = r2
		} else {
			self.char = r3
		}
	}

	switch dir {
	case database.DirectionNorth:
		combineChars('v', '|', '^')
	case database.DirectionNorthEast:
		combineChars('\\', 'X', '/')
	case database.DirectionEast:
		combineChars('<', '-', '>')
	case database.DirectionSouthEast:
		combineChars('/', 'X', '\\')
	case database.DirectionSouth:
		combineChars('^', '|', 'v')
	case database.DirectionSouthWest:
		combineChars('\\', 'X', '/')
	case database.DirectionWest:
		combineChars('>', '-', '<')
	case database.DirectionNorthWest:
		combineChars('/', 'X', '\\')
	default:
		panic("Unexpected direction given to mapTile::AddExit()")
	}
}

// vim: nocindent
