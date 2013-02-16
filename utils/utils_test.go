package utils

import "testing"

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
		{"ABC", -2},
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

// vim:nocindent
