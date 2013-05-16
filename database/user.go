package database

import (
	"crypto/sha1"
	"io"
	"kmud/utils"
	"net"
	"reflect"
)

type User struct {
	DbObject `bson:",inline"`

	ColorMode utils.ColorMode
	Password  []byte

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
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if cm != self.ColorMode {
		self.ColorMode = cm
		modified(self)
	}
}

func (self *User) GetColorMode() utils.ColorMode {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.ColorMode
}

func hash(data string) []byte {
	h := sha1.New()
	io.WriteString(h, data)
	return h.Sum(nil)
}

// SetPassword SHA1 hashes the password before saving it to the database
func (self *User) SetPassword(password string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	hashed := hash(password)

	if !reflect.DeepEqual(hashed, self.Password) {
		self.Password = hashed
		modified(self)
	}
}

func (self *User) VerifyPassword(password string) bool {
	hashed := hash(password)
	return reflect.DeepEqual(hashed, self.GetPassword())
}

// GetPassword returns the SHA1 of the user's password
func (self *User) GetPassword() []byte {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.Password
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

func UserNames(users []*User) []string {
	names := make([]string, len(users))

	for i, user := range users {
		names[i] = user.PrettyName()
	}

	return names
}

type Users []*User

func (s Users) Len() int {
	return len(s)
}

func (s Users) Less(i, j int) bool {
	return utils.NaturalLessThan(s[i].PrettyName(), s[j].PrettyName())
}

func (s Users) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// vim: nocindent
