package utils

import "testing"

var wrote string
var read string

type testWriter struct {
}

func (self testWriter) Write(p []byte) (n int, err error) {
	wrote = string(p)
	return len(p), nil
}

type testReader struct {
}

func (self testReader) Read(p []byte) (n int, err error) {
	for i := 0; i < 10; i++ {
		char := byte('a' + i)
		p[i] = char
		read = read + string(char)
	}
	p[10] = '\n'

	read = string(p)
	return 11, nil
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

func Test_WriteLine(t *testing.T) {
	var conn testWriter

	line := "This is a line"
	want := line + "\r\n"

	WriteLine(conn, line)

	if wrote != want {
		t.Errorf("WriteLine(%q) == %q, want %q", line, wrote, want)
	}
}

func Test_ReadLine(t *testing.T) {
	var conn testReader
	line, _ := readLine(conn)

	want := "abcdefghij"

	if line != want {
		t.Errorf("ReadLine(%q) == %q, want %q", line, read, want)
	}
}

// vim:nocindent
