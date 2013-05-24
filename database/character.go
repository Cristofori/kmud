package database

import (
	"fmt"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
)

type Character struct {
	DbObject `bson:",inline"`

	RoomId bson.ObjectId
	UserId bson.ObjectId

	Name         string
	Cash         int
	Inventory    []bson.ObjectId
	Conversation string
	Health       int
	HitPoints    int

	online bool
}

func NewCharacter(name string, userId bson.ObjectId, roomId bson.ObjectId) *Character {
	var character Character
	character.initDbObject()

	character.UserId = userId
	character.RoomId = roomId
	character.Cash = 0
	character.Health = 100
	character.HitPoints = 100
	character.Name = utils.FormatName(name)

	character.online = false

	modified(&character)
	return &character
}

func (self *Character) GetType() objectType {
	return CharType
}

func (self *Character) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *Character) SetName(name string) {
	if name != self.GetName() {
		self.WriteLock()
		self.Name = utils.FormatName(name)
		self.WriteUnlock()
		modified(self)
	}
}

func NewNpc(name string, roomId bson.ObjectId) *Character {
	return NewCharacter(name, "", roomId)
}

func (self *Character) SetOnline(online bool) {
	self.WriteLock()
	self.online = online
	self.WriteUnlock()
}

func (self *Character) IsOnline() bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.online || self.IsNpc()
}

func (self *Character) IsNpc() bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.UserId == ""
}

func (self *Character) IsPlayer() bool {
	return !self.IsNpc()
}

func (self *Character) SetRoomId(id bson.ObjectId) {
	self.WriteLock()
	defer self.WriteUnlock()

	if id != self.RoomId {
		self.RoomId = id
		modified(self)
	}
}

func (self *Character) GetRoomId() bson.ObjectId {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.RoomId
}

func (self *Character) SetUserId(id bson.ObjectId) {
	self.WriteLock()
	defer self.WriteUnlock()

	if id != self.UserId {
		self.UserId = id
		modified(self)
	}
}

func (self *Character) GetUserId() bson.ObjectId {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.UserId
}

func (self *Character) SetCash(cash int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if cash != self.Cash {
		self.Cash = cash
		modified(self)
	}
}

func (self *Character) AddCash(amount int) {
	self.SetCash(self.GetCash() + amount)
}

func (self *Character) GetCash() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Cash
}

func (self *Character) AddItem(item *Item) {
	if !self.HasItem(item) {
		self.WriteLock()
		defer self.WriteUnlock()

		self.Inventory = append(self.Inventory, item.GetId())
		modified(self)
	}
}

func (self *Character) RemoveItem(item *Item) {
	if self.HasItem(item) {
		self.WriteLock()
		defer self.WriteUnlock()

		for i, itemId := range self.Inventory {
			if itemId == item.GetId() {
				// TODO: Potential memory leak. See http://code.google.com/p/go-wiki/wiki/SliceTricks
				self.Inventory = append(self.Inventory[:i], self.Inventory[i+1:]...)
				break
			}
		}

		modified(self)
	}
}

func (self *Character) HasItem(item *Item) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	for _, itemId := range self.Inventory {
		if itemId == item.GetId() {
			return true
		}
	}

	return false
}

func (self *Character) GetItemIds() []bson.ObjectId {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Inventory
}

func (self *Character) SetConversation(conversation string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Conversation != conversation {
		self.Conversation = conversation
		modified(self)
	}
}

func (self *Character) GetConversation() string {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Conversation
}

func (self *Character) PrettyConversation(cm utils.ColorMode) string {
	conv := self.GetConversation()

	if conv == "" {
		return fmt.Sprintf("%s has nothing to say", self.GetName())
	}

	return fmt.Sprintf("%s%s",
		utils.Colorize(cm, utils.ColorBlue, self.GetName()),
		utils.Colorize(cm, utils.ColorWhite, ": "+conv))
}

func (self *Character) SetHealth(health int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if health != self.Health {
		self.Health = health

		if self.HitPoints > self.Health {
			self.HitPoints = self.Health
		}

		modified(self)
	}
}

func (self *Character) GetHealth() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Health
}

func (self *Character) SetHitPoints(hitpoints int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if hitpoints > self.Health {
		hitpoints = self.Health
	}

	if hitpoints != self.HitPoints {
		self.HitPoints = hitpoints
		modified(self)
	}
}

func (self *Character) GetHitPoints() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.HitPoints
}

func (self *Character) Hit(hitpoints int) {
	self.SetHitPoints(self.GetHitPoints() - hitpoints)
}

func (self *Character) Heal(hitpoints int) {
	self.SetHitPoints(self.GetHitPoints() + hitpoints)
}

func CharacterNames(characters []*Character) []string {
	names := make([]string, len(characters))

	for i, char := range characters {
		names[i] = char.GetName()
	}

	return names
}

// vim: nocindent
