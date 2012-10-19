package game

import (
	"labix.org/v2/mgo"
	"mud/database"
	"mud/utils"
	"net"
	"strings"
)

func toggleExit(room database.Room, dir database.ExitDirection) {
}

func getToggleExitMenu(room database.Room) utils.Menu {

	onOrOff := func(on bool) string {
		if on {
			return "On"
		}

		return "Off"
	}

	menu := utils.NewMenu("Edit Exits")

	menu.AddAction("n", "[N]orth: "+onOrOff(room.HasExit(database.North)))
	menu.AddAction("e", "[E]ast: "+onOrOff(room.HasExit(database.East)))
	menu.AddAction("s", "[S]outh: "+onOrOff(room.HasExit(database.South)))
	menu.AddAction("w", "[W]est: "+onOrOff(room.HasExit(database.West)))
	menu.AddAction("u", "[U]p: "+onOrOff(room.HasExit(database.Up)))
	menu.AddAction("d", "[D]own: "+onOrOff(room.HasExit(database.Down)))

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
						database.SetRoomTitle(session, room.Id, input)
					}
					utils.WriteLine(conn, room.ToString(database.EditMode))

				case "2":
					input = utils.GetRawUserInput(conn, "Enter new description: ")

					if input != "" {
						room.Description = input
						database.SetRoomDescription(session, room.Id, input)
					}
					utils.WriteLine(conn, room.ToString(database.EditMode))

				case "3":
					for {
						done := false
						menu := getToggleExitMenu(room)
						choice, _ := menu.Exec(conn)

						switch choice {
						case "n":
							toggleExit(room, database.North)
						case "e":
							toggleExit(room, database.East)
						case "s":
							toggleExit(room, database.South)
						case "w":
							toggleExit(room, database.West)
						case "u":
							toggleExit(room, database.Up)
						case "d":
							toggleExit(room, database.Down)
						case "":
							done = true
						}

						if done {
							break
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
			case "n":
				fallthrough
			case "e":
				fallthrough
			case "s":
				fallthrough
			case "w":
				fallthrough
			case "u":
				fallthrough
			case "d":
				utils.WriteLine(conn, "You can't go that way")

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
				utils.WriteLine(conn, "Goodbye")
				conn.Close()
				panic("User quit")

			default:
				exit := database.StringToDirection(input)
				if room.HasExit(exit) {
					database.SetCharacterRoom(session, character, room.ExitId(exit))
				} else {
					utils.WriteLine(conn, "You can't do that")
				}
			}
		}
	}
}

// vim: nocindent
