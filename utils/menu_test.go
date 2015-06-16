package utils

import (
	"github.com/Cristofori/kmud/testutils"
	"github.com/Cristofori/kmud/types"
	"strings"
	"testing"
)

func Test_NewMenu(t *testing.T) {
	title := "test menu"
	menu := NewMenu(title)

	if menu.title != title {
		t.Errorf("NewMenu(%s) == %s, want %s", title, menu.title, title)
	}
}

func Test_AddAction(t *testing.T) {
	menu := NewMenu("test")

	key := "KEY"

	menu.AddAction(key, "option")

	testutils.Assert(menu.HasAction(key), t, "Menu didn't have action %s", key)
}

func Test_Exec(t *testing.T) {
	menu := NewMenu("test")

	menu.AddAction("a", "Action 1")
	menu.AddAction("b", "Action 2")

	readWriter := &testutils.TestReadWriter{}

	expected := "a"
	readWriter.ToRead = expected

	choice, _ := menu.Exec(readWriter, types.ColorModeNone)

	testutils.Assert(choice == expected, t, "Expected choice to be %s", expected)
}

func Test_Print(t *testing.T) {
	menu := NewMenu("print test")

	menu.AddAction("a", "Action1")
	menu.AddAction("1", "Action2")
	menu.AddAction("c", "Action3")

	writer := &testutils.TestWriter{}

	menu.Print(writer, types.ColorModeNone)

	testutils.Assert(strings.Contains(writer.Wrote, "[A]ction1"), t, "Didn't have Action1")
	testutils.Assert(strings.Contains(writer.Wrote, "[1]Action2"), t, "Didn't have Action2")
	testutils.Assert(strings.Contains(writer.Wrote, "A[c]tion3"), t, "Didn't have Action3")
}
