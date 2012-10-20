package game

import (
	"labix.org/v2/mgo"
	"mud/database"
	"mud/utils"
	"net"
	"strings"
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
	menu.AddAction("e", "[E]ast: "+onOrOff(database.DirectionEast))
	menu.AddAction("s", "[S]outh: "+onOrOff(database.DirectionSouth))
	menu.AddAction("w", "[W]est: "+onOrOff(database.DirectionWest))
	menu.AddAction("u", "[U]p: "+onOrOff(database.DirectionUp))
	menu.AddAction("d", "[D]own: "+onOrOff(database.DirectionDown))

	return menu
}

func Exec(session *mgo.Session, conn net.Conn, character string) {

	room, err := database.GetCharacterRoom(session, character)

	processCommand := func(command string) {

		switch command {
		case "?":
			fallthrough
		case "help":
		case "dig":
		case "edit":
			utils.WriteLine(conn, room.ToString(database.EditMode))

			for {
				input := utils.GetUserInput(conn, "Select a section to edit> ")

				switch input {
				case "":
					utils.WriteLine(conn, room.ToString(database.ReadMode))
					return

				case "1":
					input = utils.GetRawUserInput(conn, "Enter new title: ")

					if input != "" {
						room.Title = input
						database.SetRoomTitle(session, room, input)
					}
					utils.WriteLine(conn, room.ToString(database.EditMode))

				case "2":
					input = utils.GetRawUserInput(conn, "Enter new description: ")

					if input != "" {
						room.Description = input
						database.SetRoomDescription(session, room, input)
					}
					utils.WriteLine(conn, room.ToString(database.EditMode))

				case "3":
					for {
						menu := getToggleExitMenu(room)
						choice, _ := menu.Exec(conn)

						toggleExit := func(direction database.ExitDirection) {
							enable := !room.HasExit(direction)
							room.SetExitEnabled(direction, enable)
							database.CommitRoom(session, room)
						}

						if choice == "" {
							break
						} else {
							direction := database.StringToDirection(choice)
							if direction != database.DirectionNone {
								toggleExit(direction)
							}
						}
					}

					utils.WriteLine(conn, room.ToString(database.EditMode))

				default:
					utils.WriteLine(conn, "Invalid selection")
				}
			}

		case "rebuild":
			input := utils.GetUserInput(conn, "Are you sure (delete all rooms and starts from scratch)? ")
			if input[0] == 'y' || input == "yes" {
				database.GenerateDefaultMap(session)
			}
			room, err = database.GetCharacterRoom(session, character)

		default:
			utils.WriteLine(conn, "Unrecognized command")
		}
	}

	utils.WriteLine(conn, "Welcome, "+utils.FormatName(character))
	utils.WriteLine(conn, room.ToString(database.ReadMode))

	for {
		utils.PanicIfError(err)

		input := utils.GetUserInput(conn, "> ")

		if strings.HasPrefix(input, "/") {
			processCommand(input[1:len(input)])
		} else {
			switch input {
			case "l":
				utils.WriteLine(conn, room.ToString(database.ReadMode))

			case "i":
				utils.WriteLine(conn, "You aren't carrying anything")

			case "":
				fallthrough
			case "logout":
				return

			case "quit":
				fallthrough
			case "exit":
				utils.WriteLine(conn, "Take luck!")
				conn.Close()
				panic("User quit")

			default:
				exit := database.StringToDirection(input)

				if exit != database.DirectionNone {
					if room.HasExit(exit) {
						utils.WriteLine(conn, "TODO: Implement moving into the next room") // TODO
					} else {
						utils.WriteLine(conn, "You can't go that way")
					}
				} else {
					utils.WriteLine(conn, "You can't do that")
				}
			}
		}
	}
}

// vim: nocindent
