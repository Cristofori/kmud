package database

import (
	"labix.org/v2/mgo/bson"
)

const (
	characterRoomId    string = "roomid"
	characterUserId    string = "userid"
	characterCash      string = "cash"
	characterInventory string = "inventory"
)

type Character struct {
	DbObject `bson:",inline"`
	online   bool
}

func NewCharacter(name string, userId bson.ObjectId, roomId bson.ObjectId) *Character {
	var character Character
	character.initDbObject(name, characterType)

	if userId != "" {
		character.SetUser(userId)
	}
	character.SetRoom(roomId)
	character.SetCash(0)

	character.SetOnline(false)

	return &character
}

func NewNpc(name string, roomId bson.ObjectId) *Character {
	return NewCharacter(name, "", roomId)
}

func (self *Character) SetOnline(online bool) {
	self.online = online
}

func (self *Character) IsOnline() bool {
	return self.online
}

func (self *Character) IsNpc() bool {
	return !self.hasField(characterUserId)
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
	if self.IsNpc() {
		return ""
	}

	return self.getField(characterUserId).(bson.ObjectId)
}

func (self *Character) SetCash(amount int) {
	self.setField(characterCash, amount)
}

func (self *Character) AddCash(amount int) {
	self.setField(characterCash, self.GetCash()+amount)
}

func (self *Character) GetCash() int {
	return self.getField(characterCash).(int)
}

func (self *Character) AddItem(item *Item) {
	itemIds := self.GetItemIds()
	itemIds = append(itemIds, item.GetId())
	self.setField(characterInventory, itemIds)
}

func (self *Character) RemoveItem(item *Item) {
	itemIds := self.GetItemIds()
	for i, itemId := range itemIds {
		if itemId == item.GetId() {
			// TODO: Potential memory leak. See http://code.google.com/p/go-wiki/wiki/SliceTricks
			itemIds = append(itemIds[:i], itemIds[i+1:]...)
			self.setField(characterInventory, itemIds)
			return
		}
	}
}

func (self *Character) GetItemIds() []bson.ObjectId {
	if self.hasField(characterInventory) {
		return self.getField(characterInventory).([]bson.ObjectId)
	}

	return []bson.ObjectId{}
}

// vim: nocindent
