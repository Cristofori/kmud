package database

import (
	"crypto/sha1"
	"fmt"
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

func NewUser(name string, password string, admin bool) *User {
	user := &User{
		Name:         utils.FormatName(name),
		Password:     hash(password),
		ColorMode:    types.ColorModeNone,
		Admin:        admin,
		online:       false,
		windowWidth:  80,
		windowHeight: 40,
	}

	dbinit(user)
	return user
}

func (self *User) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *User) SetName(name string) {
	self.writeLock(func() {
		self.Name = utils.FormatName(name)
	})
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
	self.writeLock(func() {
		self.ColorMode = cm
	})
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
	self.writeLock(func() {
		self.Password = hash(password)
	})
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

const MinWidth = 60
const MinHeight = 20

func (self *User) GetWindowSize() (width int, height int) {
	return utils.Max(self.windowWidth, MinWidth),
		utils.Max(self.windowHeight, MinHeight)
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

func (self *User) WriteLine(line string, a ...interface{}) {
	utils.WriteLine(self.conn, fmt.Sprintf(line, a...), self.GetColorMode())
}

func (self *User) Write(text string) {
	utils.Write(self.conn, text, self.GetColorMode())
}

func (self *User) SetAdmin(admin bool) {
	self.writeLock(func() {
		self.Admin = admin
	})
}

func (self *User) IsAdmin() bool {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Admin
}
