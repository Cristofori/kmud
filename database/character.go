package database

import (
	"fmt"

	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Character struct {
	Container `bson:",inline"`

	RoomId    types.Id `bson:",omitempty"`
	Name      string
	HitPoints int
	Skills    utils.Set

	Strength int
	Vitality int
}

type Pc struct {
	Character `bson:",inline"`

	UserId types.Id
	online bool
}

type Npc struct {
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

func NewPc(name string, userId types.Id, roomId types.Id) *Pc {
	pc := &Pc{
		UserId: userId,
		online: false,
	}

	pc.initCharacter(name, types.PcType, roomId)
	dbinit(pc)
	return pc
}

func NewNpc(name string, roomId types.Id, spawnerId types.Id) *Npc {
	npc := &Npc{
		SpawnerId: spawnerId,
	}

	npc.initCharacter(name, types.NpcType, roomId)
	dbinit(npc)
	return npc
}

func NewSpawner(name string, areaId types.Id) *Spawner {
	spawner := &Spawner{
		AreaId: areaId,
		Count:  1,
	}

	spawner.initCharacter(name, types.SpawnerType, nil)
	dbinit(spawner)
	return spawner
}

func (self *Character) initCharacter(name string, objType types.ObjectType, roomId types.Id) {
	self.RoomId = roomId
	self.Cash = 0
	self.HitPoints = 100
	self.Name = utils.FormatName(name)

	self.Strength = 10
	self.Vitality = 100
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

func (self *Character) GetCapacity() int {
	return self.GetStrength() * 10
}

func (self *Character) GetStrength() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Strength
}

func (self *Pc) SetOnline(online bool) {
	self.WriteLock()
	self.online = online
	self.WriteUnlock()
}

func (self *Pc) IsOnline() bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.online
}

func (self *Character) SetRoomId(id types.Id) {
	if id != self.GetRoomId() {
		self.WriteLock()
		self.RoomId = id
		self.WriteUnlock()
		self.modified()
	}
}

func (self *Character) GetRoomId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.RoomId
}

func (self *Pc) SetUserId(id types.Id) {
	self.WriteLock()
	defer self.WriteUnlock()

	if id != self.UserId {
		self.UserId = id
		self.modified()
	}
}

func (self *Pc) GetUserId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.UserId
}

func (self *Character) AddSkill(id types.Id) {
	if !self.HasSkill(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		if self.Skills == nil {
			self.Skills = utils.Set{}
		}

		self.Skills.Insert(id.Hex())
		self.modified()
	}
}

func (self *Character) RemoveSkill(id types.Id) {
	if self.HasSkill(id) {
		self.WriteLock()
		defer self.WriteUnlock()

		self.Skills.Remove(id.Hex())
		self.modified()
	}
}

func (self *Character) HasSkill(id types.Id) bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Skills.Contains(id.Hex())
}

func (self *Character) GetSkills() []types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return idSetToList(self.Skills)
}

func (self *Npc) SetConversation(conversation string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if self.Conversation != conversation {
		self.Conversation = conversation
		self.modified()
	}
}

func (self *Npc) GetConversation() string {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Conversation
}

func (self *Npc) PrettyConversation() string {
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

	if health != self.Vitality {
		self.Vitality = health

		if self.HitPoints > self.Vitality {
			self.HitPoints = self.Vitality
		}

		self.modified()
	}
}

func (self *Character) GetHealth() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Vitality
}

func (self *Character) SetHitPoints(hitpoints int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if hitpoints > self.Vitality {
		hitpoints = self.Vitality
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

func (self *Npc) GetRoaming() bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Roaming
}

func (self *Npc) SetRoaming(roaming bool) {
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
