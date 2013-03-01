package utils

import (
	"errors"
	"strings"
	"testing"
)

//import "fmt"

var wrote string

type testWriter struct {
}

func (self testWriter) Write(p []byte) (n int, err error) {
	wrote = string(p)
	return len(p), nil
}

type testReader struct {
	toRead string
}

func (self testReader) Read(p []byte) (n int, err error) {
	for i := 0; i < len(self.toRead); i++ {
		p[i] = self.toRead[i]
	}

	p[len(self.toRead)] = '\n'

	return len(self.toRead) + 1, nil
}

type testReadWriter struct {
	testReader
	testWriter
}

func Test_WriteLine(t *testing.T) {
	var writer testWriter

	line := "This is a line"
	want := line + "\r\n"

	WriteLine(writer, line)

	if wrote != want {
		t.Errorf("WriteLine(%q) == %q, want %q", line, wrote, want)
	}
}

func Test_Simplify(t *testing.T) {
	var tests = []struct {
		s, want string
	}{
		{" Test1", "test1"},
		{"TesT2 ", "test2"},
		{" tESt3 ", "test3"},
		{"\tTeSt4\t", "test4"},
		{"    teSt5     \n", "test5"},
		{"   \t \t TEST6 \n \t", "test6"},
	}

	for _, c := range tests {
		got := Simplify(c.s)
		if got != c.want {
			t.Errorf("Simplify(%q) == %q, want %q", c.s, got, c.want)
		}
	}
}

func Test_GetRawUserInput(t *testing.T) {
	var readWriter testReadWriter

	var tests = []struct {
		input, output string
	}{
		{"Test1", "Test1"},
		{"TEST2", "TEST2"},
		{" test3 ", " test3 "},
		{"\tTeSt4\t", "\tTeSt4\t"},
		{"x", ""},
		{"X", ""},
	}

	for _, test := range tests {
		readWriter.toRead = test.input
		line := GetRawUserInput(readWriter, ">")
		if line != test.output {
			t.Errorf("GetRawUserInput(%q) == %q, want %q", test.input, line, test.output)
		}
	}
}

func Test_GetUserInput(t *testing.T) {
	var readWriter testReadWriter

	var tests = []struct {
		input, output string
	}{
		{"Test1", "test1"},
		{"TEST2", "test2"},
		{"\tTeSt3\t", "test3"},
		{"    teSt4     \n", "test4"},
		{"   \t \t TEST5 \n \t", "test5"},
		{"x", ""},
		{"X", ""},
	}

	for _, test := range tests {
		readWriter.toRead = test.input
		line := GetUserInput(readWriter, ">")
		if line != test.output {
			t.Errorf("GetUserInput(%q) == %q, want %q", test.input, line, test.output)
		}
	}
}

func Test_HandleError(t *testing.T) {
	// TODO try using recover
	HandleError(nil)
}

func Test_FormatName(t *testing.T) {
	var tests = []struct {
		input  string
		output string
	}{
		{"", ""},
		{"joe", "Joe"},
		{"ASDF", "Asdf"},
		{"Bob", "Bob"},
		{"aBcDeFg", "Abcdefg"},
	}

	for _, test := range tests {
		result := FormatName(test.input)
		if result != test.output {
			t.Errorf("FormatName(%s) == %s, want %s", test.input, result, test.output)
		}
	}
}

func Test_ValidateName(t *testing.T) {
	var tests = []struct {
		input  string
		output bool
	}{
		{"t", false},
		{"te", false},
		{"tes", true},
		{"te1", false},
		{"test", true},
		{"testing", true},
		{"*(!(@#*$", false},
		{"Abc1abc", false},
		{"123456789012", false},
		{"aslsidjfljll", true},
		{"1slsidjfljll", false},
		{"aslsidjfljl3", false},
	}

	for _, test := range tests {
		result := (ValidateName(test.input) == nil)
		if result != test.output {
			t.Errorf("ValidateName(%q) == %q, want %q", test.input, result, test.output)
		}
	}
}

func Test_BestMatch(t *testing.T) {
	searchList := []string{"", "Foo", "Bar", "Joe", "Bob", "Abcdef", "Abc", "QrStUv"}

	var tests = []struct {
		input  string
		output int
	}{
		{"f", 1},
		{"B", -2},
		{"alseifjlfji", -1},
		{"AB", -2},
		{"aBc", 6},
		{"AbCd", 5},
		{"q", 7},
		{"jo", 3},
	}

	for _, test := range tests {
		result := BestMatch(test.input, searchList)
		if result != test.output {
			t.Errorf("BestMatch(%v) == %v, want %v", test.input, result, test.output)
		}
	}
}

func compareStringLists(l1 []string, l2 []string) bool {
	if len(l1) != len(l2) {
		return false
	}

	for i, str := range l1 {
		if str != l2[i] {
			return false
		}
	}

	return true
}

func Test_Argify(t *testing.T) {
	var tests = []struct {
		input   string
		output1 string
		output2 []string
	}{
		{"", "", []string{}},
		{"test", "test", []string{}},
		{"test two", "test", []string{"two"}},
		{"test one two", "test", []string{"one", "two"}},
		{"this is a somewhat longer test that should also work",
			"this", []string{"is", "a", "somewhat", "longer", "test", "that", "should", "also", "work"}},
	}

	for _, test := range tests {
		result1, result2 := Argify(test.input)

		if result1 != test.output1 || compareStringLists(result2, test.output2) == false {
			t.Errorf("Argify(%v) == %v, %v. Want %v, %v", test.input, result1, result2, test.output1, test.output2)
		}
	}
}

func Test_TrimUpperRows(t *testing.T) {
	emptyRow1 := "                                                                           "
	emptyRow2 := " "
	emptyRow3 := ""

	nonEmptyRow1 := "A"
	nonEmptyRow2 := "A                         "
	nonEmptyRow3 := "A                        B"
	nonEmptyRow4 := "                         B"
	nonEmptyRow5 := "            C             "

	var tests = []struct {
		input  []string
		output []string
	}{
		{[]string{}, []string{}},
		{[]string{nonEmptyRow1}, []string{nonEmptyRow1}},
		{[]string{emptyRow1}, []string{}},
		{[]string{emptyRow1, emptyRow2, emptyRow3}, []string{}},
		{[]string{nonEmptyRow4, nonEmptyRow3, nonEmptyRow2}, []string{nonEmptyRow4, nonEmptyRow3, nonEmptyRow2}},
		{[]string{emptyRow3, nonEmptyRow5}, []string{nonEmptyRow5}},
		{[]string{nonEmptyRow1, nonEmptyRow2, emptyRow1, emptyRow2, emptyRow3}, []string{nonEmptyRow1, nonEmptyRow2, emptyRow1, emptyRow2, emptyRow3}},
	}

	for _, test := range tests {
		result := TrimUpperRows(test.input)

		if compareStringLists(result, test.output) == false {
			t.Errorf("TrimUpperRows(\n%v) == \n%v,\nWanted: \n%v", test.input, result, test.output)
		}
	}
}

func Test_TrimLowerRows(t *testing.T) {
	emptyRow1 := "                                                                           "
	emptyRow2 := " "
	emptyRow3 := ""

	nonEmptyRow1 := "A"
	nonEmptyRow2 := "A                         "
	nonEmptyRow3 := "A                        B"
	nonEmptyRow4 := "                         B"
	nonEmptyRow5 := "            C             "

	var tests = []struct {
		input  []string
		output []string
	}{
		{[]string{}, []string{}},
		{[]string{nonEmptyRow1}, []string{nonEmptyRow1}},
		{[]string{emptyRow1}, []string{}},
		{[]string{emptyRow1, emptyRow2, emptyRow3}, []string{}},
		{[]string{nonEmptyRow4, nonEmptyRow3, nonEmptyRow2}, []string{nonEmptyRow4, nonEmptyRow3, nonEmptyRow2}},
		{[]string{emptyRow3, nonEmptyRow5}, []string{emptyRow3, nonEmptyRow5}},
		{[]string{nonEmptyRow1, nonEmptyRow2, emptyRow1, emptyRow2, emptyRow3}, []string{nonEmptyRow1, nonEmptyRow2}},
	}

	for _, test := range tests {
		result := TrimLowerRows(test.input)

		if compareStringLists(result, test.output) == false {
			t.Errorf("TrimLowerRows(\n%v) == \n%v,\nWanted: \n%v", test.input, result, test.output)
		}
	}
}

func Test_TrimEmptyRows(t *testing.T) {
	emptyRow1 := "                                                                           "
	emptyRow2 := " "
	emptyRow3 := ""

	nonEmptyRow1 := "A"
	nonEmptyRow2 := "A                         "
	nonEmptyRow3 := "A                        B"
	nonEmptyRow4 := "                         B"
	nonEmptyRow5 := "            C             "

	NL := "\r\n"

	var tests = []struct {
		input  string
		output string
	}{
		{"", ""},
		{strings.Join([]string{nonEmptyRow1}, NL),
			strings.Join([]string{nonEmptyRow1}, NL)},
		{strings.Join([]string{emptyRow1}, NL),
			strings.Join([]string{}, NL)},
		{strings.Join([]string{emptyRow1, emptyRow2, emptyRow3}, NL),
			strings.Join([]string{}, NL)},
		{strings.Join([]string{nonEmptyRow4, nonEmptyRow3, nonEmptyRow2}, NL),
			strings.Join([]string{nonEmptyRow4, nonEmptyRow3, nonEmptyRow2}, NL)},
		{strings.Join([]string{emptyRow3, nonEmptyRow5}, NL),
			strings.Join([]string{nonEmptyRow5}, NL)},
		{strings.Join([]string{nonEmptyRow1, nonEmptyRow2, emptyRow1, emptyRow2, emptyRow3}, NL),
			strings.Join([]string{nonEmptyRow1, nonEmptyRow2}, NL)},
		{strings.Join([]string{emptyRow1, emptyRow2, emptyRow3, nonEmptyRow1, nonEmptyRow2, emptyRow1, emptyRow2, emptyRow3}, NL),
			strings.Join([]string{nonEmptyRow1, nonEmptyRow2}, NL)},
	}

	for i, test := range tests {
		result := TrimEmptyRows(test.input)

		if result != test.output {
			t.Errorf("%v: TrimEmptyRows(\n%v) == \n%v,\nWanted: \n%v", i, test.input, result, test.output)
		}
	}
}

func Test_PanicIfError(t *testing.T) {
	PanicIfError(nil)

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("PanicIfError() didn't panic on a non-nil real error")
		}
	}()

	PanicIfError(errors.New("A real error!"))
}

// vim:nocindent
