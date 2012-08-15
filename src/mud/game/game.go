package game

import (
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"mud/database"
	"mud/utils"
	"net"
	"strings"
)

func Exec(session *mgo.Session, conn net.Conn, character string) {

	room, err := database.GetCharacterRoom(session, character)

	processCommand := func(session *mgo.Session, conn net.Conn, command string) {

		switch command {
		case "?":
			fallthrough
		case "help":
		case "dig":
		case "edit":
			io.WriteString(conn, room.ToString(database.EditMode))

			for {
				input := utils.GetUserInput(conn, "Select a section to edit> ")

                fmt.Printf( "Input: %v\n", input)

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
            input := utils.GetUserInput(conn, "Are you sure (deletes all rooms and starts from scratch)? ")
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
			processCommand(session, conn, input[1:len(input)])
		} else {
			switch input {
			case "quit":
				fallthrough
			case "exit":
				utils.WriteLine(conn, "Goodbye")
				conn.Close()
			case "l":
				io.WriteString(conn, room.ToString(database.ReadMode))
			case "i":
				io.WriteString(conn, "You aren't carrying anything")
			default:
				if room.HasExit(input) {
					database.SetCharacterRoom(session, character, room.ExitId(input))
				} else {
					io.WriteString(conn, "You can't do that")
				}
			}
		}
	}
}

// vim: nocindent
