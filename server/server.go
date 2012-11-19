package server

import (
	"fmt"
	"kmud/database"
	"kmud/engine"
	"kmud/game"
	"kmud/utils"
	"labix.org/v2/mgo"
	"net"
	"strconv"
)

func login(conn net.Conn) database.User {
	for {
		line := utils.GetUserInput(conn, "Username: ")

		if line == "" {
			return database.User{}
		}

		user, found := engine.M.GetUserByName(line)

		if !found {
			utils.WriteLine(conn, "User not found")
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

func newUser(conn net.Conn) database.User {
	for {
		name := utils.GetUserInput(conn, "Desired username: ")

		var user database.User
		if name == "" {
			return user
		}

		user, found := engine.M.GetUserByName(name)

		if found {
			utils.WriteLine(conn, "Name unavailable")
		} else {
			user = database.NewUser(name)
			engine.M.UpdateUser(user)

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

func newCharacter(conn net.Conn, user *database.User) database.Character {
	// TODO: character slot limit
	for {
		name := utils.GetUserInput(conn, "Desired character name: ")

		if name == "" {
			return database.Character{}
		}

		character, found := engine.M.GetCharacterByName(name)

		if found {
			utils.WriteLine(conn, "A character with that name already exists")
		} else {
			room := engine.M.GetRooms()[0] // TODO

			character = database.NewCharacter(name, user.Id, room.Id)
			engine.M.UpdateCharacter(character)
			return character
		}
	}

	panic("Unexpected code path")
	return database.Character{}
}

func quit(conn net.Conn) {
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

func userMenu(user database.User) utils.Menu {
	chars := engine.M.GetUserCharacters(user.Id)

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

func deleteMenu(user database.User) utils.Menu {
	chars := engine.M.GetUserCharacters(user.Id)

	menu := utils.NewMenu("Delete character")

	menu.AddAction("c", "[C]ancel")

	for i, char := range chars {
		indexStr := strconv.Itoa(i + 1)
		actionText := fmt.Sprintf("[%v]%v", indexStr, char.PrettyName())
		menu.AddActionData(indexStr, actionText, char.Id)
	}

	return menu
}

func adminMenu() utils.Menu {
	menu := utils.NewMenu("Admin")
	menu.AddAction("u", "[U]sers")
	return menu
}

func userAdminMenu() utils.Menu {
	menu := utils.NewMenu("User Admin")

	for i, user := range engine.M.GetUsers() {
		indexStr := strconv.Itoa(i + 1)

		online := ""
		if user.Online() {
			online = "*"
		}

		actionText := fmt.Sprintf("[%v]%v", indexStr, user.PrettyName()+online)
		menu.AddActionData(indexStr, actionText, user.Id)
	}

	return menu
}

func userSpecificMenu(user database.User) utils.Menu {
	menu := utils.NewMenu("User: " + user.PrettyName())
	menu.AddAction("d", "[D]elete")
	return menu
}

func handleConnection(session *mgo.Session, conn net.Conn) {
	defer conn.Close()
	defer session.Close()

	var user database.User
	var character database.Character

	defer func() {
		if r := recover(); r != nil {
			engine.Logout(user)
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
			choice, _ := menu.Exec(conn, utils.ColorModeNone)

			switch choice {
			case "l":
				user = login(conn)
			case "n":
				user = newUser(conn)
			case "":
				fallthrough
			case "q":
				quit(conn)
				return
			}
		} else if character.Name == "" {
			menu := userMenu(user)
			choice, charId := menu.Exec(conn, user.ColorMode)

			switch choice {
			case "":
				fallthrough
			case "l":
				engine.Logout(user)
				user = database.User{}
			case "a":
				adminMenu := adminMenu()
				for {
					choice, _ := adminMenu.Exec(conn, user.ColorMode)
					if choice == "" {
						break
					} else if choice == "u" {
						for {
							userAdminMenu := userAdminMenu()
							choice, userId := userAdminMenu.Exec(conn, user.ColorMode)
							if choice == "" {
								break
							} else {
								_, err := strconv.Atoi(choice)

								if err == nil {
									for {
										userMenu := userSpecificMenu(engine.M.GetUser(userId))
										choice, _ = userMenu.Exec(conn, user.ColorMode)
										if choice == "" {
											break
										} else if choice == "d" {
											engine.M.DeleteUser(userId)
											break
										}
									}
								}
							}
						}
					}
				}
			case "n":
				character = newCharacter(conn, &user)
			case "d":
				for {
					deleteMenu := deleteMenu(user)
					deleteChoice, deleteCharId := deleteMenu.Exec(conn, user.ColorMode)

					if deleteChoice == "" || deleteChoice == "c" {
						break
					}

					_, err := strconv.Atoi(deleteChoice)

					if err == nil {
						engine.M.DeleteCharacter(deleteCharId)
					}
				}

			default:
				_, err := strconv.Atoi(choice)

				if err == nil {
					character = engine.M.GetCharacter(charId)
				}
			}
		} else {
			game.Exec(conn, &user, &character)
			character = database.Character{}
		}
	}
}

func Exec() {
	fmt.Printf("Connecting to database... ")
	session, err := mgo.Dial("localhost")

	utils.HandleError(err)

	fmt.Println("done.")

	listener, err := net.Listen("tcp", ":8945")
	utils.HandleError(err)

	err = engine.Init(session.Copy())

	// If there are no rooms at all create one
	rooms := engine.M.GetRooms()
	if len(rooms) == 0 {
		room := database.NewRoom("")
		engine.M.UpdateRoom(room)
	}

	fmt.Println("Server listening on port 8945")

	for {
		conn, err := listener.Accept()
		utils.HandleError(err)
		fmt.Println("Client connected:", conn.RemoteAddr())
		go handleConnection(session.Copy(), conn)
	}
}

// vim: nocindent
