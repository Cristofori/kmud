package utils

import (
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

    if !menu.HasAction(key) {
        t.Errorf("Menu didn't have action %s", key)
    }
}

