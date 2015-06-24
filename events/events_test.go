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
		gotTellEvent := false

		switch e := event.(type) {
		case TellEvent:
			gotTellEvent = true
			tu.Assert(e.Message == message, t, "Didn't get the right message back:", e.Message, message)

		}
		tu.Assert(gotTellEvent, t, "Didn't get a Tell event back")
	case <-timeout:
		tu.Assert(false, t, "Timed out waiting for tell event")
	}
}
