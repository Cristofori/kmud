package events

import (
	"testing"
	"time"

	tu "github.com/Cristofori/kmud/testutils"
	"github.com/Cristofori/kmud/types"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type EventSuite struct{}

var _ = Suite(&EventSuite{})

func (s *EventSuite) TestEventLoop(c *C) {

	char := types.NewMockPC()

	eventChannel := Register(char)

	message := "hey how are yah"
	Broadcast(TellEvent{char, char, message})

	timeout := tu.Timeout(3 * time.Second)

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
	case <-timeout:
		c.Fatalf("Timed out waiting for tell event")
	}
}
