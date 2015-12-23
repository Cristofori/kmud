package events

import (
	"fmt"
	"time"

	"github.com/Cristofori/kmud/database"
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type EventReceiver interface {
	types.Identifiable
	types.Locateable
}

type SimpleReceiver struct {
}

func (*SimpleReceiver) GetId() types.Id {
	return nil
}

func (*SimpleReceiver) GetRoomId() types.Id {
	return nil
}

type eventListener struct {
	Channel  chan Event
	Receiver EventReceiver
}

var _listeners map[EventReceiver]chan Event

var eventMessages chan interface{}

type register eventListener

type unregister struct {
	Receiver EventReceiver
}

type broadcast struct {
	Event Event
}

func Register(receiver EventReceiver) chan Event {
	listener := eventListener{Receiver: receiver, Channel: make(chan Event)}
	eventMessages <- register(listener)
	return listener.Channel
}

func Unregister(char EventReceiver) {
	eventMessages <- unregister{char}
}

func Broadcast(event Event) {
	eventMessages <- broadcast{event}
}

func init() {
	_listeners = map[EventReceiver]chan Event{}
	eventMessages = make(chan interface{}, 1)

	go func() {
		for message := range eventMessages {
			switch msg := message.(type) {
			case register:
				_listeners[msg.Receiver] = msg.Channel
			case unregister:
				delete(_listeners, msg.Receiver)
			case broadcast:
				for char, channel := range _listeners {
					if msg.Event.IsFor(char) {
						go func(c chan Event) {
							c <- msg.Event
						}(channel)
					}
				}
			default:
				panic("Unhandled event message")
			}
		}

		_listeners = nil
	}()

	go func() {
		throttler := utils.NewThrottler(1 * time.Second)

		for {
			throttler.Sync()
			Broadcast(TickEvent{})
		}
	}()
}

type Event interface {
	ToString(receiver EventReceiver) string
	IsFor(receiver EventReceiver) bool
}

type TickEvent struct{}

type CreateEvent struct {
	Object *database.DbObject
}

type DestroyEvent struct {
	Object *database.DbObject
}

type DeathEvent struct {
	Character types.Character
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

type EnterEvent struct {
	Character types.Character
	RoomId    types.Id
	Direction types.Direction
}

type LeaveEvent struct {
	Character types.Character
	RoomId    types.Id
	Direction types.Direction
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
	Skill    types.Skill
	Power    int
}

type LockEvent struct {
	RoomId types.Id
	Exit   types.Direction
	Locked bool
}

func (self BroadcastEvent) ToString(receiver EventReceiver) string {
	return types.Colorize(types.ColorCyan, "Broadcast from "+self.Character.GetName()+": ") +
		types.Colorize(types.ColorWhite, self.Message)
}

func (self BroadcastEvent) IsFor(receiver EventReceiver) bool {
	return true
}

// Say
func (self SayEvent) ToString(receiver EventReceiver) string {
	who := ""
	if receiver == self.Character {
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
func (self EmoteEvent) ToString(receiver EventReceiver) string {
	return types.Colorize(types.ColorYellow, self.Character.GetName()+" "+self.Emote)
}

func (self EmoteEvent) IsFor(receiver EventReceiver) bool {
	return receiver.GetRoomId() == self.Character.GetRoomId()
}

// Tell
func (self TellEvent) ToString(receiver EventReceiver) string {
	if receiver == self.To {
		return types.Colorize(types.ColorMagenta,
			fmt.Sprintf("Message from %s: %s", self.From.GetName(), types.Colorize(types.ColorWhite, self.Message)))
	} else {
		return types.Colorize(types.ColorMagenta,
			fmt.Sprintf("Message to %s: %s", self.To.GetName(), types.Colorize(types.ColorWhite, self.Message)))
	}
}

func (self TellEvent) IsFor(receiver EventReceiver) bool {
	return receiver == self.To || receiver == self.From
}

// Enter
func (self EnterEvent) ToString(receiver EventReceiver) string {
	message := fmt.Sprintf("%v%s %vhas entered the room", types.ColorBlue, self.Character.GetName(), types.ColorWhite)
	if self.Direction != types.DirectionNone {
		message = fmt.Sprintf("%s from the %s", message, self.Direction.ToString())
	}
	return message
}

func (self EnterEvent) IsFor(receiver EventReceiver) bool {
	return self.RoomId == receiver.GetRoomId() && receiver != self.Character
}

// Leave
func (self LeaveEvent) ToString(receiver EventReceiver) string {
	message := fmt.Sprintf("%v%s %vhas left the room", types.ColorBlue, self.Character.GetName(), types.ColorWhite)
	if self.Direction != types.DirectionNone {
		message = fmt.Sprintf("%s to the %s", message, self.Direction.ToString())
	}
	return message
}

func (self LeaveEvent) IsFor(receiver EventReceiver) bool {
	return self.RoomId == receiver.GetRoomId()
}

// RoomUpdate
func (self RoomUpdateEvent) ToString(receiver EventReceiver) string {
	return types.Colorize(types.ColorWhite, "This room has been modified")
}

func (self RoomUpdateEvent) IsFor(receiver EventReceiver) bool {
	return receiver.GetRoomId() == self.Room.GetId()
}

// Login
func (self LoginEvent) ToString(receiver EventReceiver) string {
	return types.Colorize(types.ColorBlue, self.Character.GetName()) +
		types.Colorize(types.ColorWhite, " has connected")
}

func (self LoginEvent) IsFor(receiver EventReceiver) bool {
	return receiver != self.Character
}

// Logout
func (self LogoutEvent) ToString(receiver EventReceiver) string {
	return fmt.Sprintf("%s has disconnected", self.Character.GetName())
}

func (self LogoutEvent) IsFor(receiver EventReceiver) bool {
	return true
}

// CombatStart
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
func (self CombatEvent) ToString(receiver EventReceiver) string {
	skillMsg := ""
	if self.Skill != nil {
		skillMsg = fmt.Sprintf(" with %s", self.Skill.GetName())
	}

	if receiver == self.Attacker {
		return types.Colorize(types.ColorRed, fmt.Sprintf("You hit %s%s for %v damage", self.Defender.GetName(), skillMsg, self.Power))
	} else if receiver == self.Defender {
		return types.Colorize(types.ColorRed, fmt.Sprintf("%s hits you%s for %v damage", self.Attacker.GetName(), skillMsg, self.Power))
	}

	return ""
}

func (self CombatEvent) IsFor(receiver EventReceiver) bool {
	return receiver == self.Attacker || receiver == self.Defender
}

// Timer
func (self TickEvent) ToString(receiver EventReceiver) string {
	return ""
}

func (self TickEvent) IsFor(receiver EventReceiver) bool {
	return true
}

// Create
func (self CreateEvent) ToString(receiver EventReceiver) string {
	return ""
}

func (self CreateEvent) IsFor(receiver EventReceiver) bool {
	return true
}

// Destroy
func (self DestroyEvent) ToString(receiver EventReceiver) string {
	return ""
}

func (self DestroyEvent) IsFor(receiver EventReceiver) bool {
	return true
}

// Death
func (self DeathEvent) IsFor(receiver EventReceiver) bool {
	return receiver == self.Character ||
		receiver.GetRoomId() == self.Character.GetRoomId()
}

func (self DeathEvent) ToString(receiver EventReceiver) string {
	if receiver == self.Character {
		return types.Colorize(types.ColorRed, ">> You have died")
	}

	return types.Colorize(types.ColorRed, fmt.Sprintf(">> %s has died", self.Character.GetName()))
}

// Lock
func (self LockEvent) IsFor(receiver EventReceiver) bool {
	return receiver.GetRoomId() == self.RoomId
}

func (self LockEvent) ToString(receiver EventReceiver) string {
	status := "unlocked"
	if self.Locked {
		status = "locked"
	}

	return types.Colorize(types.ColorBlue,
		fmt.Sprintf("The exit to the %s has been %s", self.Exit.ToString(),
			types.Colorize(types.ColorWhite, status)))
}
