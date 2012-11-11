package engine

import (
	"fmt"
	"kmud/database"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
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
	SayEventType        EventType = iota
	EnterEventType      EventType = iota
	LeaveEventType      EventType = iota
	RoomUpdateEventType EventType = iota
	LoginEventType      EventType = iota
	LogoutEventType     EventType = iota
)

type Event interface {
	Type() EventType
	ToString(receiver database.Character) string
}

type MessageEvent struct {
	Character database.Character
	Message   string
}

type SayEvent struct {
	Character database.Character
	Message   string
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

func (self MessageEvent) ToString(receiver database.Character) string {
	user := GetCharacterUser(receiver)
	cm := user.ColorMode
	return utils.Colorize(cm, utils.ColorBlue, "Message from "+self.Character.PrettyName()+": ") +
		utils.Colorize(cm, utils.ColorWhite, self.Message)
}

func (self SayEvent) Type() EventType {
	return SayEventType
}

func (self SayEvent) ToString(receiver database.Character) string {
	if receiver.RoomId != self.Character.RoomId {
		return ""
	}

	user := GetCharacterUser(receiver)
	cm := user.ColorMode

	who := ""
	if receiver.Id == self.Character.Id {
		who = "You say"
	} else {
		who = self.Character.PrettyName() + " says"
	}

	return utils.Colorize(cm, utils.ColorBlue, who+", ") +
		utils.Colorize(cm, utils.ColorWhite, "\""+self.Message+"\"")
}

func (self EnterEvent) Type() EventType {
	return EnterEventType
}

func (self EnterEvent) ToString(receiver database.Character) string {
	return fmt.Sprintf("%s has entered the room", self.Character.PrettyName())
}

func (self LeaveEvent) Type() EventType {
	return LeaveEventType
}

func (self LeaveEvent) ToString(receiver database.Character) string {
	return fmt.Sprintf("%s has left the room", self.Character.PrettyName())
}

func (self RoomUpdateEvent) Type() EventType {
	return RoomUpdateEventType
}

func (self RoomUpdateEvent) ToString(receiver database.Character) string {
	return "This room has been modified"
}

func (self LoginEvent) Type() EventType {
	return LoginEventType
}

func (self LoginEvent) ToString(receiver database.Character) string {
	return fmt.Sprintf("%s has connected", self.Character.PrettyName())
}

func (self LogoutEvent) Type() EventType {
	return LogoutEventType
}

func (self LogoutEvent) ToString(receiver database.Character) string {
	return fmt.Sprintf("%s has disconnected", self.Character.PrettyName())
}

// vim: nocindent
