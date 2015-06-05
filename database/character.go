package database

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"kmud/utils"
)

type Character struct {
	DbObject `bson:",inline"`
	RoomId   bson.ObjectId `bson:",omitempty"`

	Name      string
	Cash      int
	Inventory []bson.ObjectId
	Health    int
	HitPoints int

	objType ObjectType
}

type NonPlayerChar struct {
	Character `bson:",inline"`

	Roaming      bool
	Conversation string
}

type PlayerChar struct {
	Character `bson:",inline"`

	UserId bson.ObjectId
	online bool
}

type CharacterList []*Character
type PlayerCharList []*PlayerChar
type NonPlayerCharList []*NonPlayerChar

func initCharacter(character *Character, name string, objType ObjectType, roomId bson.ObjectId) {
	character.SetRoomId(roomId)
	character.SetCash(0)
	character.SetHealth(100)
	character.SetHitPoints(100)
	character.SetName(utils.FormatName(name))
	character.objType = objType

	character.initDbObject()
}

func NewPlayerChar(name string, userId bson.ObjectId, roomId bson.ObjectId) *PlayerChar {
	var player PlayerChar

	player.UserId = userId
	player.online = false

	initCharacter(&player.Character, name, PcType, roomId)
	commitObject(&player)
	return &player
}

func NewNonPlayerChar(name string, roomId bson.ObjectId) *NonPlayerChar {
	var npc NonPlayerChar
	initCharacter(&npc.Character, name, NpcType, roomId)
	commitObject(&npc)
	return &npc
}

func (self *Character) GetType() ObjectType {
	return self.objType
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
		objectModified(self)
	}
}

/*
func NewNpcTemplate(name string) *Character {
	return NewCharacter(name, "", "")
}

func NewNpcFromTemplate(template *Character, roomId bson.ObjectId) *Character {
	return NewNpc(template.GetName(), template.GetRoomId())
}
*/

func (self *PlayerChar) SetOnline(online bool) {
	self.WriteLock()
	self.online = online
	self.WriteUnlock()
}

func (self *PlayerChar) IsOnline() bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.online
}

/*
func (self *Character) IsNpcTemplate() bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.UserId == "" && self.RoomId == ""
}
*/

func (self *Character) SetRoomId(id bson.ObjectId) {
	self.WriteLock()
	defer self.WriteUnlock()

	if id != self.RoomId {
		self.RoomId = id
		objectModified(self)
	}
}

func (self *Character) GetRoomId() bson.ObjectId {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.RoomId
}

func (self *PlayerChar) SetUserId(id bson.ObjectId) {
	self.WriteLock()
	defer self.WriteUnlock()

	if id != self.UserId {
		self.UserId = id
		objectModified(self)
	}
}

func (self *PlayerChar) GetUserId() bson.ObjectId {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.UserId
}

func (self *Character) SetCash(cash int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if cash != self.Cash {
		self.Cash = cash
		objectModified(self)
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
		objectModified(self)
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

		objectModified(self)
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

func (self *NonPlayerChar) SetConversation(conversation string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Conversation != conversation {
		self.Conversation = conversation
		objectModified(self)
	}
}

func (self *NonPlayerChar) GetConversation() string {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Conversation
}

func (self *NonPlayerChar) PrettyConversation() string {
	conv := self.GetConversation()

	if conv == "" {
		return fmt.Sprintf("%s has nothing to say", self.GetName())
	}

	return fmt.Sprintf("%s%s",
		utils.Colorize(utils.ColorBlue, self.GetName()),
		utils.Colorize(utils.ColorWhite, ": "+conv))
}

func (self *Character) SetHealth(health int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if health != self.Health {
		self.Health = health

		if self.HitPoints > self.Health {
			self.HitPoints = self.Health
		}

		objectModified(self)
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
		objectModified(self)
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

func (self *NonPlayerChar) GetRoaming() bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Roaming
}

func (self *NonPlayerChar) SetRoaming(roaming bool) {
	self.WriteLock()
	defer self.WriteUnlock()

	self.Roaming = roaming
	objectModified(self)
}

func (self PlayerCharList) Characters() CharacterList {
	chars := make([]*Character, len(self))

	for i, char := range self {
		chars[i] = &char.Character
	}

	return chars
}

func (self NonPlayerCharList) Characters() CharacterList {
	chars := make([]*Character, len(self))

	for i, npc := range self {
		chars[i] = &npc.Character
	}

	return chars
}

func (self CharacterList) Names() []string {
	names := make([]string, len(self))

	for i, char := range self {
		names[i] = char.GetName()
	}

	return names
}

// vim: nocindent
