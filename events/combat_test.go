package events

import (
	"testing"
	"time"

	tu "github.com/Cristofori/kmud/testutils"
	"github.com/Cristofori/kmud/types"
)

func Test_CombatLoop(t *testing.T) {
	_combatInterval = 10 * time.Millisecond

	StartEvents()
	StartCombatLoop()

	char1 := types.NewMockPC()
	char2 := types.NewMockPC()
	char1.RoomId = char2.RoomId

	eventChannel1 := Register(char1)
	eventChannel2 := Register(char2)

	StartFight(char1, char2)

	tu.Assert(InCombat(char1) == true, t, "char1 did not get flagged as in combat")
	tu.Assert(InCombat(char2) == true, t, "char2 did not get flagged as in combat")

	verifyEvents := func(channel chan Event) {
		timeout := tu.Timeout(30 * time.Millisecond)
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
	verifyEvents(eventChannel2)

	StopCombatLoop()

	timeout := tu.Timeout(20 * time.Millisecond)

	select {
	case <-eventChannel1:
		tu.Assert(false, t, "Shouldn't have gotten any combat events after stopping the combat loop (channel 1)")
	case <-eventChannel2:
		tu.Assert(false, t, "Shouldn't have gotten any combat events after stopping the combat loop (channel 2)")
	case <-timeout:
	}
}
