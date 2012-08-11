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

    if command == "newroom" {
    }
}

func Exec(session *mgo.Session, conn net.Conn, character string) {
	utils.WriteLine(conn, "Welcome, "+utils.FormatName(character))
	for {
		room, err := database.GetCharacterRoom(session, character)

		if err != nil {
			fmt.Printf("Database error: %s\n", err.Error())
			break
		}

		utils.WriteLine(conn, room.ToString())
		line, err := utils.GetUserInput(conn, "\n> ")

		if err != nil {
			fmt.Printf("Lost connection to user %v\n", character)
			break
		}

        if strings.HasPrefix(line, "/") {
            processCommand( session, conn, line[1:len(line)] )
        }

		if line == "quite" || line == "exit" {
			utils.WriteLine(conn, "Goodbye")
			conn.Close()
			break
		}

		io.WriteString(conn, line)
	}
}

// vim: nocindent
