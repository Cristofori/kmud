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

func processCommand(session *mgo.Session, conn net.Conn, command string ) {
    fmt.Printf("Processing command: %v\n", command)

    switch command {
        case "new":
        case "edit":
            // Enter edit mode (show room with [] markers indicating portions that can be modified)
    }
}

func Exec(session *mgo.Session, conn net.Conn, character string) {
	utils.WriteLine(conn, "Welcome, "+utils.FormatName(character))
	for {
		room, err := database.GetCharacterRoom(session, character)
        utils.PanicIfError(err)

		utils.WriteLine(conn, room.ToString())
		input := utils.GetUserInput(conn, "\n> ")

        if strings.HasPrefix(input, "/") {
            processCommand( session, conn, input[1:len(input)] )
        }

		if input == "quit" || input == "exit" {
			utils.WriteLine(conn, "Goodbye")
			conn.Close()
			break
		}

        if room.HasExit(input) {
            database.SetCharacterRoom(session, character, room.ExitId(input))
        } else {
            io.WriteString(conn, "You can't do that.")
        }
	}
}

// vim: nocindent
