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
}

type Zone interface {
	Identifiable
}
