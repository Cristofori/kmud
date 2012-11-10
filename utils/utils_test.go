package utils

import "testing"

var wrote string

type testWriter struct {
}

func (self testWriter) Write(p []byte) (n int, err error) {
	wrote = string(p)
	return len(p), nil
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

// vim:nocindent
