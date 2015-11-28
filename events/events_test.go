package events

import (
	"testing"
	"time"

	"github.com/Cristofori/kmud/testutils"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type EventSuite struct{}

var _ = Suite(&EventSuite{})

func (s *EventSuite) TestEventLoop(c *C) {

	char := testutils.NewMockPC()

	eventChannel := Register(char)

	message := "hey how are yah"
	Broadcast(TellEvent{char, char, message})

	select {
	case event := <-eventChannel:
		gotTellEvent := false

		switch e := event.(type) {
		case TellEvent:
			gotTellEvent = true
			c.Assert(e.Message, Equals, message)

		}

		if gotTellEvent == false {
			c.Fatalf("Didn't get a Tell event back")
		}
	case <-time.After(3 * time.Second):
		c.Fatalf("Timed out waiting for tell event")
	}
}
