package server

import (
	"fmt"
	"kmud/database"
	"kmud/engine"
	"kmud/model"
	"kmud/session"
	"kmud/telnet"
	"kmud/utils"
	"labix.org/v2/mgo"
	"net"
	"sort"
	"strconv"
	"time"
)

type Server struct {
	listener net.Listener
}

type wrappedConnection struct {
	telnet  *telnet.Telnet
	watcher *utils.WatchableReadWriter
}

func (s *wrappedConnection) Write(p []byte) (int, error) {
	return s.watcher.Write(p)
}

func (s *wrappedConnection) Read(p []byte) (int, error) {
	return s.watcher.Read(p)
}

func (s *wrappedConnection) Close() error {
	return s.telnet.Close()
}

func (s *wrappedConnection) LocalAddr() net.Addr {
	return s.telnet.LocalAddr()
}

func (s *wrappedConnection) RemoteAddr() net.Addr {
	return s.telnet.RemoteAddr()
}

func (s *wrappedConnection) SetDeadline(dl time.Time) error {
	return s.telnet.SetDeadline(dl)
}

func (s *wrappedConnection) SetReadDeadline(dl time.Time) error {
	return s.telnet.SetReadDeadline(dl)
}

func (s *wrappedConnection) SetWriteDeadline(dl time.Time) error {
	return s.telnet.SetWriteDeadline(dl)
}

func login(conn *wrappedConnection) *database.User {
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
			conn.telnet.WillEcho()
			for {
				password := utils.GetRawUserInputSuffix(conn, "Password: ", "\r\n")

				// TODO - Disabling password verification to make development easier
				if user.VerifyPassword(password) || true {
					break
				}

				if attempts >= 3 {
					utils.WriteLine(conn, "Too many failed login attempts")
					conn.Close()
					panic("Booted user due to too many failed logins (" + user.GetName() + ")")
				}

				attempts++

				time.Sleep(2 * time.Second)
				utils.WriteLine(conn, "Invalid password")
			}
			conn.telnet.WontEcho()

			return user
		}
	}
}

func newUser(conn *wrappedConnection) *database.User {
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
			conn.telnet.WillEcho()
			for {
				pass1 := utils.GetRawUserInputSuffix(conn, "Desired password: ", "\r\n")

				if len(pass1) < 7 {
					utils.WriteLine(conn, "Passwords must be at least 7 letters in length")
					continue
				}

				pass2 := utils.GetRawUserInputSuffix(conn, "Confirm password: ", "\r\n")

				if pass1 != pass2 {
					utils.WriteLine(conn, "Passwords do not match")
					continue
				}

				password = pass1

				break
			}
			conn.telnet.WontEcho()

			user = model.M.CreateUser(name, password)
			return user
		}
	}
}

func newPlayer(conn *wrappedConnection, user *database.User) *database.Character {
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
			return model.M.CreatePlayer(name, user, room)
		}
	}
}

func quit(conn *wrappedConnection) {
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

	menu := utils.NewMenu(user.GetName())
	menu.AddAction("l", "[L]ogout")
	menu.AddAction("a", "[A]dmin")
	menu.AddAction("n", "[N]ew character")
	if len(chars) > 0 {
		menu.AddAction("d", "[D]elete character")
	}

	// TODO: Sort character list

	for i, char := range chars {
		index := i + 1
		actionText := fmt.Sprintf("[%v]%v", index, char.GetName())
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
		actionText := fmt.Sprintf("[%v]%v", index, char.GetName())
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

	users := model.M.GetUsers()
	sort.Sort(users)

	for i, user := range users {
		index := i + 1

		online := ""
		if user.Online() {
			online = "*"
		}

		actionText := fmt.Sprintf("[%v]%v", index, user.GetName()+online)
		menu.AddActionData(index, actionText, user.GetId())
	}

	return menu
}

func userSpecificMenu(user *database.User) utils.Menu {
	suffix := ""
	if user.Online() {
		suffix = "(Online)"
	} else {
		suffix = "(Offline)"
	}

	menu := utils.NewMenu("User: " + user.GetName() + " " + suffix)
	menu.AddAction("d", "[D]elete")

	if user.Online() {
		menu.AddAction("w", "[W]atch")
	}

	return menu
}

func handleConnection(conn *wrappedConnection) {
	defer conn.Close()

	var user *database.User
	var player *database.Character

	defer func() {
		if r := recover(); r != nil {
			username := ""
			charname := ""

			if user != nil {
				user.SetOnline(false)
				username = user.GetName()
			}

			if player != nil {
				charname = player.GetName()
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

			if user == nil {
				continue
			}

			user.SetOnline(true)
			user.SetConnection(conn)

			conn.telnet.DoWindowSize()
			conn.telnet.DoTerminalType()

			conn.telnet.Listen(func(code telnet.TelnetCode, data []byte) {
				switch code {
				case telnet.WS:
					if len(data) != 4 {
						fmt.Println("Malformed window size data:", data)
						return
					}

					width := (255 * data[0]) + data[1]
					height := (255 * data[2]) + data[3]
					user.SetWindowSize(int(width), int(height))

				case telnet.TT:
					user.SetTerminalType(string(data))
				}
			})

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
											model.M.DeleteUserId(userId)
											break
										} else if choice == "w" {
											userToWatch := model.M.GetUser(userId)

											if userToWatch == user {
												utils.WriteLine(conn, "You can't watch yourself!")
											} else {
												userConn := userToWatch.GetConnection().(*wrappedConnection)

												userConn.watcher.AddWatcher(conn)
												utils.GetRawUserInput(conn, "Type anything to stop watching\r\n")
												userConn.watcher.RemoveWatcher(conn)
											}
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
						model.M.DeleteCharacterId(deleteCharId)
					}
				}

			default:
				_, err := strconv.Atoi(choice)

				if err == nil {
					player = model.M.GetCharacter(charId)
				}
			}
		} else {
			session := session.NewSession(conn, user, player)
			session.Exec()
			player = nil
		}
	}
}

func (self *Server) Start() {
	fmt.Printf("Connecting to database... ")
	session, err := mgo.Dial("localhost")

	utils.HandleError(err)

	fmt.Println("done.")

	self.listener, err = net.Listen("tcp", ":8945")
	utils.HandleError(err)

	err = model.Init(database.NewMongoSession(session.Copy()))

	// If there are no rooms at all create one
	rooms := model.M.GetRooms()
	if len(rooms) == 0 {
		zones := model.M.GetZones()

		var zone *database.Zone

		if len(zones) == 0 {
			zone, _ = model.M.CreateZone("Default")
		} else {
			zone = zones[0]
		}

		model.M.CreateRoom(zone, database.Coordinate{X: 0, Y: 0, Z: 0})
	}

	fmt.Println("Server listening on port 8945")
}

func (self *Server) Listen() {
	for {
		conn, err := self.listener.Accept()
		utils.HandleError(err)
		fmt.Println("Client connected:", conn.RemoteAddr())
		t := telnet.NewTelnet(conn)

		wc := utils.NewWatchableReadWriter(t)

		go handleConnection(&wrappedConnection{t, wc})
	}
}

func (self *Server) Exec() {
    database.GetTime()
	self.Start()
	engine.Start()
	self.Listen()
}

// vim: nocindent
