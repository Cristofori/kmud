package events

import (
	"testing"
	"time"

	tu "github.com/Cristofori/kmud/testutils"
	"github.com/Cristofori/kmud/types"
)

func Test_EventLoop(t *testing.T) {
	StartEvents()

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
