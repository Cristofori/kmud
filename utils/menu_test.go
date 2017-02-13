package utils

import (
	"fmt"
	"testing"

	"github.com/Cristofori/kmud/testutils"
	"github.com/Cristofori/kmud/types"
)

func Test_Menu(t *testing.T) {
	var comm testutils.TestCommunicable
	comm.ToRead = "a"

	called := false
	handled := false
	onExitCalled := false

	ExecMenu("Menu", &comm, func(menu *Menu) {
		called = true

		menu.AddAction("a", "Apple", func() {
			handled = true
			menu.Exit()
		})

		menu.OnExit(func() {
			onExitCalled = true
		})
	})

	testutils.Assert(called == true, t, "Failed to exec menu")
	testutils.Assert(handled == true, t, "Failed to handle menu action")
	testutils.Assert(onExitCalled == true, t, "Failed to call the OnExit handler")
}

func Test_Search(t *testing.T) {
	var comm1 testutils.TestCommunicable
	var comm2 testutils.TestCommunicable
	var menu1 Menu
	var menu2 Menu

	title := "Menu"
	menu1.SetTitle(title)
	menu2.SetTitle(title)

	menu1.AddActionI(0, "Action One", func() {
		menu1.Exit()
	})
	menu2.AddActionI(0, "Action One", func() {
		menu2.Exit()
	})

	menu2.AddActionI(1, "Action Two", func() {
		menu1.Exit()
	})

	filter := "one"
	menu1.Print(&comm1, 0, filter)
	menu2.Print(&comm2, 0, filter)

	testutils.Assert(comm1.Wrote == comm2.Wrote, t, fmt.Sprintf("Failed to correctly filter menu, got: \n%s, expected: \n%s",
		types.StripColors(comm2.Wrote), types.StripColors(comm1.Wrote)))
}
