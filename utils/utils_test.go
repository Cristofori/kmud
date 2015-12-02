package utils

import (
	"errors"
	"io"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/Cristofori/kmud/testutils"
	"github.com/Cristofori/kmud/types"
)

func Test_WriteLine(t *testing.T) {
	writer := &testutils.TestWriter{}

	line := "This is a line"
	want := line + "\r\n"

	WriteLine(writer, line, types.ColorModeNone)

	if writer.Wrote != want {
		t.Errorf("WriteLine(%q) == %q, want %q", line, writer.Wrote, want)
	}
}

func Test_Simplify(t *testing.T) {
	tests := []struct {
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
	readWriter := &testutils.TestReadWriter{}

	tests := []struct {
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
		readWriter.ToRead = test.input
		line := GetRawUserInput(readWriter, ">", types.ColorModeNone)
		if line != test.output {
			t.Errorf("GetRawUserInput(%q) == %q, want %q", test.input, line, test.output)
		}
	}
}

func Test_GetUserInput(t *testing.T) {
	readWriter := &testutils.TestReadWriter{}

	tests := []struct {
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
		readWriter.ToRead = test.input
		line := GetUserInput(readWriter, ">", types.ColorModeNone)
		if line != test.output {
			t.Errorf("GetUserInput(%q) == %q, want %q", test.input, line, test.output)
		}
	}
}

func Test_GetUserInputPanicOnEOF(t *testing.T) {
	readWriter := &testutils.TestReadWriter{}
	readWriter.SetError(io.EOF)

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("GetUserInput() didn't panic on EOF")
		}
	}()

	GetUserInput(readWriter, "", types.ColorModeNone)
}

func Test_HandleError(t *testing.T) {
	// TODO try using recover
	HandleError(nil)
}

func Test_FormatName(t *testing.T) {
	tests := []struct {
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
	tests := []struct {
		input  string
		output bool
	}{
		{"t", false},
		{"te", false},
		{"tes", true},
		{"te1", true},
		{"test", true},
		{"testing", true},
		{"*(!(@#*$", false},
		{"Abc1abc", true},
		{"123456789012", false},
		{"aslsidjfljll", true},
		{"1slsidjfljll", false},
		{"aslsidjfljl3", true},
	}

	for _, test := range tests {
		result := (ValidateName(test.input) == nil)
		if result != test.output {
			t.Errorf("ValidateName(%s) == %v, want %v", test.input, result, test.output)
		}
	}
}

func Test_BestMatch(t *testing.T) {
	searchList := []string{"", "Foo", "Bar", "Joe", "Bob", "Abcdef", "Abc", "QrStUv"}

	tests := []struct {
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

func Test_Argify(t *testing.T) {
	tests := []struct {
		input   string
		output1 string
		output2 string
	}{
		{"", "", ""},
		{"test", "test", ""},
		{"test two", "test", "two"},
		{"test one two", "test", "one two"},
		{"this is a somewhat longer test that should also work",
			"this", "is a somewhat longer test that should also work"},
	}

	for _, test := range tests {
		result1, result2 := Argify(test.input)

		if result1 != test.output1 || reflect.DeepEqual(result2, test.output2) == false {
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

	tests := []struct {
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

		if reflect.DeepEqual(result, test.output) == false {
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

	tests := []struct {
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

		if reflect.DeepEqual(result, test.output) == false {
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

	tests := []struct {
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

func Test_Random(t *testing.T) {
	tests := []struct {
		low  int
		high int
	}{
		{0, 0},
		{0, 1},
		{-10, 0},
		{1, 2},
		{1000, 2000},
	}

	for i := 0; i < 100; i++ {
		rand.Seed(int64(i))
		for _, test := range tests {
			result := Random(test.low, test.high)

			if result < test.low || result > test.high {
				t.Errorf("Random number was out of range %v-%v, got %v", test.low, test.high, result)
			}
		}
	}
}

func Test_Atois(t *testing.T) {
	tests := []struct {
		input  []string
		output []int
		err    error
	}{
		{[]string{"0"}, []int{0}, nil},
		{[]string{"0", "1", "3"}, []int{0, 1, 3}, nil},
		{[]string{"asdf"}, []int{}, errors.New("not nil")},
		{[]string{"1", "2", "asdf"}, []int{}, errors.New("not nil")},
	}

	for _, test := range tests {
		output, err := Atois(test.input)
		if (err == nil && test.err != nil) || (err != nil && test.err == nil) {
			t.Errorf("Error flags did not match: %v, got %v", test.input, err)
		} else if err == nil {
			if len(output) != len(test.output) {
				t.Errorf("Mismatched length: %v, %v", output, test.output)
			} else {
				for i, o := range output {
					if o != test.output[i] {
						t.Errorf("Expected: %v, got: %v", test.output, output)
					}
				}
			}
		}
	}
}

func Test_Atoir(t *testing.T) {
	tests := []struct {
		input  string
		min    int
		max    int
		output int
		err    error
	}{
		{"1", 0, 10, 1, nil},
		{"asdf", 0, 10, 0, errors.New("error")},
		{"999999", -5, 5, 0, errors.New("error")},
		{"-10", -5, 5, 0, errors.New("error")},
		{"-3", -5, 5, -3, nil},
		{"3", -5, 5, 3, nil},
	}

	for _, test := range tests {
		output, err := Atoir(test.input, test.min, test.max)
		if (err == nil && test.err != nil) || (err != nil && test.err == nil) {
			t.Errorf("Error flags did not match: %v, got %v", test.input, err)
		} else if err == nil {
			if test.output != output {
				t.Errorf("Expected %v, got %v", test.output, output)
			}
		}
	}
}

func Test_Min(t *testing.T) {
	tests := []struct {
		x      int
		y      int
		output int
	}{
		{0, 1, 0},
		{-10, 10, -10},
		{100, 200, 100},
		{200, 100, 100},
		{0, 0, 0},
		{1, 1, 1},
		{-1, -1, -1},
	}

	for _, test := range tests {
		output := Min(test.x, test.y)
		if output != test.output {
			t.Errorf("Min(%v, %v) = %v, expected %v", test.x, test.y, output, test.output)
		}
	}
}

func Test_Max(t *testing.T) {
	tests := []struct {
		x      int
		y      int
		output int
	}{
		{0, 1, 1},
		{-10, 10, 10},
		{100, 200, 200},
		{200, 100, 200},
		{0, 0, 0},
		{1, 1, 1},
		{-1, -1, -1},
	}

	for _, test := range tests {
		output := Max(test.x, test.y)
		if output != test.output {
			t.Errorf("Max(%v, %v) = %v, expected %v", test.x, test.y, output, test.output)
		}
	}
}

func Test_Bound(t *testing.T) {
	tests := []struct {
		input  int
		lower  int
		upper  int
		output int
	}{
		{5, 0, 10, 5},
		{-10, 0, 10, 0},
		{20, 0, 10, 10},
		{-15, -20, -10, -15},
		{19, 10, 20, 19},
	}

	for _, test := range tests {
		output := Bound(test.input, test.lower, test.upper)
		if output != test.output {
			t.Errorf("Bound(%v, %v, %v) = %v, expected %v", test.input, test.lower, test.upper, output, test.output)
		}
	}
}
