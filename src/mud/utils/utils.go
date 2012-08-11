package utils

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
	"unicode"
)

func WriteLine(conn net.Conn, line string) (int, error) {
	return io.WriteString(conn, line+"\n")
}

func readLine(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	bytes, _, err := reader.ReadLine()
	line := string(bytes)
	line = Simplify(line)
	return line, err
}

func Simplify(str string) string {
	simpleStr := strings.TrimSpace(str)
	simpleStr = strings.ToLower(str)
	return simpleStr
}

func GetUserInput(conn net.Conn, prompt string) (string, error) {
	for {
		io.WriteString(conn, prompt)
		input, err := readLine(conn)

		if err != nil {
			return "", err
		}

		if input != "" {
			return input, nil
		}
	}

	panic("Unexpected code path")
	return "", nil
}

func HandleError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func FormatName(name string) string {
	if name == "" {
		return name
	}

	runes := []rune(Simplify(name))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// vim: nocindent
