package engine

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"mud/database"
)

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

type EventType int

const (
	MessageEventType    EventType = iota
	EnterEventType      EventType = iota
	LeaveEventType      EventType = iota
	RoomUpdateEventType EventType = iota
)

type Event interface {
	Type() EventType
	ToString() string
}

type MessageEvent struct {
	Message string
}

type EnterEvent struct {
	Character database.Character
	RoomId    bson.ObjectId
}

type LeaveEvent struct {
	Character database.Character
	RoomId    bson.ObjectId
}

type RoomUpdateEvent struct {
	Room database.Room
}

func (self MessageEvent) Type() EventType {
	return MessageEventType
}

func (self MessageEvent) ToString() string {
	return self.Message
}

func (self EnterEvent) Type() EventType {
	return EnterEventType
}

func (self EnterEvent) ToString() string {
	return fmt.Sprintf("%s has entered the room", self.Character.PrettyName())
}

func (self LeaveEvent) Type() EventType {
	return LeaveEventType
}

func (self LeaveEvent) ToString() string {
	return fmt.Sprintf("%s has left the room", self.Character.PrettyName())
}

func (self RoomUpdateEvent) Type() EventType {
	return RoomUpdateEventType
}

func (self RoomUpdateEvent) ToString() string {
	return "This room has been modified"
}

// vim: nocindent
