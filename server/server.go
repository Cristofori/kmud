package server

import (
	"fmt"
	"kmud/database"
	"kmud/game"
	"kmud/model"
	"kmud/utils"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net"
	"strconv"
)

func login(conn net.Conn) database.User {
	for {
		line := utils.GetUserInput(conn, "Username: ")

		if line == "" {
			return database.User{}
		}

		user, found := model.M.GetUserByName(line)

		if !found {
			utils.WriteLine(conn, "User not found")
		} else {
			err := model.Login(user)

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

		user, found := model.M.GetUserByName(name)

		if found {
			utils.WriteLine(conn, "That name is unavailable")
		} else if err := utils.ValidateName(name); err != nil {
			utils.WriteLine(conn, err.Error())
		} else {
			user = database.NewUser(name)
			model.M.UpdateUser(user)

			err := model.Login(user)

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

func newCharacter(conn net.Conn, user *database.User) *database.Character {
	// TODO: character slot limit
	const SizeLimit = 12
	for {
		name := utils.GetUserInput(conn, "Desired character name: ")

		if name == "" {
			return nil
		}

		_, found := model.M.GetCharacterByName(name)

		if found {
			utils.WriteLine(conn, "That name is unavailable")
		} else if err := utils.ValidateName(name); err != nil {
			utils.WriteLine(conn, err.Error())
		} else {
			room := model.M.GetRooms()[0] // TODO: Better way to pick an initial character location
			return model.M.CreateCharacter(name, user.Id, room.Id)
		}
	}

	panic("Unexpected code path")
	return nil
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
	chars := model.M.GetUserCharacters(user.Id)

	menu := utils.NewMenu(user.PrettyName())
	menu.AddAction("l", "[L]ogout")
	menu.AddAction("a", "[A]dmin")
	menu.AddAction("n", "[N]ew character")
	if len(chars) > 0 {
		menu.AddAction("d", "[D]elete character")
	}

	for i, char := range chars {
		index := i + 1
		actionText := fmt.Sprintf("[%v]%v", index, char.PrettyName())
		menu.AddActionData(index, actionText, char.Id)
	}

	return menu
}

func deleteMenu(user database.User) utils.Menu {
	chars := model.M.GetUserCharacters(user.Id)

	menu := utils.NewMenu("Delete character")

	menu.AddAction("c", "[C]ancel")

	for i, char := range chars {
		index := i + 1
		actionText := fmt.Sprintf("[%v]%v", index, char.PrettyName())
		menu.AddActionData(index, actionText, char.Id)
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

	for i, user := range model.M.GetUsers() {
		index := i + 1

		online := ""
		if user.Online() {
			online = "*"
		}

		actionText := fmt.Sprintf("[%v]%v", index, user.PrettyName()+online)
		menu.AddActionData(index, actionText, user.Id)
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
	var character *database.Character

	defer func() {
		if r := recover(); r != nil {
			model.Logout(user)
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
		} else if character == nil {
			menu := userMenu(user)
			choice, charId := menu.Exec(conn, user.GetColorMode())

			switch choice {
			case "":
				fallthrough
			case "l":
				model.Logout(user)
				user = database.User{}
			case "a":
				adminMenu := adminMenu()
				for {
					choice, _ := adminMenu.Exec(conn, user.GetColorMode())
					if choice == "" {
						break
					} else if choice == "u" {
						for {
							userAdminMenu := userAdminMenu()
							choice, userId := userAdminMenu.Exec(conn, user.GetColorMode())
							if choice == "" {
								break
							} else {
								_, err := strconv.Atoi(choice)

								if err == nil {
									for {
										userMenu := userSpecificMenu(model.M.GetUser(userId))
										choice, _ = userMenu.Exec(conn, user.GetColorMode())
										if choice == "" {
											break
										} else if choice == "d" {
											model.M.DeleteUser(userId)
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
					deleteChoice, deleteCharId := deleteMenu.Exec(conn, user.GetColorMode())

					if deleteChoice == "" || deleteChoice == "c" {
						break
					}

					_, err := strconv.Atoi(deleteChoice)

					if err == nil {
						model.M.DeleteCharacter(deleteCharId)
					}
				}

			default:
				_, err := strconv.Atoi(choice)

				if err == nil {
					character = model.M.GetCharacter(charId)
				}
			}
		} else {
			game.Exec(conn, &user, character)
			character = nil
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

	err = model.Init(session.Copy())

	// If there are no rooms at all create one
	rooms := model.M.GetRooms()
	if len(rooms) == 0 {
		zones := model.M.GetZones()

		var zoneId bson.ObjectId

		if len(zones) == 0 {
			newZone := model.M.CreateZone("Default")
			zoneId = newZone.Id
		} else {
			zoneId = zones[0].Id
		}

		model.M.CreateRoom(zoneId)
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
