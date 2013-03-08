package database

import (
	"crypto/sha1"
	"io"
	"kmud/utils"
	"net"
)

const (
	userColorMode string = "colormode"
	userPassword  string = "password"
)

type User struct {
	DbObject     `bson:",inline"`
	colorMode    utils.ColorMode
	online       bool
	conn         net.Conn
	windowWidth  int
	windowHeight int
	terminalType string
}

func NewUser(name string, password string) *User {
	var user User
	user.initDbObject(name, userType)

	user.SetPassword(password)
	user.SetColorMode(utils.ColorModeNone)
	user.SetOnline(false)

	user.windowWidth = 80
	user.windowHeight = 40

	return &user
}

func (self *User) SetOnline(online bool) {
	self.online = online

	if !online {
		self.conn = nil
	}
}

func (self *User) Online() bool {
	return self.online
}

func (self *User) SetColorMode(cm utils.ColorMode) {
	self.setField(userColorMode, cm)
}

func (self *User) GetColorMode() utils.ColorMode {
	return self.getField(userColorMode).(utils.ColorMode)
}

func hash(data string) []byte {
	h := sha1.New()
	io.WriteString(h, data)
	return h.Sum(nil)
}

// SetPassword SHA1 hashes the password before saving it to the database
func (self *User) SetPassword(password string) {
	self.setField(userPassword, hash(userPassword))
}

func (self *User) VerifyPassword(password string) bool {
	return true // TODO

	hashed := hash(password)

	for i, b := range self.GetPassword() {
		if b != hashed[i] {
			return false
		}
	}

	return true
}

// GetPassword returns the SHA1 of the user's password
func (self *User) GetPassword() []byte {
	if self.hasField(userPassword) {
		return self.getField(userPassword).([]byte)
	}

	return []byte{}
}

func (self *User) SetConnection(conn net.Conn) {
	self.conn = conn
}

func (self *User) GetConnection() net.Conn {
	return self.conn
}

func (self *User) SetWindowSize(width int, height int) {
	self.windowWidth = width
	self.windowHeight = height
}

func (self *User) WindowSize() (width int, height int) {
	return self.windowWidth, self.windowHeight
}

func (self *User) SetTerminalType(tt string) {
	self.terminalType = tt
}

func (self *User) TerminalType() string {
	return self.terminalType
}

// vim: nocindent
