package server

import (
	"fmt"
	"kmud/database"
	"kmud/game"
	"kmud/model"
	"kmud/utils"
	"labix.org/v2/mgo"
	"net"
	"strconv"
	"time"
)

func login(conn net.Conn) *database.User {
	for {
		username := utils.GetUserInput(conn, "Username: ")

		if username == "" {
			return nil
		}

		user := model.M.GetUserByName(username)

		if user == nil {
			utils.WriteLine(conn, "User not found")
		} else if user.Online() {
			utils.WriteLine(conn, "That user is already online")
		} else {
			attempts := 1
			for {
				password := utils.GetPassword(conn, "Password: ")
				if user.VerifyPassword(password) {
					break
				}

				if attempts >= 3 {
					utils.WriteLine(conn, "Too many failed login attempts")
					conn.Close()
					panic("Booted user due to too many failed logins (" + user.PrettyName() + ")")
				}

				attempts++

				utils.WriteLine(conn, "Invalid password")
				time.Sleep(1 * time.Second)
			}

			user.SetOnline(true)
			return user
		}
	}

	panic("Unexpected code path")
	return nil
}

func newUser(conn net.Conn) *database.User {
	for {
		name := utils.GetUserInput(conn, "Desired username: ")

		if name == "" {
			return nil
		}

		user := model.M.GetUserByName(name)
		password := ""

		if user != nil {
			utils.WriteLine(conn, "That name is unavailable")
		} else if err := utils.ValidateName(name); err != nil {
			utils.WriteLine(conn, err.Error())
		} else {
			for {
				pass1 := utils.GetPassword(conn, "Desired password: ")

				if len(pass1) < 7 {
					utils.WriteLine(conn, "Passwords must be at least 7 letters in length")
					continue
				}

				pass2 := utils.GetPassword(conn, "Reenter password: ")

				if pass1 != pass2 {
					utils.WriteLine(conn, "Passwords do not match")
					continue
				}

				password = pass1

				break
			}

			user = model.M.CreateUser(name, password)
			user.SetOnline(true)
			return user
		}
	}

	panic("Unexpected code path")
	return nil
}

func newPlayer(conn net.Conn, user *database.User) *database.Character {
	// TODO: character slot limit
	const SizeLimit = 12
	for {
		name := utils.GetUserInput(conn, "Desired character name: ")

		if name == "" {
			return nil
		}

		char := model.M.GetCharacterByName(name)

		if char != nil {
			utils.WriteLine(conn, "That name is unavailable")
		} else if err := utils.ValidateName(name); err != nil {
			utils.WriteLine(conn, err.Error())
		} else {
			room := model.M.GetRooms()[0] // TODO: Better way to pick an initial character location
			return model.M.CreateCharacter(name, user, room)
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

func userMenu(user *database.User) utils.Menu {
	chars := model.M.GetUserCharacters(user)

	menu := utils.NewMenu(user.PrettyName())
	menu.AddAction("l", "[L]ogout")
	menu.AddAction("a", "[A]dmin")
	menu.AddAction("n", "[N]ew character")
	if len(chars) > 0 {
		menu.AddAction("d", "[D]elete character")
	}

	// TODO: Sort character list

	for i, char := range chars {
		index := i + 1
		actionText := fmt.Sprintf("[%v]%v", index, char.PrettyName())
		menu.AddActionData(index, actionText, char.GetId())
	}

	return menu
}

func deleteMenu(user *database.User) utils.Menu {
	chars := model.M.GetUserCharacters(user)

	menu := utils.NewMenu("Delete character")

	menu.AddAction("c", "[C]ancel")

	// TODO: Sort character list

	for i, char := range chars {
		index := i + 1
		actionText := fmt.Sprintf("[%v]%v", index, char.PrettyName())
		menu.AddActionData(index, actionText, char.GetId())
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
		menu.AddActionData(index, actionText, user.GetId())
	}

	return menu
}

func userSpecificMenu(user *database.User) utils.Menu {
	menu := utils.NewMenu("User: " + user.PrettyName())
	menu.AddAction("d", "[D]elete")
	return menu
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var user *database.User
	var player *database.Character

	defer func() {
		if r := recover(); r != nil {
			username := ""
			charname := ""

			if user != nil {
				user.SetOnline(false)
				username = user.PrettyName()
			}

			if player != nil {
				charname = player.PrettyName()
			}

			fmt.Printf("Lost connection to client (%v/%v): %v, %v\n",
				username,
				charname,
				conn.RemoteAddr(),
				r)
		}
	}()

	for {
		if user == nil {
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
		} else if player == nil {
			menu := userMenu(user)
			choice, charId := menu.Exec(conn, user.GetColorMode())

			switch choice {
			case "":
				fallthrough
			case "l":
				user.SetOnline(false)
				user = nil
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
				player = newPlayer(conn, user)
			case "d":
				for {
					deleteMenu := deleteMenu(user)
					deleteChoice, deleteCharId := deleteMenu.Exec(conn, user.GetColorMode())

					if deleteChoice == "" || deleteChoice == "c" {
						break
					}

					_, err := strconv.Atoi(deleteChoice)

					if err == nil {
						// TODO: Delete confirmation
						model.M.DeleteCharacter(deleteCharId)
					}
				}

			default:
				_, err := strconv.Atoi(choice)

				if err == nil {
					player = model.M.GetCharacter(charId)
				}
			}
		} else {
			game.Exec(conn, user, player)
			player = nil
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

		var zone *database.Zone

		if len(zones) == 0 {
			zone = model.M.CreateZone("Default")
		} else {
			zone = zones[0]
		}

		model.M.CreateRoom(zone)
	}

	fmt.Println("Server listening on port 8945")

	for {
		conn, err := listener.Accept()
		utils.HandleError(err)
		fmt.Println("Client connected:", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

// vim: nocindent
