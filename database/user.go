package database

import (
	"crypto/sha1"
	"io"
	"net"
	"reflect"

	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type User struct {
	DbObject `bson:",inline"`

	Name      string
	ColorMode types.ColorMode
	Password  []byte
	Admin     bool

	online       bool
	conn         net.Conn
	windowWidth  int
	windowHeight int
	terminalType string
}

func NewUser(name string, password string) *User {
	var user User

	user.Name = utils.FormatName(name)
	user.Password = hash(password)
	user.ColorMode = types.ColorModeNone
	user.online = false

	user.windowWidth = 80
	user.windowHeight = 40

	user.initDbObject(&user)
	return &user
}

func (self *User) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *User) SetName(name string) {
	if name != self.GetName() {
		self.WriteLock()
		self.Name = utils.FormatName(name)
		self.WriteUnlock()
		self.modified()
	}
}

func (self *User) SetOnline(online bool) {
	self.online = online

	if !online {
		self.conn = nil
	}
}

func (self *User) IsOnline() bool {
	return self.online
}

func (self *User) SetColorMode(cm types.ColorMode) {
	if cm != self.GetColorMode() {
		self.WriteLock()
		self.ColorMode = cm
		self.WriteUnlock()

		self.modified()
	}
}

func (self *User) GetColorMode() types.ColorMode {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.ColorMode
}

func hash(data string) []byte {
	h := sha1.New()
	io.WriteString(h, data)
	return h.Sum(nil)
}

// SetPassword SHA1 hashes the password before saving it to the database
func (self *User) SetPassword(password string) {
	hashed := hash(password)

	if !reflect.DeepEqual(hashed, self.GetPassword()) {
		self.WriteLock()
		self.Password = hashed
		self.WriteUnlock()

		self.modified()
	}
}

func (self *User) VerifyPassword(password string) bool {
	hashed := hash(password)
	return reflect.DeepEqual(hashed, self.GetPassword())
}

// GetPassword returns the SHA1 of the user's password
func (self *User) GetPassword() []byte {
	self.ReadLock()
	defer self.ReadUnlock()

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

func (self *User) GetWindowSize() (width int, height int) {
	return self.windowWidth, self.windowHeight
}

func (self *User) SetTerminalType(tt string) {
	self.terminalType = tt
}

func (self *User) GetTerminalType() string {
	return self.terminalType
}

func (self *User) GetInput(prompt string) string {
	return utils.GetUserInput(self.conn, prompt, self.GetColorMode())
}

func (self *User) WriteLine(line string) (int, error) {
	return utils.WriteLine(self.conn, line, self.GetColorMode())
}

func (self *User) Write(text string) (int, error) {
	return utils.Write(self.conn, text, self.GetColorMode())
}

func (self *User) SetAdmin(admin bool) {
	self.WriteLock()
	defer self.WriteUnlock()

	self.Admin = admin
	self.modified()
}

func (self *User) IsAdmin() bool {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Admin
}
