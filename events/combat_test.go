package events

import (
	"fmt"
	"testing"
	"time"

	tu "github.com/Cristofori/kmud/testutils"
	"github.com/Cristofori/kmud/types"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type CombatSuite struct{}

var _ = Suite(&CombatSuite{})

func (s *CombatSuite) SetUpSuite(c *C) {
	fmt.Println("Starting combat loop")
	_combatInterval = 10 * time.Millisecond
	StartEvents()
	StartCombatLoop()
}

func (s *CombatSuite) TearDownSuite(c *C) {
	// StopEvents()
	StopCombatLoop()
}

func (s *CombatSuite) TestCombatLoop(c *C) {

	char1 := types.NewMockPC()
	char2 := types.NewMockPC()
	char1.RoomId = char2.RoomId

	eventChannel1 := Register(char1)
	eventChannel2 := Register(char2)

	StartFight(char1, char2)

	c.Assert(InCombat(char1), Equals, true)
	c.Assert(InCombat(char2), Equals, true)

	verifyEvents := func(channel chan Event) {
		timeout := tu.Timeout(30 * time.Millisecond)

		gotCombatEvent := false
		gotStartEvent := false

	Loop:
		for {
			select {
			case event := <-channel:
				switch event.(type) {
				case TickEvent:
					fmt.Println("Tick event")
				case CombatEvent:
					fmt.Println("Combat event")
					gotCombatEvent = true
				case CombatStartEvent:
					fmt.Println("Combat start event")
					gotStartEvent = true
				default:
					c.FailNow()
				}
			case <-timeout:
				c.Fatalf("Timed out waiting for combat event")
				break Loop
			}

			if gotCombatEvent && gotStartEvent {
				break
			}
		}
	}
	verifyEvents(eventChannel1)
	verifyEvents(eventChannel2)

	StopFight(char1)

	e := <-eventChannel1
	switch e.(type) {
	case CombatStopEvent:
	default:
		c.Fatalf("Didn't get a combat stop event (channel 1)")
	}

	e = <-eventChannel2
	switch e.(type) {
	case CombatStopEvent:
	default:
		c.Fatalf("Didn't get a combat stop event (channel 1)")
	}

	timeout := tu.Timeout(20 * time.Millisecond)

	select {
	case <-eventChannel1:
		c.Fatalf("Shouldn't have gotten any combat events after stopping combat (channel 1)")
	case <-eventChannel2:
		c.Fatalf("Shouldn't have gotten any combat events after stopping combat (channel 2)")
	case <-timeout:
	}

	StartFight(char1, char2)
	<-eventChannel1
	<-eventChannel2

	StopCombatLoop()

	timeout = tu.Timeout(20 * time.Millisecond)

	select {
	case <-eventChannel1:
		c.Fatalf("Shouldn't have gotten any combat events after stopping the combat loop (channel 1)")
	case <-eventChannel2:
		c.Fatalf("Shouldn't have gotten any combat events after stopping the combat loop (channel 2)")
	case <-timeout:
	}
}
