package game

import (
"io"
    "mud/database"
    "mud/utils"
    "labix.org/v2/mgo"
    "net"
    "fmt"
)

type gameState struct {
}

func Exec(session *mgo.Session, conn net.Conn, user string ) {
    utils.WriteLine(conn, "Welcome")
    for {
        location, err := database.GetUserLocation(session, user)

        if err != nil {
            fmt.Printf("Database error: %s\n", err.Error())
            break
        }

        utils.WriteLine(conn, "Location: " + location)
        line, err := utils.GetUserInput(conn, "\n> ")

        if err != nil {
            fmt.Printf("Lost connection to user %v\n", user)
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
