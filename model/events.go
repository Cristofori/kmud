package model

import (
	"fmt"
	"time"

	"github.com/Cristofori/kmud/database"
	"github.com/Cristofori/kmud/utils"
)

type EventListener struct {
	Channel chan Event
	Name    string
}

var _listeners []EventListener
var _register chan EventListener
var _unregister chan EventListener
var _broadcast chan Event

func Register(name string) EventListener {
	listener := EventListener{Name: name, Channel: make(chan Event)}
	_register <- listener
	return listener
}

func Unregister(l EventListener) {
	_unregister <- l
}

func Broadcast(event Event) {
	_broadcast <- event
}

func StartEvents() {
	_register = make(chan EventListener)
	_unregister = make(chan EventListener)
	_broadcast = make(chan Event)

	go func() {
		for {
			select {
			case newListener := <-_register:
				_listeners = append(_listeners, newListener)
			case listenerToUnregsiter := <-_unregister:
				for i, listener := range _listeners {
					if listener == listenerToUnregsiter {
						_listeners = append(_listeners[:i], _listeners[i+1:]...)
						break
					}
				}
			case event := <-_broadcast:
				for _, listener := range _listeners {
					go func(l EventListener) {
						l.Channel <- event
					}(listener)
				}
			}
		}
	}()

	go func() {
		throttler := utils.NewThrottler(1 * time.Second)

		for {
			throttler.Sync()
			Broadcast(TickEvent{})
		}
	}()
}

type EventType string

const (
	CreateEventType      EventType = "Create"
	DestroyEventType     EventType = "Destroy"
	BroadcastEventType   EventType = "Broadcast"
	SayEventType         EventType = "Say"
	EmoteEventType       EventType = "Emote"
	TellEventType        EventType = "Tell"
	EnterEventType       EventType = "Enter"
	LeaveEventType       EventType = "Leave"
	RoomUpdateEventType  EventType = "RoomUpdate"
	LoginEventType       EventType = "Login"
	LogoutEventType      EventType = "Logout"
	CombatStartEventType EventType = "CombatStart"
	CombatStopEventType  EventType = "CombatStop"
	CombatEventType      EventType = "Combat"
	TimerEventType       EventType = "Timer"
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

type TickEvent struct {
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
func (self TickEvent) Type() EventType {
	return TimerEventType
}

func (self TickEvent) ToString(receiver *database.Character) string {
	return ""
}

func (self TickEvent) IsFor(receiver *database.PlayerChar) bool {
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
