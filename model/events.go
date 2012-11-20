package model

import (
	"fmt"
	"kmud/database"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
	"sync"
)

var _listeners map[*chan Event]*database.Character

var _mutex sync.Mutex

func Register(character *database.Character) *chan Event {
	_mutex.Lock()
	if _listeners == nil {
		_listeners = map[*chan Event]*database.Character{}
	}
	_mutex.Unlock()

	listener := make(chan Event, 100)

	_mutex.Lock()
	_listeners[&listener] = character
	_mutex.Unlock()

	character.SetOnline(true)
	M.UpdateCharacter(*character) // TODO: Avoid unnecessary database call

	queueEvent(LoginEvent{*character})

	return &listener
}

func Unregister(listener *chan Event) {
	_mutex.Lock()
	character := _listeners[listener]
	_mutex.Unlock()

	character.SetOnline(false)
	M.UpdateCharacter(*character) // TODO: Avoid unnecessary database call

	queueEvent(LogoutEvent{*character})
	delete(_listeners, listener)
}

func broadcast(event Event) {
	for listener := range _listeners {
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
	cm := getColorMode(receiver)
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

	cm := getColorMode(receiver)

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
	if receiver.RoomId != self.RoomId {
		return ""
	}

	if receiver.Id == self.Character.Id {
		return ""
	}

	cm := getColorMode(receiver)

	return utils.Colorize(cm, utils.ColorBlue, self.Character.PrettyName()) +
		utils.Colorize(cm, utils.ColorWhite, " has entered the room")
}

func (self LeaveEvent) Type() EventType {
	return LeaveEventType
}

func (self LeaveEvent) ToString(receiver database.Character) string {
	if receiver.RoomId != self.RoomId {
		return ""
	}

	if receiver.Id == self.Character.Id {
		return ""
	}

	cm := getColorMode(receiver)

	return utils.Colorize(cm, utils.ColorBlue, self.Character.PrettyName()) +
		utils.Colorize(cm, utils.ColorWhite, " has left the room")
}

func (self RoomUpdateEvent) Type() EventType {
	return RoomUpdateEventType
}

func (self RoomUpdateEvent) ToString(receiver database.Character) string {
	if receiver.RoomId != self.Room.Id {
		return ""
	}

	cm := getColorMode(receiver)

	return utils.Colorize(cm, utils.ColorWhite, "This room has been modified")
}

func (self LoginEvent) Type() EventType {
	return LoginEventType
}

func (self LoginEvent) ToString(receiver database.Character) string {
	if receiver.Id == self.Character.Id {
		return ""
	}

	cm := getColorMode(receiver)

	return utils.Colorize(cm, utils.ColorBlue, self.Character.PrettyName()) +
		utils.Colorize(cm, utils.ColorWhite, " has connected")
}

func (self LogoutEvent) Type() EventType {
	return LogoutEventType
}

func (self LogoutEvent) ToString(receiver database.Character) string {
	return fmt.Sprintf("%s has disconnected", self.Character.PrettyName())
}

func getColorMode(char database.Character) utils.ColorMode {
	user := M.GetUser(char.UserId)
	return user.ColorMode
}

// vim: nocindent
