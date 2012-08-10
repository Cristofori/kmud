package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"mud/database"
    "mud/game"
    "mud/utils"
	"net"
)

type Menu struct {
	Actions map[string]bool
	Text    string
}

func NewMenu() Menu {
	var menu Menu
	menu.Actions = map[string]bool{}
	return menu
}

func (self *Menu) Exec(session *mgo.Session, conn net.Conn) (string, error) {

	for {
		input, err := utils.GetUserInput(conn, self.Text)

		if err != nil {
			return "", err
		}

		if self.Actions[input] {
            return input, nil
        }
	}

    panic("Unexpected code path")
	return "", nil
}

func login(session *mgo.Session, conn net.Conn) (string, error) {

	for {
		line, err := utils.GetUserInput(conn, "Username: ")

		if err != nil {
			return "", err
		}

		found, err := database.FindUser(session, line)

		if err != nil {
			return "", err
		}

		if !found {
			utils.WriteLine(conn, "User not found")
		} else {
			return line, nil
		}
	}

    panic("Unexpected code path")
	return "", nil
}

func newUser(session *mgo.Session, conn net.Conn) (string, error) {

	for {
		line, err := utils.GetUserInput(conn, "Desired username: ")

		if err != nil {
			return "", err
		}

		err = database.NewUser(session, line)
		if err == nil {
			return line, nil
		}

		utils.WriteLine(conn, err.Error())
	}

    panic("Unexpected code path")
	return "", nil
}

func quit(session *mgo.Session, conn net.Conn) error {
	utils.WriteLine(conn, "Goodbye!")
	conn.Close()
	return nil
}

func mainMenu() Menu {

	menu := NewMenu()

	menu.Text = `
-=-=- MUD -=-=-
  [L]ogin
  [N]ew user
  [A]bout
  [Q]uit
> `

	menu.Actions["l"] = true
	menu.Actions["n"] = true
	menu.Actions["q"] = true

	return menu
}

func handleConnection(session *mgo.Session, conn net.Conn) {

	defer conn.Close()
	defer session.Close()

    user := ""

	for {
		if user == "" {
			menu := mainMenu()
			choice, err := menu.Exec(session, conn)

            switch choice {
                case "l":
                    var err error
                    user, err = login(session, conn)
                    if err != nil {
                        return
                    }
                case "n":
                    var err error
                    user, err = newUser(session, conn)
                    if err != nil {
                        return
                    }
                case "q":
                    quit(session, conn)
                    return
            }

			if err != nil {
				return
			}
		} else {
            game.Exec(session, conn, user)
            user = ""
        }
	}
}

func main() {

	fmt.Printf("Connecting to database... ")
	session, err := mgo.Dial("localhost")

	utils.HandleError(err)

	fmt.Printf("done.\n")

	listener, err := net.Listen("tcp", ":8945")
	utils.HandleError(err)

	fmt.Printf("Server listening on port 8945\n")

	for {
		conn, err := listener.Accept()
		utils.HandleError(err)
		go handleConnection(session.Copy(), conn)
	}
}

// vim: nocindent
