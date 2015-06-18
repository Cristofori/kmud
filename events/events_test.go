package events

import (
	"testing"
	"time"

	tu "github.com/Cristofori/kmud/testutils"
	"github.com/Cristofori/kmud/types"
)

func Test_EventLoop(t *testing.T) {
	StartEvents()
	StartCombatLoop()

	char := types.NewMockPC()

	eventChannel := Register(char)

	message := "hey how are yah"
	Broadcast(TellEvent{char, char, message})

	timeout := tu.Timeout(3 * time.Second)

	select {
	case event := <-eventChannel:
		tu.Assert(event.Type() == TellEventType, t, "Didn't get a Tell event back")
		tellEvent := event.(TellEvent)
		tu.Assert(tellEvent.Message == message, t, "Didn't get the right message back:", tellEvent.Message, message)
	case <-timeout:
		tu.Assert(false, t, "Timed out waiting for tell event")
	}
}

func Test_CombatLoop(t *testing.T) {
	char1 := types.NewMockPC()
	char2 := types.NewMockPC()
	char1.RoomId = char2.RoomId

	eventChannel1 := Register(char1)

	StartFight(char1, char2)

	verifyEvents := func(channel chan Event) {
		timeout := tu.Timeout(4 * time.Second)
		expectedTypes := make(map[EventType]bool)
		expectedTypes[CombatEventType] = true
		expectedTypes[CombatStartEventType] = true

	Loop:
		for {
			select {
			case event := <-channel:
				if event.Type() != TickEventType {
					tu.Assert(expectedTypes[event.Type()] == true, t, "Unexpected event type:", event.Type())
					delete(expectedTypes, event.Type())
				}
			case <-timeout:
				tu.Assert(false, t, "Timed out waiting for combat event")
				break Loop
			}

			if len(expectedTypes) == 0 {
				break
			}
		}
	}
	verifyEvents(eventChannel1)
}
