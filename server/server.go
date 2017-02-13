package server

import (
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"sort"
	"strconv"
	"time"

	"github.com/Cristofori/kmud/database"
	"github.com/Cristofori/kmud/engine"
	"github.com/Cristofori/kmud/model"
	"github.com/Cristofori/kmud/session"
	"github.com/Cristofori/kmud/telnet"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
	"gopkg.in/mgo.v2"
)

type Server struct {
	listener net.Listener
}

type connectionHandler struct {
	user types.User
	pc   types.PC
	conn *wrappedConnection
}

type wrappedConnection struct {
	telnet.Telnet
	watcher *utils.WatchableReadWriter
}

func (s *wrappedConnection) Write(p []byte) (int, error) {
	return s.watcher.Write(p)
}

func (s *wrappedConnection) Read(p []byte) (int, error) {
	return s.watcher.Read(p)
}

func login(conn *wrappedConnection) types.User {
	for {
		username := utils.GetUserInput(conn, "Username: ", types.ColorModeNone)

		if username == "" {
			return nil
		}

		user := model.GetUserByName(username)

		if user == nil {
			utils.WriteLine(conn, "User not found", types.ColorModeNone)
		} else if user.IsOnline() {
			utils.WriteLine(conn, "That user is already online", types.ColorModeNone)
		} else {
			attempts := 1
			conn.WillEcho()
			for {
				password := utils.GetRawUserInputSuffix(conn, "Password: ", "\r\n", types.ColorModeNone)

				// TODO - Disabling password verification to make development easier
				if user.VerifyPassword(password) || true {
					break
				}

				if attempts >= 3 {
					utils.WriteLine(conn, "Too many failed login attempts", types.ColorModeNone)
					conn.Close()
					panic("Booted user due to too many failed logins (" + user.GetName() + ")")
				}

				attempts++

				time.Sleep(2 * time.Second)
				utils.WriteLine(conn, "Invalid password", types.ColorModeNone)
			}
			conn.WontEcho()

			return user
		}
	}
}

func newUser(conn *wrappedConnection) types.User {
	for {
		name := utils.GetUserInput(conn, "Desired username: ", types.ColorModeNone)

		if name == "" {
			return nil
		}

		user := model.GetUserByName(name)
		password := ""

		if user != nil {
			utils.WriteLine(conn, "That name is unavailable", types.ColorModeNone)
		} else if err := utils.ValidateName(name); err != nil {
			utils.WriteLine(conn, err.Error(), types.ColorModeNone)
		} else {
			conn.WillEcho()
			for {
				pass1 := utils.GetRawUserInputSuffix(conn, "Desired password: ", "\r\n", types.ColorModeNone)

				if len(pass1) < 7 {
					utils.WriteLine(conn, "Passwords must be at least 7 letters in length", types.ColorModeNone)
					continue
				}

				pass2 := utils.GetRawUserInputSuffix(conn, "Confirm password: ", "\r\n", types.ColorModeNone)

				if pass1 != pass2 {
					utils.WriteLine(conn, "Passwords do not match", types.ColorModeNone)
					continue
				}

				password = pass1

				break
			}
			conn.WontEcho()

			admin := model.UserCount() == 0
			user = model.CreateUser(name, password, admin)
			return user
		}
	}
}

func (self *connectionHandler) newPlayer() types.PC {
	// TODO: character slot limit
	const SizeLimit = 12
	for {
		name := self.user.GetInput("Desired character name: ")

		if name == "" {
			return nil
		}

		char := model.GetCharacterByName(name)

		if char != nil {
			self.user.WriteLine("That name is unavailable")
		} else if err := utils.ValidateName(name); err != nil {
			self.user.WriteLine(err.Error())
		} else {
			room := model.GetRooms()[0] // TODO: Better way to pick an initial character location
			return model.CreatePlayerCharacter(name, self.user.GetId(), room)
		}
	}
}

func (self *connectionHandler) WriteLine(line string, a ...interface{}) {
	utils.WriteLine(self.conn, fmt.Sprintf(line, a...), types.ColorModeNone)
}

func (self *connectionHandler) Write(text string) {
	utils.Write(self.conn, text, types.ColorModeNone)
}

func (self *connectionHandler) GetInput(prompt string) string {
	return utils.GetUserInput(self.conn, prompt, types.ColorModeNone)
}

func (sefl *connectionHandler) GetWindowSize() (int, int) {
	return 80, 80
}

func (self *connectionHandler) mainMenu() {
	utils.ExecMenu(
		"MUD",
		self,
		func(menu *utils.Menu) {
			menu.AddAction("l", "Login", func() {
				self.user = login(self.conn)
				self.loggedIn()
			})

			menu.AddAction("n", "New user", func() {
				self.user = newUser(self.conn)
				self.loggedIn()
			})

			menu.OnExit(func() {
				utils.WriteLine(self.conn, "Take luck!", types.ColorModeNone)
				self.conn.Close()
			})
		})
}

func (self *connectionHandler) userMenu() {
	utils.ExecMenu(
		self.user.GetName(),
		self.user,
		func(menu *utils.Menu) {
			menu.OnExit(func() {
				self.user.SetOnline(false)
				self.user = nil
			})

			if self.user.IsAdmin() {
				menu.AddAction("a", "Admin", func() {
					self.adminMenu()
				})
			}

			menu.AddAction("n", "New character", func() {
				self.pc = self.newPlayer()
			})

			// TODO: Sort character list
			chars := model.GetUserCharacters(self.user.GetId())

			if len(chars) > 0 {
				menu.AddAction("d", "Delete character", func() {
					self.deleteMenu()
				})
			}

			for i, char := range chars {
				c := char
				menu.AddAction(strconv.Itoa(i+1), char.GetName(), func() {
					self.pc = c
					self.launchSession()
				})
			}
		})
}

func (self *connectionHandler) deleteMenu() {
	utils.ExecMenu(
		"Delete character",
		self.user,
		func(menu *utils.Menu) {
			// TODO: Sort character list
			chars := model.GetUserCharacters(self.user.GetId())
			for i, char := range chars {
				c := char
				menu.AddAction(strconv.Itoa(i+1), char.GetName(), func() {
					// TODO: Delete confirmation
					model.DeleteCharacter(c.GetId())
				})
			}
		})
}

func (self *connectionHandler) adminMenu() {
	utils.ExecMenu(
		"Admin",
		self.user,
		func(menu *utils.Menu) {
			menu.AddAction("u", "Users", func() {
				self.userAdminMenu()
			})
		})
}

func (self *connectionHandler) userAdminMenu() {
	utils.ExecMenu("User Admin", self.user, func(menu *utils.Menu) {
		users := model.GetUsers()
		sort.Sort(users)

		for i, user := range users {
			online := ""
			if user.IsOnline() {
				online = "*"
			}

			u := user
			menu.AddAction(strconv.Itoa(i+1), user.GetName()+online, func() {
				self.specificUserMenu(u)
			})
		}
	})
}

func (self *connectionHandler) specificUserMenu(user types.User) {
	suffix := ""
	if user.IsOnline() {
		suffix = "(Online)"
	} else {
		suffix = "(Offline)"
	}

	utils.ExecMenu(
		fmt.Sprintf("User: %s %s", user.GetName(), suffix),
		self.user,
		func(menu *utils.Menu) {
			menu.AddAction("d", "Delete", func() {
				model.DeleteUser(user.GetId())
				menu.Exit()
			})

			menu.AddAction("a", fmt.Sprintf("Admin - %v", user.IsAdmin()), func() {
				u := model.GetUser(user.GetId())
				u.SetAdmin(!u.IsAdmin())
			})

			if user.IsOnline() {
				menu.AddAction("w", "Watch", func() {
					if user == self.user {
						self.user.WriteLine("You can't watch yourself!")
					} else {
						userConn := user.GetConnection().(*wrappedConnection)

						userConn.watcher.AddWatcher(self.conn)
						utils.GetRawUserInput(self.conn, "Type anything to stop watching\r\n", self.user.GetColorMode())
						userConn.watcher.RemoveWatcher(self.conn)
					}
				})
			}
		})
}

func (self *connectionHandler) Handle() {
	go func() {
		defer self.conn.Close()

		defer func() {
			r := recover()

			username := ""
			charname := ""

			if self.user != nil {
				self.user.SetOnline(false)
				username = self.user.GetName()
			}

			if self.pc != nil {
				self.pc.SetOnline(false)
				charname = self.pc.GetName()
			}

			if r != io.EOF {
				debug.PrintStack()
			}

			fmt.Printf("Lost connection to client (%v/%v): %v, %v\n",
				username,
				charname,
				self.conn.RemoteAddr(),
				r)
		}()

		self.mainMenu()
	}()
}

func (self *connectionHandler) loggedIn() {
	if self.user == nil {
		return
	}

	self.user.SetOnline(true)
	self.user.SetConnection(self.conn)

	self.conn.DoWindowSize()
	self.conn.DoTerminalType()

	self.conn.Listen(func(code telnet.TelnetCode, data []byte) {
		switch code {
		case telnet.WS:
			if len(data) != 4 {
				fmt.Println("Malformed window size data:", data)
				return
			}

			width := int((255 * data[0])) + int(data[1])
			height := int((255 * data[2])) + int(data[3])
			self.user.SetWindowSize(width, height)

		case telnet.TT:
			self.user.SetTerminalType(string(data))
		}
	})

	self.userMenu()
}

func (self *connectionHandler) launchSession() {
	if self.pc == nil {
		return
	}

	session := session.NewSession(self.conn, self.user, self.pc)
	session.Exec()
	self.pc = nil
}

func (self *Server) Start() {
	fmt.Printf("Connecting to database... ")
	session, err := mgo.Dial("localhost")

	utils.HandleError(err)

	fmt.Println("done.")

	self.listener, err = net.Listen("tcp", ":8945")
	utils.HandleError(err)

	database.Init(database.NewMongoSession(session.Copy()), "mud")
}

func (self *Server) Bootstrap() {
	// Create the world object if necessary
	model.GetWorld()

	// If there are no rooms at all create one
	rooms := model.GetRooms()
	if len(rooms) == 0 {
		zones := model.GetZones()

		var zone types.Zone

		if len(zones) == 0 {
			zone, _ = model.CreateZone("Default")
		} else {
			zone = zones[0]
		}

		model.CreateRoom(zone, types.Coordinate{X: 0, Y: 0, Z: 0})
	}
}

func (self *Server) Listen() {
	for {
		conn, err := self.listener.Accept()
		utils.HandleError(err)
		fmt.Println("Client connected:", conn.RemoteAddr())
		t := telnet.NewTelnet(conn)

		wc := utils.NewWatchableReadWriter(t)

		ch := connectionHandler{
			conn: &wrappedConnection{Telnet: *t, watcher: wc},
		}

		ch.Handle()
	}
}

func (self *Server) Exec() {
	self.Start()
	self.Bootstrap()
	engine.Start()
	self.Listen()
}
