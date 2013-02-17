package model

import (
	"container/list"
	"fmt"
	"kmud/database"
	"kmud/utils"
	"labix.org/v2/mgo/bson"
	"sync"
)

var _listeners map[chan Event]*database.Character

var _mutex sync.Mutex

var eventQueueChannel chan Event

func Register(character *database.Character) chan Event {
	_mutex.Lock()
	if _listeners == nil {
		_listeners = map[chan Event]*database.Character{}
	}
	_mutex.Unlock()

	listener := make(chan Event, 100)

	_mutex.Lock()
	_listeners[listener] = character
	_mutex.Unlock()

	character.SetOnline(true)

	queueEvent(LoginEvent{character})

	return listener
}

func Unregister(listener chan Event) {
	_mutex.Lock()
	character := _listeners[listener]
	_mutex.Unlock()

	character.SetOnline(false)

	queueEvent(LogoutEvent{character})
	delete(_listeners, listener)
}

func eventLoop() {
	var m sync.Mutex
	cond := sync.NewCond(&m)

	eventQueue := list.New()

	go func() {
		for {
			event := <-eventQueueChannel

			cond.L.Lock()
			eventQueue.PushBack(event)
			cond.L.Unlock()
			cond.Signal()
		}
	}()

	for {
		cond.L.Lock()
		for eventQueue.Len() == 0 {
			cond.Wait()
		}

		event := eventQueue.Remove(eventQueue.Front())
		cond.L.Unlock()

		broadcast(event.(Event))
	}
}

func queueEvent(event Event) {
	eventQueueChannel <- event
}

func broadcast(event Event) {
	for listener := range _listeners {
		listener <- event
	}
}

type EventType int

const (
	BroadcastEventType   EventType = iota
	SayEventType         EventType = iota
	EmoteEventType       EventType = iota
	TellEventType        EventType = iota
	EnterEventType       EventType = iota
	LeaveEventType       EventType = iota
	RoomUpdateEventType  EventType = iota
	LoginEventType       EventType = iota
	LogoutEventType      EventType = iota
	AttackStartEventType EventType = iota
	AttackStopEventType  EventType = iota
)

type Event interface {
	Type() EventType
	ToString(receiver *database.Character) string
}

type BroadcastEvent struct {
	Character *database.Character
	Message   string
}

type SayEvent struct {
	Character *database.Character
	Message   string
}

type EmoteEvent struct {
	Character *database.Character
	Emote     string
}

type TellEvent struct {
	From    *database.Character
	To      *database.Character
	Message string
}

type EnterEvent struct {
	Character *database.Character
	RoomId    bson.ObjectId
}

type LeaveEvent struct {
	Character *database.Character
	RoomId    bson.ObjectId
}

type RoomUpdateEvent struct {
	Room *database.Room
}

type LoginEvent struct {
	Character *database.Character
}

type LogoutEvent struct {
	Character *database.Character
}

type AttackStartEvent struct {
	Attacker *database.Character
	Defender *database.Character
}

type AttackStopEvent struct {
	Attacker *database.Character
	Defender *database.Character
}

func (self BroadcastEvent) Type() EventType {
	return BroadcastEventType
}

func (self BroadcastEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)
	return utils.Colorize(cm, utils.ColorCyan, "Broadcast from "+self.Character.PrettyName()+": ") +
		utils.Colorize(cm, utils.ColorWhite, self.Message)
}

func (self SayEvent) Type() EventType {
	return SayEventType
}

func (self SayEvent) ToString(receiver *database.Character) string {
	if receiver.GetRoomId() != self.Character.GetRoomId() {
		return ""
	}

	cm := getColorMode(receiver)

	who := ""
	if receiver.GetId() == self.Character.GetId() {
		who = "You say"
	} else {
		who = self.Character.PrettyName() + " says"
	}

	return utils.Colorize(cm, utils.ColorBlue, who+", ") +
		utils.Colorize(cm, utils.ColorWhite, "\""+self.Message+"\"")
}

func (self EmoteEvent) Type() EventType {
	return EmoteEventType
}

func (self EmoteEvent) ToString(receiver *database.Character) string {
	if receiver.GetRoomId() != self.Character.GetRoomId() {
		return ""
	}

	cm := getColorMode(receiver)

	return utils.Colorize(cm, utils.ColorYellow, self.Character.PrettyName()+" "+self.Emote)
}

func (self TellEvent) Type() EventType {
	return TellEventType
}

func (self TellEvent) ToString(receiver *database.Character) string {
	if receiver.GetId() != self.To.GetId() {
		return ""
	}

	cm := getColorMode(receiver)

	return utils.Colorize(cm, utils.ColorMagenta, fmt.Sprintf("Message from %s: ", self.From.PrettyName())) +
		utils.Colorize(cm, utils.ColorWhite, self.Message)
}

func (self EnterEvent) Type() EventType {
	return EnterEventType
}

func (self EnterEvent) ToString(receiver *database.Character) string {
	if receiver.GetRoomId() != self.RoomId {
		return ""
	}

	if receiver.GetId() == self.Character.GetId() {
		return ""
	}

	cm := getColorMode(receiver)

	return utils.Colorize(cm, utils.ColorBlue, self.Character.PrettyName()) +
		utils.Colorize(cm, utils.ColorWhite, " has entered the room")
}

func (self LeaveEvent) Type() EventType {
	return LeaveEventType
}

func (self LeaveEvent) ToString(receiver *database.Character) string {
	if receiver.GetRoomId() != self.RoomId {
		return ""
	}

	if receiver.GetId() == self.Character.GetId() {
		return ""
	}

	cm := getColorMode(receiver)

	return utils.Colorize(cm, utils.ColorBlue, self.Character.PrettyName()) +
		utils.Colorize(cm, utils.ColorWhite, " has left the room")
}

func (self RoomUpdateEvent) Type() EventType {
	return RoomUpdateEventType
}

func (self RoomUpdateEvent) ToString(receiver *database.Character) string {
	if receiver.GetRoomId() != self.Room.GetId() {
		return ""
	}

	cm := getColorMode(receiver)

	return utils.Colorize(cm, utils.ColorWhite, "This room has been modified")
}

func (self LoginEvent) Type() EventType {
	return LoginEventType
}

func (self LoginEvent) ToString(receiver *database.Character) string {
	if receiver.GetId() == self.Character.GetId() {
		return ""
	}

	cm := getColorMode(receiver)

	return utils.Colorize(cm, utils.ColorBlue, self.Character.PrettyName()) +
		utils.Colorize(cm, utils.ColorWhite, " has connected")
}

func (self LogoutEvent) Type() EventType {
	return LogoutEventType
}

func (self LogoutEvent) ToString(receiver *database.Character) string {
	return fmt.Sprintf("%s has disconnected", self.Character.PrettyName())
}

func getColorMode(char *database.Character) utils.ColorMode {
	user := M.GetUser(char.GetUserId())
	return user.GetColorMode()
}

func (self AttackStartEvent) Type() EventType {
	return AttackStartEventType
}

func (self AttackStartEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)

	if receiver == self.Defender {
		return utils.Colorize(cm, utils.ColorRed, fmt.Sprintf("%s is attacking you!", self.Attacker.PrettyName()))
	} else if receiver == self.Attacker {
		return utils.Colorize(cm, utils.ColorRed, fmt.Sprintf("You are attacking %s!", self.Defender.PrettyName()))
	}

	return ""
}

func (self AttackStopEvent) Type() EventType {
	return AttackStopEventType
}

func (self AttackStopEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)

	if receiver == self.Defender {
		return utils.Colorize(cm, utils.ColorGreen, fmt.Sprintf("%s has stopped attacking you", self.Attacker.PrettyName()))
	} else if receiver == self.Attacker {
		return utils.Colorize(cm, utils.ColorGreen, fmt.Sprintf("You stopped attacking %s", self.Defender.PrettyName()))
	}

	return ""
}

// vim: nocindent
