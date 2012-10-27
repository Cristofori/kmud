package engine

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"mud/database"
)

var _listeners map[*chan Event]database.Character

func Register(character database.Character) *chan Event {
	_mutex.Lock()
	defer _mutex.Unlock()

	if _listeners == nil {
		_listeners = map[*chan Event]database.Character{}
	}

	listener := make(chan Event, 100)
	_listeners[&listener] = character

	character.SetOnline(true)
	_model.Characters[character.Id] = character

	queueEvent(LoginEvent{character})

	return &listener
}

func Unregister(listener *chan Event) {
	_mutex.Lock()
	defer _mutex.Unlock()

	character := _listeners[listener]
	character.SetOnline(false)
	_model.Characters[character.Id] = character

	queueEvent(LogoutEvent{character})
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
	LoginEventType      EventType = iota
	LogoutEventType     EventType = iota
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

type LoginEvent struct {
	Character database.Character
}

type LogoutEvent struct {
	Character database.Character
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

func (self LoginEvent) Type() EventType {
	return LoginEventType
}

func (self LoginEvent) ToString() string {
	return fmt.Sprintf("%s has connected", self.Character.PrettyName())
}

func (self LogoutEvent) Type() EventType {
	return LogoutEventType
}

func (self LogoutEvent) ToString() string {
	return fmt.Sprintf("%s has disconnected", self.Character.PrettyName())
}

// vim: nocindent
