package model

import (
	"container/list"
	"fmt"
	"kmud/database"
	"kmud/utils"
	"sync"
	"time"
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

	_mutex.Lock()
	delete(_listeners, listener)
	_mutex.Unlock()
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

	go func() {
		throttler := utils.NewThrottler(1 * time.Second)

		for {
			throttler.Sync()
			queueEvent(TimerEvent{})
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
	_mutex.Lock()
	for listener := range _listeners {
		if event.IsFor(_listeners[listener]) {
			listener <- event
		}
	}
	_mutex.Unlock()
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
	CombatStartEventType EventType = iota
	CombatStopEventType  EventType = iota
	CombatEventType      EventType = iota
	TimerEventType       EventType = iota
)

type Event interface {
	Type() EventType
	ToString(receiver *database.Character) string
	IsFor(receiver *database.Character) bool
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
	Character  *database.Character
	Room       *database.Room
    SourceRoom *database.Room
}

type LeaveEvent struct {
	Character *database.Character
	Room      *database.Room
    DestRoom  *database.Room
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

type CombatStartEvent struct {
	Attacker *database.Character
	Defender *database.Character
}

type CombatStopEvent struct {
	Attacker *database.Character
	Defender *database.Character
}

type CombatEvent struct {
	Attacker *database.Character
	Defender *database.Character
	Damage   int
}

type TimerEvent struct {
}

func (self BroadcastEvent) Type() EventType {
	return BroadcastEventType
}

func (self BroadcastEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)
	return utils.Colorize(cm, utils.ColorCyan, "Broadcast from "+self.Character.GetName()+": ") +
		utils.Colorize(cm, utils.ColorWhite, self.Message)
}

func (self BroadcastEvent) IsFor(receiver *database.Character) bool {
	return true
}

// Say
func (self SayEvent) Type() EventType {
	return SayEventType
}

func (self SayEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)

	who := ""
	if receiver.GetId() == self.Character.GetId() {
		who = "You say"
	} else {
		who = self.Character.GetName() + " says"
	}

	return utils.Colorize(cm, utils.ColorBlue, who+", ") +
		utils.Colorize(cm, utils.ColorWhite, "\""+self.Message+"\"")
}

func (self SayEvent) IsFor(receiver *database.Character) bool {
	return receiver.GetRoomId() == self.Character.GetRoomId()
}

// Emote
func (self EmoteEvent) Type() EventType {
	return EmoteEventType
}

func (self EmoteEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)
	return utils.Colorize(cm, utils.ColorYellow, self.Character.GetName()+" "+self.Emote)
}

func (self EmoteEvent) IsFor(receiver *database.Character) bool {
	return receiver.GetRoomId() == self.Character.GetRoomId()
}

// Tell
func (self TellEvent) Type() EventType {
	return TellEventType
}

func (self TellEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)
	return utils.Colorize(cm, utils.ColorMagenta, fmt.Sprintf("Message from %s: ", self.From.GetName())) +
		utils.Colorize(cm, utils.ColorWhite, self.Message)
}

func (self TellEvent) IsFor(receiver *database.Character) bool {
	return receiver.GetId() == self.To.GetId()
}

// Enter
func (self EnterEvent) Type() EventType {
	return EnterEventType
}

func (self EnterEvent) ToString(receiver *database.Character) string {
	if receiver.GetId() == self.Character.GetId() {
		return ""
	}

	cm := getColorMode(receiver)

    str := utils.Colorize(cm, utils.ColorBlue, self.Character.GetName()) +
		utils.Colorize(cm, utils.ColorWhite, " has entered the room")

    dir := DirectionBetween(self.Room, self.SourceRoom)
    if dir != database.DirectionNone {
        str = str + utils.Colorize(cm, utils.ColorWhite, " from the " + database.DirectionToString(dir))
    }

    return str
}

func (self EnterEvent) IsFor(receiver *database.Character) bool {
	return receiver.GetRoomId() == self.Room.GetId()
}

// Leave
func (self LeaveEvent) Type() EventType {
	return LeaveEventType
}

func (self LeaveEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)

	str := utils.Colorize(cm, utils.ColorBlue, self.Character.GetName()) +
		utils.Colorize(cm, utils.ColorWhite, " has left the room")

    dir := DirectionBetween(self.Room, self.DestRoom)
    if dir != database.DirectionNone {
        str = str + utils.Colorize(cm, utils.ColorWhite, " to the " + database.DirectionToString(dir))
    }

    return str
}

func (self LeaveEvent) IsFor(receiver *database.Character) bool {
	return receiver.GetRoomId() == self.Room.GetId() &&
		receiver.GetId() != self.Character.GetId()
}

// RoomUpdate
func (self RoomUpdateEvent) Type() EventType {
	return RoomUpdateEventType
}

func (self RoomUpdateEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)
	return utils.Colorize(cm, utils.ColorWhite, "This room has been modified")
}

func (self RoomUpdateEvent) IsFor(receiver *database.Character) bool {
	return receiver.GetRoomId() == self.Room.GetId()
}

// Login
func (self LoginEvent) Type() EventType {
	return LoginEventType
}

func (self LoginEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)
	return utils.Colorize(cm, utils.ColorBlue, self.Character.GetName()) +
		utils.Colorize(cm, utils.ColorWhite, " has connected")
}

func (self LoginEvent) IsFor(receiver *database.Character) bool {
	return receiver.GetId() != self.Character.GetId()
}

// Logout
func (self LogoutEvent) Type() EventType {
	return LogoutEventType
}

func (self LogoutEvent) ToString(receiver *database.Character) string {
	return fmt.Sprintf("%s has disconnected", self.Character.GetName())
}

func (self LogoutEvent) IsFor(receiver *database.Character) bool {
	return true
}

// CombatStart
func (self CombatStartEvent) Type() EventType {
	return CombatStartEventType
}

func (self CombatStartEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)

	if receiver == self.Attacker {
		return utils.Colorize(cm, utils.ColorRed, fmt.Sprintf("You are attacking %s!", self.Defender.GetName()))
	} else if receiver == self.Defender {
		return utils.Colorize(cm, utils.ColorRed, fmt.Sprintf("%s is attacking you!", self.Attacker.GetName()))
	}

	return ""
}

func (self CombatStartEvent) IsFor(receiver *database.Character) bool {
	return receiver == self.Attacker || receiver == self.Defender
}

// CombatStop
func (self CombatStopEvent) Type() EventType {
	return CombatStopEventType
}

func (self CombatStopEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)

	if receiver == self.Attacker {
		return utils.Colorize(cm, utils.ColorGreen, fmt.Sprintf("You stopped attacking %s", self.Defender.GetName()))
	} else if receiver == self.Defender {
		return utils.Colorize(cm, utils.ColorGreen, fmt.Sprintf("%s has stopped attacking you", self.Attacker.GetName()))
	}

	return ""
}

func (self CombatStopEvent) IsFor(receiver *database.Character) bool {
	return receiver == self.Attacker || receiver == self.Defender
}

// Combat
func (self CombatEvent) Type() EventType {
	return CombatEventType
}

func (self CombatEvent) ToString(receiver *database.Character) string {
	cm := getColorMode(receiver)

	if receiver == self.Attacker {
		return utils.Colorize(cm, utils.ColorRed, fmt.Sprintf("You hit %s for %v damage", self.Defender.GetName(), self.Damage))
	} else if receiver == self.Defender {
		return utils.Colorize(cm, utils.ColorRed, fmt.Sprintf("%s hits you for %v damage", self.Attacker.GetName(), self.Damage))
	}

	return ""
}

func (self CombatEvent) IsFor(receiver *database.Character) bool {
	return receiver == self.Attacker || receiver == self.Defender
}

// Timer

func (self TimerEvent) Type() EventType {
	return TimerEventType
}

func (self TimerEvent) ToString(receiver *database.Character) string {
	return ""
}

func (self TimerEvent) IsFor(receiver *database.Character) bool {
	return true
}

// vim: nocindent
