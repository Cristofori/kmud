package server

import (
	"fmt"
	"labix.org/v2/mgo"
	"mud/database"
	"mud/engine"
	"mud/game"
	"mud/utils"
	"net"
	"runtime"
	"strconv"
)

func login(session *mgo.Session, conn net.Conn) database.User {
	for {
		line := utils.GetUserInput(conn, "Username: ")

		if line == "" {
			return database.User{}
		}

		user, err := engine.GetUserByName(line)

		if err != nil {
			utils.WriteLine(conn, err.Error())
		} else {
			err := engine.Login(user)

			if err == nil {
				return user
			} else {
				utils.WriteLine(conn, err.Error())
			}
		}
	}

	panic("Unexpected code path")
	return database.User{}
}

func newUser(session *mgo.Session, conn net.Conn) database.User {
	for {
		line := utils.GetUserInput(conn, "Desired username: ")

		var user database.User
		if line == "" {
			return user
		}

		user, err := engine.CreateUser(line)

		if err == nil {
			err = engine.Login(user)
			if err == nil {
				return user
			}
		}

		utils.WriteLine(conn, err.Error())
	}

	panic("Unexpected code path")
	return database.User{}
}

func newCharacter(session *mgo.Session, conn net.Conn, user *database.User) database.Character {
	// TODO: character slot limit
	for {
		line := utils.GetUserInput(conn, "Desired character name: ")

		if line == "" {
			return database.Character{}
		}

		character, err := engine.CreateCharacter(user, line)

		if err == nil {
			return character
		}

		utils.WriteLine(conn, err.Error())
	}

	panic("Unexpected code path")
	return database.Character{}
}

func quit(session *mgo.Session, conn net.Conn) {
	utils.WriteLine(conn, "Take luck!")
	conn.Close()
}

func mainMenu() utils.Menu {
	menu := utils.NewMenu("MUD")

	menu.AddAction("l", "[L]ogin")
	menu.AddAction("n", "[N]ew user")
	menu.AddAction("q", "[Q]uit")

	return menu
}

func userMenu(session *mgo.Session, user database.User) utils.Menu {
	chars := engine.GetCharacters(user)

	menu := utils.NewMenu(user.PrettyName())
	menu.AddAction("l", "[L]ogout")
	menu.AddAction("a", "[A]dmin")
	menu.AddAction("n", "[N]ew character")
	if len(chars) > 0 {
		menu.AddAction("d", "[D]elete character")
	}

	for i, char := range chars {
		indexStr := strconv.Itoa(i + 1)
		actionText := fmt.Sprintf("[%v]%v", indexStr, char.PrettyName())
		menu.AddActionData(indexStr, actionText, char.Id)
	}

	return menu
}

func deleteMenu(session *mgo.Session, user database.User) utils.Menu {
	chars := engine.GetCharacters(user)

	menu := utils.NewMenu("Delete character")

	menu.AddAction("c", "[C]ancel")

	for i, char := range chars {
		indexStr := strconv.Itoa(i + 1)
		actionText := fmt.Sprintf("[%v]%v", indexStr, char.PrettyName())
		menu.AddActionData(indexStr, actionText, char.Id)
	}

	return menu
}

func adminMenu(session *mgo.Session) utils.Menu {
	menu := utils.NewMenu("Admin")
	menu.AddAction("r", "[R]ebuild")
	return menu
}

func handleConnection(session *mgo.Session, conn net.Conn) {
	defer conn.Close()
	defer session.Close()

	var user database.User
	var character database.Character

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Lost connection to client (%v/%v): %v, %v\n",
				user.PrettyName(),
				character.PrettyName(),
				conn.RemoteAddr(),
				r)
		}
	}()

	for {
		if user.Name == "" {
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
		} else if character.Name == "" {
			menu := userMenu(session, user)
			choice, charId := menu.Exec(conn)

			switch choice {
			case "":
				fallthrough
			case "l":
				engine.Logout(user)
				user = database.User{}
			case "a":
				adminMenu := adminMenu(session)
				choice, _ := adminMenu.Exec(conn)
				if choice == "r" {
					engine.GenerateDefaultMap()
				}
			case "n":
				character = newCharacter(session, conn, &user)
			case "d":
				deleteMenu := deleteMenu(session, user)
				deleteChoice, deleteCharId := deleteMenu.Exec(conn)

				_, err := strconv.Atoi(deleteChoice)

				if err == nil {
					err = engine.DeleteCharacter(&user, deleteCharId)

					if err != nil {
						utils.WriteLine(conn, fmt.Sprintf("Error deleting character: %s", err))
					}
				}

			default:
				_, err := strconv.Atoi(choice)

				if err == nil {
					character = engine.GetCharacter(charId)
				}
			}
		} else {
			game.Exec(conn, &user, &character)
			character = database.Character{}
		}
	}
}

func Exec() {
	runtime.GOMAXPROCS(8)

	fmt.Printf("Connecting to database... ")
	session, err := mgo.Dial("localhost")

	utils.HandleError(err)

	fmt.Printf("done.\n")

	listener, err := net.Listen("tcp", ":8945")
	utils.HandleError(err)

	err = engine.StartUp(session.Copy())

	fmt.Println("Server listening on port 8945")

	for {
		conn, err := listener.Accept()
		utils.HandleError(err)
		fmt.Printf("Client connected: %v\n", conn.RemoteAddr())
		go handleConnection(session.Copy(), conn)
	}
}

// vim: nocindent
