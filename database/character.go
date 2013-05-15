package database

import (
	"fmt"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
)

const (
	characterRoomId       string = "roomid"
	characterUserId       string = "userid"
	characterCash         string = "cash"
	characterInventory    string = "inventory"
	characterConversation string = "conversation"
	characterHealth       string = "health"
	characterHitPoints    string = "hitpoints"
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
	character.SetHealth(100)
	character.SetHitPoints(100)

	character.SetOnline(false)

	return &character
}

func NewNpc(name string, roomId bson.ObjectId) *Character {
	return NewCharacter(name, "", roomId)
}

func (self *Character) SetOnline(online bool) {
	self.mutex.Lock()
	self.online = online
	self.mutex.Unlock()
}

func (self *Character) IsOnline() bool {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.online || self.IsNpc()
}

func (self *Character) IsNpc() bool {
	return !self.hasField(characterUserId)
}

func (self *Character) IsPlayer() bool {
	return !self.IsNpc()
}

func (self *Character) GetRoomId() bson.ObjectId {
	return self.getField(characterRoomId).(bson.ObjectId)
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

func (self *Character) HasItem(item *Item) bool {
	items := self.GetItemIds()

	for _, itemId := range items {
		if itemId == item.GetId() {
			return true
		}
	}

	return false
}

func (self *Character) GetItemIds() []bson.ObjectId {
	if self.hasField(characterInventory) {
		return self.getField(characterInventory).([]bson.ObjectId)
	}

	return []bson.ObjectId{}
}

func (self *Character) GetConversation() string {
	if self.hasField(characterConversation) {
		return self.getField(characterConversation).(string)
	}

	return ""
}

func (self *Character) PrettyConversation(cm utils.ColorMode) string {
	conv := self.GetConversation()

	if conv == "" {
		return fmt.Sprintf("%s has nothing to say", self.PrettyName())
	}

	return fmt.Sprintf("%s%s",
		utils.Colorize(cm, utils.ColorBlue, self.PrettyName()),
		utils.Colorize(cm, utils.ColorWhite, ": "+conv))
}

func (self *Character) SetConversation(conversation string) {
	self.setField(characterConversation, conversation)
}

func CharacterNames(characters []*Character) []string {
	names := make([]string, len(characters))

	for i, char := range characters {
		names[i] = char.PrettyName()
	}

	return names
}

func (self *Character) SetHealth(health int) {
	self.setField(characterHealth, health)

	// TODO - Fix this after we get around to using the mark and sweep database back-end
	//        and we don't have to deal with all the type assertions that we use and
	//        nil interfaces we get.

	// This code panics
	// if self.GetHitPoints() > self.GetHealth() {
	//     self.SetHitPoints(self.GetHealth())
	// }
}

func (self *Character) GetHealth() int {
	return self.getField(characterHealth).(int)
}

func (self *Character) SetHitPoints(hitpoints int) {
	if hitpoints > self.GetHealth() {
		hitpoints = self.GetHealth()
	}

	self.setField(characterHitPoints, hitpoints)
}

func (self *Character) GetHitPoints() int {
	return self.getField(characterHitPoints).(int)
}

func (self *Character) Hit(hitpoints int) int {
	self.SetHitPoints(self.GetHitPoints() - hitpoints)
	return self.GetHitPoints()
}

func (self *Character) Heal(hitpoints int) {
	self.SetHitPoints(self.GetHitPoints() + hitpoints)
}

// vim: nocindent
