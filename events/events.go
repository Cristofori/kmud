package events

import (
	"fmt"
	"time"

	"github.com/Cristofori/kmud/database"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type eventListener struct {
	Channel   chan Event
	Character types.Character
}

var _listeners map[types.Character]chan Event

var _register chan eventListener
var _unregister chan types.Character
var _broadcast chan Event

func Register(receiver types.Character) chan Event {
	listener := eventListener{Character: receiver, Channel: make(chan Event, 20)}
	_register <- listener
	return listener.Channel
}

func Unregister(char types.Character) {
	_unregister <- char
}

func Broadcast(event Event) {
	_broadcast <- event
}

func StartEvents() {
	if _listeners != nil {
		return
	}

	_listeners = map[types.Character]chan Event{}

	_register = make(chan eventListener)
	_unregister = make(chan types.Character)
	_broadcast = make(chan Event)

	go func() {
		for {
			select {
			case l := <-_register:
				_listeners[l.Character] = l.Channel
			case char := <-_unregister:
				delete(_listeners, char)
			case event := <-_broadcast:
				for char, channel := range _listeners {

					if event.IsFor(char) {
						if len(channel) == cap(channel) {
							// TODO - Kill the session rather than the whole server
							panic("Buffer full!" + char.GetName() + " " + string(event.Type()))
						}

						channel <- event
					}
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

	StartCombatLoop()
}

type EventType string

const (
	CreateEventType      EventType = "Create"
	DestroyEventType     EventType = "Destroy"
	BroadcastEventType   EventType = "Broadcast"
	SayEventType         EventType = "Say"
	EmoteEventType       EventType = "Emote"
	TellEventType        EventType = "Tell"
	MoveEventType        EventType = "Move"
	RoomUpdateEventType  EventType = "RoomUpdate"
	LoginEventType       EventType = "Login"
	LogoutEventType      EventType = "Logout"
	CombatStartEventType EventType = "CombatStart"
	CombatStopEventType  EventType = "CombatStop"
	CombatEventType      EventType = "Combat"
	TickEventType        EventType = "Tick"
)

type EventReceiver interface {
	types.Identifiable
	types.Locateable
}

type Event interface {
	Type() EventType
	ToString(receiver EventReceiver) string
	IsFor(receiver EventReceiver) bool
}

type CreateEvent struct {
	Object *database.DbObject
}

type DestroyEvent struct {
	Object *database.DbObject
}

type BroadcastEvent struct {
	Character types.Character
	Message   string
}

type SayEvent struct {
	Character types.Character
	Message   string
}

type EmoteEvent struct {
	Character types.Character
	Emote     string
}

type TellEvent struct {
	From    types.Character
	To      types.Character
	Message string
}

type MoveEvent struct {
	Character types.Character
	Room      types.Room
	Message   string
}

type RoomUpdateEvent struct {
	Room *database.Room
}

type LoginEvent struct {
	Character types.Character
}

type LogoutEvent struct {
	Character types.Character
}

type CombatStartEvent struct {
	Attacker types.Character
	Defender types.Character
}

type CombatStopEvent struct {
	Attacker types.Character
	Defender types.Character
}

type CombatEvent struct {
	Attacker types.Character
	Defender types.Character
	Damage   int
}

type TickEvent struct {
}

func (self BroadcastEvent) Type() EventType {
	return BroadcastEventType
}

func (self BroadcastEvent) ToString(receiver EventReceiver) string {
	return types.Colorize(types.ColorCyan, "Broadcast from "+self.Character.GetName()+": ") +
		types.Colorize(types.ColorWhite, self.Message)
}

func (self BroadcastEvent) IsFor(receiver EventReceiver) bool {
	return true
}

// Say
func (self SayEvent) Type() EventType {
	return SayEventType
}

func (self SayEvent) ToString(receiver EventReceiver) string {
	who := ""
	if receiver.GetId() == self.Character.GetId() {
		who = "You say"
	} else {
		who = self.Character.GetName() + " says"
	}

	return types.Colorize(types.ColorBlue, who+", ") +
		types.Colorize(types.ColorWhite, "\""+self.Message+"\"")
}

func (self SayEvent) IsFor(receiver EventReceiver) bool {
	return receiver.GetRoomId() == self.Character.GetRoomId()
}

// Emote
func (self EmoteEvent) Type() EventType {
	return EmoteEventType
}

func (self EmoteEvent) ToString(receiver EventReceiver) string {
	return types.Colorize(types.ColorYellow, self.Character.GetName()+" "+self.Emote)
}

func (self EmoteEvent) IsFor(receiver EventReceiver) bool {
	return receiver.GetRoomId() == self.Character.GetRoomId()
}

// Tell
func (self TellEvent) Type() EventType {
	return TellEventType
}

func (self TellEvent) ToString(receiver EventReceiver) string {
	return types.Colorize(types.ColorMagenta, fmt.Sprintf("Message from %s: ", self.From.GetName())) +
		types.Colorize(types.ColorWhite, self.Message)
}

func (self TellEvent) IsFor(receiver EventReceiver) bool {
	return receiver.GetId() == self.To.GetId()
}

// Leave
func (self MoveEvent) Type() EventType {
	return MoveEventType
}

func (self MoveEvent) ToString(receiver EventReceiver) string {
	return self.Message
}

func (self MoveEvent) IsFor(receiver EventReceiver) bool {
	return receiver.GetRoomId() == self.Room.GetId() &&
		receiver.GetId() != self.Character.GetId()
}

// RoomUpdate
func (self RoomUpdateEvent) Type() EventType {
	return RoomUpdateEventType
}

func (self RoomUpdateEvent) ToString(receiver EventReceiver) string {
	return types.Colorize(types.ColorWhite, "This room has been modified")
}

func (self RoomUpdateEvent) IsFor(receiver EventReceiver) bool {
	return receiver.GetRoomId() == self.Room.GetId()
}

// Login
func (self LoginEvent) Type() EventType {
	return LoginEventType
}

func (self LoginEvent) ToString(receiver EventReceiver) string {
	return types.Colorize(types.ColorBlue, self.Character.GetName()) +
		types.Colorize(types.ColorWhite, " has connected")
}

func (self LoginEvent) IsFor(receiver EventReceiver) bool {
	return receiver.GetId() != self.Character.GetId()
}

// Logout
func (self LogoutEvent) Type() EventType {
	return LogoutEventType
}

func (self LogoutEvent) ToString(receiver EventReceiver) string {
	return fmt.Sprintf("%s has disconnected", self.Character.GetName())
}

func (self LogoutEvent) IsFor(receiver EventReceiver) bool {
	return true
}

// CombatStart
func (self CombatStartEvent) Type() EventType {
	return CombatStartEventType
}

func (self CombatStartEvent) ToString(receiver EventReceiver) string {
	if receiver == self.Attacker {
		return types.Colorize(types.ColorRed, fmt.Sprintf("You are attacking %s!", self.Defender.GetName()))
	} else if receiver == self.Defender {
		return types.Colorize(types.ColorRed, fmt.Sprintf("%s is attacking you!", self.Attacker.GetName()))
	}

	return ""
}

func (self CombatStartEvent) IsFor(receiver EventReceiver) bool {
	return receiver == self.Attacker || receiver == self.Defender
}

// CombatStop
func (self CombatStopEvent) Type() EventType {
	return CombatStopEventType
}

func (self CombatStopEvent) ToString(receiver EventReceiver) string {
	if receiver == self.Attacker {
		return types.Colorize(types.ColorGreen, fmt.Sprintf("You stopped attacking %s", self.Defender.GetName()))
	} else if receiver == self.Defender {
		return types.Colorize(types.ColorGreen, fmt.Sprintf("%s has stopped attacking you", self.Attacker.GetName()))
	}

	return ""
}

func (self CombatStopEvent) IsFor(receiver EventReceiver) bool {
	return receiver == self.Attacker || receiver == self.Defender
}

// Combat
func (self CombatEvent) Type() EventType {
	return CombatEventType
}

func (self CombatEvent) ToString(receiver EventReceiver) string {
	if receiver == self.Attacker {
		return types.Colorize(types.ColorRed, fmt.Sprintf("You hit %s for %v damage", self.Defender.GetName(), self.Damage))
	} else if receiver == self.Defender {
		return types.Colorize(types.ColorRed, fmt.Sprintf("%s hits you for %v damage", self.Attacker.GetName(), self.Damage))
	}

	return ""
}

func (self CombatEvent) IsFor(receiver EventReceiver) bool {
	return receiver == self.Attacker || receiver == self.Defender
}

// Timer
func (self TickEvent) Type() EventType {
	return TickEventType
}

func (self TickEvent) ToString(receiver EventReceiver) string {
	return ""
}

func (self TickEvent) IsFor(receiver EventReceiver) bool {
	return true
}

// Create
func (self CreateEvent) Type() EventType {
	return CreateEventType
}

func (self CreateEvent) ToString(receiver EventReceiver) string {
	return ""
}

func (self CreateEvent) IsFor(receiver EventReceiver) bool {
	return true
}

// Destroy
func (self DestroyEvent) Type() EventType {
	return DestroyEventType
}

func (self DestroyEvent) ToString(receiver EventReceiver) string {
	return ""
}

func (self DestroyEvent) IsFor(receiver EventReceiver) bool {
	return true
}

// vim: nocindent
