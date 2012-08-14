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

func processCommand(session *mgo.Session, conn net.Conn, command string) {
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

func Exec(session *mgo.Session, conn net.Conn, character string) {
	utils.WriteLine(conn, "Welcome, "+utils.FormatName(character))

	room, err := database.GetCharacterRoom(session, character)
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
				utils.WriteLine(conn, room.ToString())
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
