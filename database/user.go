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
	DbObject  `bson:",inline"`
	colorMode utils.ColorMode
	online    bool
	conn      net.Conn
}

func NewUser(name string, password string) *User {
	var user User
	user.initDbObject(name, userType)

	user.SetPassword(password)
	user.SetColorMode(utils.ColorModeNone)
	user.SetOnline(false)

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

// vim: nocindent
