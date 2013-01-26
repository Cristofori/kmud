package database

import (
	"labix.org/v2/mgo/bson"
)

const (
	characterRoomId string = "roomid"
	characterUserId string = "userid"
	characterCash   string = "cash"
)

type Character struct {
	DbObject  `bson:",inline"`
	Inventory []bson.ObjectId
	online    bool
}

func NewCharacter(name string, userId bson.ObjectId, roomId bson.ObjectId) Character {
	var character Character
	character.initDbObject()

	character.Id = bson.NewObjectId()
	character.SetUser(userId)
	character.Name = name
	character.SetRoom(roomId)
	character.SetCash(0)
	character.online = false
	return character
}

func NewNpc(name string, roomId bson.ObjectId) Character {
	return NewCharacter(name, "", roomId)
}

func (self *Character) SetOnline(online bool) {
	self.online = online
}

func (self *Character) IsOnline() bool {
	return self.online
}

func (self *Character) IsNpc() bool {
	return self.GetUserId() == ""
}

func (self *Character) SetName(name string) {
	self.Name = name
}

func (self *Character) GetName() string {
	return self.Name
}

func (self *Character) GetRoomId() bson.ObjectId {
	return self.Fields[characterRoomId].(bson.ObjectId)
}

func (self *Character) SetRoom(id bson.ObjectId) {
	self.setField(characterRoomId, id)
}

func (self *Character) SetUser(id bson.ObjectId) {
	self.setField(characterUserId, id)
}

func (self *Character) GetUserId() bson.ObjectId {
	return self.Fields[characterUserId].(bson.ObjectId)
}

func (self *Character) SetCash(amount int) {
	self.setField(characterCash, amount)
}

func (self *Character) AddCash(amount int) {
	self.setField(characterCash, self.GetCash()+amount)
}

func (self *Character) GetCash() int {
	return self.Fields[characterCash].(int)
}
