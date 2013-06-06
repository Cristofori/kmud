package utils

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func Test_piecesOf(t *testing.T) {
	var tests = []struct {
		input  string
		output []string
	}{
		{"", []string{}},
		{"foo", []string{"foo"}},
		{"123", []string{"123"}},
		{"foo1bar", []string{"foo", "1", "bar"}},
		{"2bar3foo", []string{"2", "bar", "3", "foo"}},
	}

	for _, test := range tests {
		result := piecesOf(test.input)

		if reflect.DeepEqual(result, test.output) == false {
			t.Errorf("piecesOf(%v) == %v. Want %v", test.input, result, test.output)
		}
	}
}

func Test_sort(t *testing.T) {
	input := SortableStrings{
		"1",
		"a",
		"10",
		"pony",
		"11",
		"3",
		"went",
		"2",
		"to",
		"4",
		"market",
		"y10",
		"y1",
		"y20",
		"y2",
		"z10x10",
		"z10x2",
		"z10x4",
		"z10x3",
		"z10x44",
		"z10x33",
		"z10x1",
	}

	output := SortableStrings{
		"1",
		"2",
		"3",
		"4",
		"10",
		"11",
		"a",
		"market",
		"pony",
		"to",
		"went",
		"y1",
		"y2",
		"y10",
		"y20",
		"z10x1",
		"z10x2",
		"z10x3",
		"z10x4",
		"z10x10",
		"z10x33",
		"z10x44",
	}

	sort.Sort(input)

	if reflect.DeepEqual(input, output) == false {
		t.Errorf("Undeisred sorting result: %v\n", strings.Join(input, ", "))
		t.Errorf("Wanted                  : %v\n", strings.Join(output, ", "))
	}
}

// vim: nocindent
