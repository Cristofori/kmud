package game

import (
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"mud/database"
	"mud/utils"
	"net"
)

type gameState struct {
}

func Exec(session *mgo.Session, conn net.Conn, character string) {
	utils.WriteLine(conn, "Welcome, "+utils.FormatName(character))
	for {
		location, err := database.GetCharacterLocation(session, character)

		if err != nil {
			fmt.Printf("Database error: %s\n", err.Error())
			break
		}

		utils.WriteLine(conn, "Location: "+location)
		line, err := utils.GetUserInput(conn, "\n> ")

		if err != nil {
			fmt.Printf("Lost connection to user %v\n", character)
			break
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
