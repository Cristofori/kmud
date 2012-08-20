package game

import (
	"io"
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
			io.WriteString(conn, room.ToString(database.EditMode))

			for {
				input := utils.GetUserInput(conn, "\nSelect a section to edit> ")

				switch input {
				case "":
					utils.WriteLine(conn, room.ToString(database.ReadMode))
					return
				case "1":
					input = utils.GetRawUserInput(conn, "Enter new title: ")
					room.Title = input
					database.SetRoomTitle(session, room.Id, input)
					utils.WriteLine(conn, room.ToString(database.EditMode))
				case "2":
					input = utils.GetRawUserInput(conn, "Enter new description: ")
					room.Description = input
					database.SetRoomDescription(session, room.Id, input)
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

					io.WriteString(conn, room.ToString(database.EditMode))

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
			io.WriteString(conn, "Unrecognized command")
		}
	}

	utils.WriteLine(conn, "Welcome, "+utils.FormatName(character))
	io.WriteString(conn, room.ToString(database.ReadMode))

	for {
		utils.PanicIfError(err)

		input := utils.GetUserInput(conn, "\n> ")

		if strings.HasPrefix(input, "/") {
			processCommand(input[1:len(input)])
		} else {
			switch input {
			case "quit":
				fallthrough
			case "exit":
				utils.WriteLine(conn, "Goodbye")
				conn.Close()
				panic("User quit")
			case "l":
				io.WriteString(conn, room.ToString(database.ReadMode))
			case "i":
				io.WriteString(conn, "You aren't carrying anything")
			default:
				exit := database.StringToDirection(input)
				if room.HasExit(exit) {
					database.SetCharacterRoom(session, character, room.ExitId(exit))
				} else {
					io.WriteString(conn, "You can't do that")
				}
			}
		}
	}
}

// vim: nocindent
