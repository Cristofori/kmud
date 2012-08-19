package game

import (
	"io"
	"labix.org/v2/mgo"
	"mud/database"
	"mud/utils"
	"net"
	"strings"
)

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
				input := utils.GetUserInput(conn, "Select a section to edit> ")

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
