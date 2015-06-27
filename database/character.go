package database

import (
	"fmt"

	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Character struct {
	DbObject `bson:",inline"`
	RoomId   types.Id `bson:",omitempty"`

	Name      string
	Cash      int
	Inventory []types.Id
	Health    int
	HitPoints int

	objType types.ObjectType
}

type PlayerChar struct {
	Character `bson:",inline"`

	UserId types.Id
	online bool
}

type NonPlayerChar struct {
	Character `bson:",inline"`

	SpawnerId types.Id `bson:",omitempty"`

	Roaming      bool
	Conversation string
}

type Spawner struct {
	Character `bson:",inline"`

	AreaId types.Id
	Count  int
}

func NewPlayerChar(name string, userId types.Id, roomId types.Id) *PlayerChar {
	var pc PlayerChar

	pc.UserId = userId
	pc.online = false

	pc.initCharacter(name, types.PcType, roomId)
	pc.initDbObject(&pc)
	return &pc
}

func NewNonPlayerChar(name string, roomId types.Id, spawnerId types.Id) *NonPlayerChar {
	var npc NonPlayerChar

	npc.SpawnerId = spawnerId

	npc.initCharacter(name, types.NpcType, roomId)
	npc.initDbObject(&npc)
	return &npc
}

func NewSpawner(name string, areaId types.Id) *Spawner {
	var spawner Spawner

	spawner.AreaId = areaId
	spawner.Count = 1

	spawner.initCharacter(name, types.SpawnerType, nil)
	spawner.initDbObject(&spawner)
	return &spawner
}

func (self *Character) initCharacter(name string, objType types.ObjectType, roomId types.Id) {
	self.RoomId = roomId
	self.Cash = 0
	self.Health = 100
	self.HitPoints = 100
	self.Name = utils.FormatName(name)
	self.objType = objType
}

func (self *Character) GetType() types.ObjectType {
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
		self.modified()
	}
}

// Used when loading existing characters from the DB
func (self *Character) SetObjectType(t types.ObjectType) {
	self.objType = t
}

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

func (self *Character) SetRoomId(id types.Id) {
	self.WriteLock()
	defer self.WriteUnlock()

	if id != self.RoomId {
		self.RoomId = id
		self.modified()
	}
}

func (self *Character) GetRoomId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.RoomId
}

func (self *PlayerChar) SetUserId(id types.Id) {
	self.WriteLock()
	defer self.WriteUnlock()

	if id != self.UserId {
		self.UserId = id
		self.modified()
	}
}

func (self *PlayerChar) GetUserId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.UserId
}

func (self *Character) SetCash(cash int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if cash != self.Cash {
		self.Cash = cash
		self.modified()
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

func (self *Character) AddItem(id types.Id) {
	if !self.HasItem(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		self.Inventory = append(self.Inventory, id)
		self.modified()
	}
}

func (self *Character) RemoveItem(id types.Id) {
	if self.HasItem(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		for i, itemId := range self.Inventory {
			if itemId == id {
				// TODO: Potential memory leak. See http://code.google.com/p/go-wiki/wiki/SliceTricks
				self.Inventory = append(self.Inventory[:i], self.Inventory[i+1:]...)
				break
			}
		}

		self.modified()
	}
}

func (self *Character) HasItem(id types.Id) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	for _, itemId := range self.Inventory {
		if itemId == id {
			return true
		}
	}

	return false
}

func (self *Character) GetItemIds() []types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Inventory
}

func (self *NonPlayerChar) SetConversation(conversation string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Conversation != conversation {
		self.Conversation = conversation
		self.modified()
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
		types.Colorize(types.ColorBlue, self.GetName()),
		types.Colorize(types.ColorWhite, ": "+conv))
}

func (self *Character) SetHealth(health int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if health != self.Health {
		self.Health = health

		if self.HitPoints > self.Health {
			self.HitPoints = self.Health
		}

		self.modified()
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
		self.modified()
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
	self.modified()
}

func (self *Spawner) SetCount(count int) {
	self.WriteLock()
	defer self.WriteUnlock()

	self.Count = count
	self.modified()
}

func (self *Spawner) GetCount() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Count
}

func (self *Spawner) GetAreaId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.AreaId
}

// vim: nocindent
