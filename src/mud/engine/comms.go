package engine

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"mud/database"
)

type EventType int

const (
	MessageEventType EventType = iota
	MoveEventType    EventType = iota
)

type Event interface {
	Type() EventType
	ToString() string
}

var _listeners map[*chan Event]bool

func Register() *chan Event {
	if _listeners == nil {
		_listeners = map[*chan Event]bool{}
	}

	listener := make(chan Event, 100)
	_listeners[&listener] = true
	return &listener
}

func Unregister(listener *chan Event) {
	delete(_listeners, listener)
}

func broadcast(event Event) {
	for listener, _ := range _listeners {
		*listener <- event
	}
}

type MessageEvent struct {
	Message string
}

type MoveEvent struct {
	Character database.Character
	RoomId    bson.ObjectId
}

func (self MessageEvent) Type() EventType {
	return MessageEventType
}

func (self MessageEvent) ToString() string {
	return self.Message
}

func (self MoveEvent) Type() EventType {
	return MoveEventType
}

func (self MoveEvent) ToString() string {
	return fmt.Sprintf("%s has entered the room", self.Character.PrettyName())
}

// vim: nocindent
