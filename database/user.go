package database

import (
	"kmud/utils"
	"labix.org/v2/mgo/bson"
)

const (
	userColorMode string = "colormode"
)

type User struct {
	DbObject  `bson:",inline"`
	ColorMode utils.ColorMode
	online    bool
}

func NewUser(name string) User {
	var user User
	user.Id = bson.NewObjectId()
	user.Name = name
	user.ColorMode = utils.ColorModeNone
	return user
}

func (self *User) SetOnline(online bool) {
	self.online = online
}

func (self *User) Online() bool {
	return self.online
}

// vim: nocindent
