package types

import (
	"github.com/Cristofori/kmud/utils/naturalsort"
	"gopkg.in/mgo.v2/bson"
	"net"
)

type ObjectType int

const (
	NpcType  ObjectType = iota
	PcType   ObjectType = iota
	UserType ObjectType = iota
	ZoneType ObjectType = iota
	AreaType ObjectType = iota
	RoomType ObjectType = iota
	ItemType ObjectType = iota
)

type Identifiable interface {
	GetId() bson.ObjectId
	GetType() ObjectType
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
	GetRoomId() bson.ObjectId
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
	AddItem(bson.ObjectId)
	RemoveItem(bson.ObjectId)
	GetItemIds() []bson.ObjectId
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
	SetRoomId(bson.ObjectId)
	AddCash(int)
	GetCash() int
	Hit(int)
	Heal(int)
	GetHitPoints() int
	GetHealth() int
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
	GetZoneId() bson.ObjectId
	GetAreaId() bson.ObjectId
	SetAreaId(bson.ObjectId)
	GetLocation() Coordinate
	SetExitEnabled(Direction, bool)
	HasExit(Direction) bool
	NextLocation(Direction) Coordinate
	GetExits() []Direction
	ToString(PCList, NPCList, ItemList, Area) string
	SetTitle(string)
	SetDescription(string)
	GetProperties() map[string]string
	SetProperty(string, string)
	RemoveProperty(string)
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

type User interface {
	Object
	Nameable
	Loginable
	VerifyPassword(string) bool
	SetConnection(net.Conn)
	GetConnection() net.Conn
	SetWindowSize(int, int)
	GetWindowSize() (int, int)
	SetTerminalType(string)
	GetTerminalType() string
	GetColorMode() ColorMode
	WriteLine(string) (int, error)
	GetInput(string) string
	SetColorMode(ColorMode)
	Write(string) (int, error)
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
