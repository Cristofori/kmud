package model

import (
	"container/list"
	"fmt"
	"kmud/database"
	"kmud/utils"
	"sync"
	"time"
)

var _listeners []chan Event
var _mutex sync.Mutex
var _eventQueueChannel chan Event

func Login(character *database.PlayerChar) {
	character.SetOnline(true)
	queueEvent(LoginEvent{character})
}

func Logout(character *database.PlayerChar) {
	character.SetOnline(false)
	queueEvent(LogoutEvent{character})
}

func Register() chan Event {
	listener := make(chan Event, 100)

	_mutex.Lock()
	_listeners = append(_listeners, listener)
	_mutex.Unlock()

	return listener
}

func Unregister(listenerToUnregsiter chan Event) {
	_mutex.Lock()
	for i, listener := range _listeners {
		if listener == listenerToUnregsiter {
			_listeners = append(_listeners[:i], _listeners[i+1:]...)
			break
		}
	}
	_mutex.Unlock()
}

func eventLoop() {
	_eventQueueChannel = make(chan Event, 100)

	var m sync.Mutex
	cond := sync.NewCond(&m)

	eventQueue := list.New()

	go func() {
		for {
			event := <-_eventQueueChannel

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
	_eventQueueChannel <- event
}

func broadcast(event Event) {
	_mutex.Lock()
	for _, listener := range _listeners {
		listener <- event
	}
	_mutex.Unlock()
}

type EventType int

const (
	CreateEventType      EventType = iota
	DestroyEventType     EventType = iota
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
	IsFor(receiver *database.PlayerChar) bool
}

type CreateEvent struct {
	Object *database.DbObject
}

type DestroyEvent struct {
	Object *database.DbObject
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
	Character *database.PlayerChar
}

type LogoutEvent struct {
	Character *database.PlayerChar
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
	return utils.Colorize(utils.ColorCyan, "Broadcast from "+self.Character.GetName()+": ") +
		utils.Colorize(utils.ColorWhite, self.Message)
}

func (self BroadcastEvent) IsFor(receiver *database.PlayerChar) bool {
	return true
}

// Say
func (self SayEvent) Type() EventType {
	return SayEventType
}

func (self SayEvent) ToString(receiver *database.Character) string {
	who := ""
	if receiver.GetId() == self.Character.GetId() {
		who = "You say"
	} else {
		who = self.Character.GetName() + " says"
	}

	return utils.Colorize(utils.ColorBlue, who+", ") +
		utils.Colorize(utils.ColorWhite, "\""+self.Message+"\"")
}

func (self SayEvent) IsFor(receiver *database.PlayerChar) bool {
	return receiver.GetRoomId() == self.Character.GetRoomId()
}

// Emote
func (self EmoteEvent) Type() EventType {
	return EmoteEventType
}

func (self EmoteEvent) ToString(receiver *database.Character) string {
	return utils.Colorize(utils.ColorYellow, self.Character.GetName()+" "+self.Emote)
}

func (self EmoteEvent) IsFor(receiver *database.PlayerChar) bool {
	return receiver.GetRoomId() == self.Character.GetRoomId()
}

// Tell
func (self TellEvent) Type() EventType {
	return TellEventType
}

func (self TellEvent) ToString(receiver *database.Character) string {
	return utils.Colorize(utils.ColorMagenta, fmt.Sprintf("Message from %s: ", self.From.GetName())) +
		utils.Colorize(utils.ColorWhite, self.Message)
}

func (self TellEvent) IsFor(receiver *database.PlayerChar) bool {
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

	str := fmt.Sprintf("%v%s %vhas entered the room", utils.ColorBlue, self.Character.GetName(), utils.ColorWhite)

	dir := DirectionBetween(self.Room, self.SourceRoom)
	if dir != database.DirectionNone {
		str = str + " from the " + database.DirectionToString(dir)
	}

	return str
}

func (self EnterEvent) IsFor(receiver *database.PlayerChar) bool {
	return receiver.GetRoomId() == self.Room.GetId()
}

// Leave
func (self LeaveEvent) Type() EventType {
	return LeaveEventType
}

func (self LeaveEvent) ToString(receiver *database.Character) string {
	str := fmt.Sprintf("%v%s %vhas left the room", utils.ColorBlue, self.Character.GetName(), utils.ColorWhite)

	dir := DirectionBetween(self.Room, self.DestRoom)
	if dir != database.DirectionNone {
		str = str + " to the " + database.DirectionToString(dir)
	}

	return str
}

func (self LeaveEvent) IsFor(receiver *database.PlayerChar) bool {
	return receiver.GetRoomId() == self.Room.GetId() &&
		receiver.GetId() != self.Character.GetId()
}

// RoomUpdate
func (self RoomUpdateEvent) Type() EventType {
	return RoomUpdateEventType
}

func (self RoomUpdateEvent) ToString(receiver *database.Character) string {
	return utils.Colorize(utils.ColorWhite, "This room has been modified")
}

func (self RoomUpdateEvent) IsFor(receiver *database.PlayerChar) bool {
	return receiver.GetRoomId() == self.Room.GetId()
}

// Login
func (self LoginEvent) Type() EventType {
	return LoginEventType
}

func (self LoginEvent) ToString(receiver *database.Character) string {
	return utils.Colorize(utils.ColorBlue, self.Character.GetName()) +
		utils.Colorize(utils.ColorWhite, " has connected")
}

func (self LoginEvent) IsFor(receiver *database.PlayerChar) bool {
	return receiver.GetId() != self.Character.GetId()
}

// Logout
func (self LogoutEvent) Type() EventType {
	return LogoutEventType
}

func (self LogoutEvent) ToString(receiver *database.Character) string {
	return fmt.Sprintf("%s has disconnected", self.Character.GetName())
}

func (self LogoutEvent) IsFor(receiver *database.PlayerChar) bool {
	return true
}

// CombatStart
func (self CombatStartEvent) Type() EventType {
	return CombatStartEventType
}

func (self CombatStartEvent) ToString(receiver *database.Character) string {
	if receiver == self.Attacker {
		return utils.Colorize(utils.ColorRed, fmt.Sprintf("You are attacking %s!", self.Defender.GetName()))
	} else if receiver == self.Defender {
		return utils.Colorize(utils.ColorRed, fmt.Sprintf("%s is attacking you!", self.Attacker.GetName()))
	}

	return ""
}

func (self CombatStartEvent) IsFor(receiver *database.PlayerChar) bool {
	return &receiver.Character == self.Attacker || &receiver.Character == self.Defender
}

// CombatStop
func (self CombatStopEvent) Type() EventType {
	return CombatStopEventType
}

func (self CombatStopEvent) ToString(receiver *database.Character) string {
	if receiver == self.Attacker {
		return utils.Colorize(utils.ColorGreen, fmt.Sprintf("You stopped attacking %s", self.Defender.GetName()))
	} else if receiver == self.Defender {
		return utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s has stopped attacking you", self.Attacker.GetName()))
	}

	return ""
}

func (self CombatStopEvent) IsFor(receiver *database.PlayerChar) bool {
	return &receiver.Character == self.Attacker || &receiver.Character == self.Defender
}

// Combat
func (self CombatEvent) Type() EventType {
	return CombatEventType
}

func (self CombatEvent) ToString(receiver *database.Character) string {
	if receiver == self.Attacker {
		return utils.Colorize(utils.ColorRed, fmt.Sprintf("You hit %s for %v damage", self.Defender.GetName(), self.Damage))
	} else if receiver == self.Defender {
		return utils.Colorize(utils.ColorRed, fmt.Sprintf("%s hits you for %v damage", self.Attacker.GetName(), self.Damage))
	}

	return ""
}

func (self CombatEvent) IsFor(receiver *database.PlayerChar) bool {
	return &receiver.Character == self.Attacker || &receiver.Character == self.Defender
}

// Timer
func (self TimerEvent) Type() EventType {
	return TimerEventType
}

func (self TimerEvent) ToString(receiver *database.Character) string {
	return ""
}

func (self TimerEvent) IsFor(receiver *database.PlayerChar) bool {
	return true
}

// Create
func (self CreateEvent) Type() EventType {
	return CreateEventType
}

func (self CreateEvent) ToString(receiver *database.Character) string {
	return ""
}

func (self CreateEvent) IsFor(receiver *database.PlayerChar) bool {
	return true
}

// Destroy
func (self DestroyEvent) Type() EventType {
	return DestroyEventType
}

func (self DestroyEvent) ToString(receiver *database.Character) string {
	return ""
}

func (self DestroyEvent) IsFor(receiver *database.PlayerChar) bool {
	return true
}

// vim: nocindent
