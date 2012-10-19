package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"mud/database"
	"mud/game"
	"mud/utils"
	"net"
	"strconv"
)

func login(session *mgo.Session, conn net.Conn) string {

	for {
		line := utils.GetUserInput(conn, "Username: ")

		if line == "" {
			return ""
		}

		found, err := database.FindUser(session, line)
		utils.PanicIfError(err)

		if !found {
			utils.WriteLine(conn, "User not found")
		} else {
			return line
		}
	}

	panic("Unexpected code path")
	return ""
}

func newUser(session *mgo.Session, conn net.Conn) string {

	for {
		line := utils.GetUserInput(conn, "Desired username: ")
		err := database.NewUser(session, line)

		if err == nil {
			return line
		}

		utils.WriteLine(conn, err.Error())
	}

	panic("Unexpected code path")
	return ""
}

func newCharacter(session *mgo.Session, conn net.Conn, user string) string {
	// TODO: character slot limit
	for {
		line := utils.GetUserInput(conn, "Desired character name: ")

		if line == "" {
			return ""
		}

		err := database.NewCharacter(session, user, line)

		if err == nil {
			return line
		}

		utils.WriteLine(conn, err.Error())
	}

	panic("Unexpected code path")
	return ""
}

func quit(session *mgo.Session, conn net.Conn) error {
	utils.WriteLine(conn, "Goodbye!")
	conn.Close()
	return nil
}

func mainMenu() utils.Menu {

	menu := utils.NewMenu("MUD")

	menu.AddAction("l", "[L]ogin")
	menu.AddAction("n", "[N]ew user")
	menu.AddAction("q", "[Q]uit")

	return menu
}

func userMenu(session *mgo.Session, user string) utils.Menu {

	chars, _ := database.GetUserCharacters(session, user)

	menu := utils.NewMenu(user)
	menu.AddAction("l", "[L]ogout")
	menu.AddAction("a", "[A]dmin")
	menu.AddAction("n", "[N]ew character")
	if len(chars) > 0 {
		menu.AddAction("d", "[D]elete character")
	}

	for i, char := range chars {
		indexStr := strconv.Itoa(i + 1)
		actionText := fmt.Sprintf("[%v]%v", indexStr, utils.FormatName(char))
		menu.AddActionData(indexStr, actionText, char)
	}

	return menu

}

func deleteMenu(session *mgo.Session, user string) utils.Menu {
	chars, _ := database.GetUserCharacters(session, user)

	menu := utils.NewMenu("Delete character")

	menu.AddAction("c", "[C]ancel")

	for i, char := range chars {
		indexStr := strconv.Itoa(i + 1)
		actionText := fmt.Sprintf("[%v]%v", indexStr, utils.FormatName(char))
		menu.AddActionData(indexStr, actionText, char)
	}

	return menu
}

func handleConnection(session *mgo.Session, conn net.Conn) {

	defer conn.Close()
	defer session.Close()

	user := ""
	character := ""

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Lost connection to client (%v/%v): %v, %v\n",
				utils.FormatName(user),
				utils.FormatName(character),
				conn.RemoteAddr(),
				r)
		}
	}()

	for {
		if user == "" {
			menu := mainMenu()
			choice, _ := menu.Exec(conn)

			switch choice {
			case "l":
				user = login(session, conn)
			case "n":
				user = newUser(session, conn)
			case "":
				fallthrough
			case "q":
				quit(session, conn)
				return
			}
		} else if character == "" {
			menu := userMenu(session, user)
			choice, charName := menu.Exec(conn)

			switch choice {
			case "":
				fallthrough
			case "l":
				user = ""
			case "a":
				utils.WriteLine(conn, "*** Admin menu goes here") // TODO
			case "n":
				character = newCharacter(session, conn, user)
			case "d":
				deleteMenu := deleteMenu(session, user)
				deleteChoice, deleteCharName := deleteMenu.Exec(conn)

				_, err := strconv.Atoi(deleteChoice)

				if err == nil {
					database.DeleteCharacter(session, user, deleteCharName)
				}

			default:
				_, err := strconv.Atoi(choice)

				if err == nil {
					character = charName
				}
			}
		} else {
			game.Exec(session, conn, character)
			character = ""
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

	fmt.Println("Server listening on port 8945")

	for {
		conn, err := listener.Accept()
		utils.HandleError(err)
		fmt.Printf("Client connected: %v\n", conn.RemoteAddr())
		go handleConnection(session.Copy(), conn)
	}
}

// vim: nocindent