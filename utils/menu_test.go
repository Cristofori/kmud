package utils

import (
	"testing"

	"github.com/Cristofori/kmud/testutils"
)

func Test_Menu(t *testing.T) {
	var Comm testutils.TestCommunicable
	Comm.ToRead = "a"

	called := false
	handled := false
	ExecMenu("Menu", &Comm, func(menu *Menu) {
		called = true

		menu.AddAction("a", "Apple", func() bool {
			handled = true
			return false
		})
	})

	testutils.Assert(called == true, t, "Failed to exec menu")
	testutils.Assert(handled == true, t, "Failed to handle menu action")
}
