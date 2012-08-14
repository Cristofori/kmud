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
        fmt.Printf("Processing command: %v\n", command)

        switch command {
        case "?":
            fallthrough
        case "help":
        case "dig":
        case "edit":
            // Enter edit mode (show room with [] markers indicating portions that can be modified)
        default:
            io.WriteString(conn, "Unrecognized command")
        }
    }

	utils.WriteLine(conn, "Welcome, "+utils.FormatName(character))

	utils.WriteLine(conn, room.ToString())

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
				io.WriteString(conn, room.ToString())
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
