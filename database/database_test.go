package database

import (
	"labix.org/v2/mgo/bson"
	"testing"
)

func compareSlices(slice1 []bson.ObjectId, slice2 []bson.ObjectId) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i, id := range slice1 {
		if id != slice2[i] {
			return false
		}
	}

	return true
}

func sliceContains(id bson.ObjectId, slice []bson.ObjectId) bool {
	for _, sliceId := range slice {
		if sliceId == id {
			return true
		}
	}

	return false
}

func Test_RemoveId(t *testing.T) {
	var ids = []bson.ObjectId{
		"0",
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
	}

	length := len(ids)

	removeElement := func(id bson.ObjectId) {
		ids = removeId(id, ids)

		if len(ids) != length-1 {
			t.Errorf("Failed to remove exactly one element from slice: %s, %s", id, ids)
		}

		if sliceContains(id, ids) {
			t.Errorf("Failed to remove element from slice: %s, %s", id, ids)
		}

		length--
	}

	removeElement("0")
	removeElement("9")
	removeElement("1")
	removeElement("8")
	removeElement("2")
	removeElement("3")
	removeElement("6")
	removeElement("7")
	removeElement("4")
	removeElement("5")
}

// vim: nocindent
