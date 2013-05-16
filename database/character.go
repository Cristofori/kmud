package database

import (
	"fmt"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
)

type Character struct {
	DbObject `bson:",inline"`

	RoomId       bson.ObjectId
	UserId       bson.ObjectId
	Cash         int
	Inventory    []bson.ObjectId
	Conversation string
	Health       int
	HitPoints    int

	online bool
}

func NewCharacter(name string, userId bson.ObjectId, roomId bson.ObjectId) *Character {
	var character Character
	character.initDbObject(name, charType)

	character.UserId = userId
	character.RoomId = roomId
	character.Cash = 0
	character.Health = 100
	character.HitPoints = 100

	character.online = false

	modified(&character)
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
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.UserId == ""
}

func (self *Character) IsPlayer() bool {
	return !self.IsNpc()
}

func (self *Character) SetRoomId(id bson.ObjectId) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if id != self.RoomId {
		self.RoomId = id
		modified(self)
	}
}

func (self *Character) GetRoomId() bson.ObjectId {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.RoomId
}

func (self *Character) SetUserId(id bson.ObjectId) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if id != self.UserId {
		self.UserId = id
		modified(self)
	}
}

func (self *Character) GetUserId() bson.ObjectId {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.UserId
}

func (self *Character) SetCash(cash int) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if cash != self.Cash {
		self.Cash = cash
		modified(self)
	}
}

func (self *Character) AddCash(amount int) {
	self.SetCash(self.GetCash() + amount)
}

func (self *Character) GetCash() int {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.Cash
}

func (self *Character) AddItem(item *Item) {
	if !self.HasItem(item) {
		self.mutex.Lock()
		defer self.mutex.Unlock()

		self.Inventory = append(self.Inventory, item.GetId())
		modified(self)
	}
}

func (self *Character) RemoveItem(item *Item) {
	if self.HasItem(item) {
		self.mutex.Lock()
		defer self.mutex.Unlock()

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
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	for _, itemId := range self.Inventory {
		if itemId == item.GetId() {
			return true
		}
	}

	return false
}

func (self *Character) GetItemIds() []bson.ObjectId {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.Inventory
}

func (self *Character) SetConversation(conversation string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.Conversation != conversation {
		self.Conversation = conversation
		modified(self)
	}
}

func (self *Character) GetConversation() string {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.Conversation
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

func (self *Character) SetHealth(health int) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if health != self.Health {
		self.Health = health

		if self.HitPoints > self.Health {
			self.HitPoints = self.Health
		}

		modified(self)
	}
}

func (self *Character) GetHealth() int {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.Health
}

func (self *Character) SetHitPoints(hitpoints int) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if hitpoints > self.Health {
		hitpoints = self.Health
	}

	if hitpoints != self.HitPoints {
		self.HitPoints = hitpoints
		modified(self)
	}
}

func (self *Character) GetHitPoints() int {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

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
		names[i] = char.PrettyName()
	}

	return names
}

// vim: nocindent
