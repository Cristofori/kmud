package types

import (
	"net"

	"github.com/Cristofori/kmud/utils/naturalsort"
)

type Id interface {
	String() string
	Hex() string
}

type ObjectType string

const (
	NpcType     ObjectType = "Npc"
	PcType      ObjectType = "Pc"
	SpawnerType ObjectType = "Spawner"
	UserType    ObjectType = "User"
	ZoneType    ObjectType = "Zone"
	AreaType    ObjectType = "Area"
	RoomType    ObjectType = "Room"
	ItemType    ObjectType = "Item"
	SkillType   ObjectType = "Skill"
	StoreType   ObjectType = "Store"
)

type Identifiable interface {
	GetId() Id
}

type ReadLockable interface {
	ReadLock()
	ReadUnlock()
}

type Destroyable interface {
	Destroy()
	IsDestroyed() bool
}

type Locateable interface {
	GetRoomId() Id
}

type Nameable interface {
	GetName() string
	SetName(string)
}

type Loginable interface {
	IsOnline() bool
	SetOnline(bool)
}

type Container interface {
	AddItem(Id)
	RemoveItem(Id)
	GetItems() []Id
}

type Object interface {
	Identifiable
	ReadLockable
	Destroyable
}

type Character interface {
	Object
	Nameable
	Locateable
	Container
	SetRoomId(Id)
	AddCash(int)
	GetCash() int
	Hit(int)
	Heal(int)
	GetHitPoints() int
	SetHitPoints(int)
	GetHealth() int
	SetHealth(int)
	GetSkills() []Id
	AddSkill(Id)
}

type CharacterList []Character

func (self CharacterList) Names() []string {
	names := make([]string, len(self))
	for i, char := range self {
		names[i] = char.GetName()
	}
	return names
}

type PC interface {
	Character
	Loginable
}

type PCList []PC

func (self PCList) Characters() CharacterList {
	chars := make(CharacterList, len(self))
	for i, pc := range self {
		chars[i] = pc
	}
	return chars
}

type NPC interface {
	Character
	SetRoaming(bool)
	GetRoaming() bool
	SetConversation(string)
	GetConversation() string
	PrettyConversation() string
}

type NPCList []NPC

type Spawner interface {
	Character
	GetAreaId() Id
	SetCount(int)
	GetCount() int
}

type SpawnerList []Spawner

func (self NPCList) Characters() CharacterList {
	chars := make(CharacterList, len(self))
	for i, npc := range self {
		chars[i] = npc
	}
	return chars
}

type Room interface {
	Object
	Container
	GetZoneId() Id
	GetAreaId() Id
	SetAreaId(Id)
	GetLocation() Coordinate
	SetExitEnabled(Direction, bool)
	HasExit(Direction) bool
	NextLocation(Direction) Coordinate
	GetExits() []Direction
	GetTitle() string
	SetTitle(string)
	GetDescription() string
	SetDescription(string)
	GetProperties() map[string]string
	SetProperty(string, string)
	RemoveProperty(string)
	SetLink(string, Id)
	RemoveLink(string)
	GetLinks() map[string]Id
	LinkNames() []string
	SetLocked(Direction, bool)
	IsLocked(Direction) bool
}

type RoomList []Room

type Area interface {
	Object
	Nameable
}

type AreaList []Area

type Zone interface {
	Object
	Nameable
}

type ZoneList []Zone

type Communicable interface {
	WriteLine(string)
	Write(string)
	GetInput(prompt string) string
}

type User interface {
	Object
	Nameable
	Loginable
	Communicable
	VerifyPassword(string) bool
	SetConnection(net.Conn)
	GetConnection() net.Conn
	SetWindowSize(int, int)
	GetWindowSize() (int, int)
	SetTerminalType(string)
	GetTerminalType() string
	GetColorMode() ColorMode
	SetColorMode(ColorMode)
	IsAdmin() bool
	SetAdmin(bool)
}

type UserList []User

func (self UserList) Len() int {
	return len(self)
}

func (self UserList) Less(i, j int) bool {
	return naturalsort.NaturalLessThan(self[i].GetName(), self[j].GetName())
}

func (self UserList) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type Item interface {
	Object
	Nameable
}

type ItemList []Item

func (self ItemList) Names() []string {
	names := make([]string, len(self))
	for i, item := range self {
		names[i] = item.GetName()
	}
	return names
}

type Skill interface {
	Object
	Nameable
	SetPower(int)
	GetPower() int
	SetCost(int)
	GetCost() int
	GetEffect() SkillEffect
	SetEffect(SkillEffect)
	GetVariance() int
	SetVariance(int)
	GetSpeed() int
	SetSpeed(int)
}

type SkillList []Skill

type SkillEffect string

const (
	DamageEffect  SkillEffect = "damage"
	DotEffect     SkillEffect = "dot"
	HealingEffect SkillEffect = "heal"
	HotEffect     SkillEffect = "hot"
)

func (self SkillList) Names() []string {
	names := make([]string, len(self))
	for i, skill := range self {
		names[i] = skill.GetName()
	}
	return names
}

type Store interface {
	Object
	Nameable
	Container
}
