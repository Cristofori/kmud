package types

import (
	"gopkg.in/mgo.v2/bson"
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

type ReadLocker interface {
	ReadLock()
	ReadUnlock()
}

type Destroyable interface {
	Destroy()
	IsDestroyed() bool
}

type Nameable interface {
	GetName() string
}

type Object interface {
	Identifiable
	ReadLocker
	Destroyable
}

type Character interface {
	Identifiable
	Nameable
	GetRoomId() bson.ObjectId
	SetRoomId(bson.ObjectId)
	IsOnline() bool
}

type CharacterList []Character

func (self CharacterList) Names() []string {
	names := make([]string, len(self))

	for i, char := range self {
		names[i] = char.GetName()
	}

	return names
}

type NPC interface {
	Character
	GetRoaming() bool
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

func (self NPCList) Names() []string {
	names := make([]string, len(self))

	for i, npc := range self {
		names[i] = npc.GetName()
	}

	return names
}

type Zone interface {
	Identifiable
}
