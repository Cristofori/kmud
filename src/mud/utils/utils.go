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
	return string(bytes), err
}

func Simplify(str string) string {
	simpleStr := strings.TrimSpace(str)
	simpleStr = strings.ToLower(str)
	return simpleStr
}

func GetUserInput(conn net.Conn, prompt string) string {
	input := GetRawUserInput(conn, prompt)
	return Simplify(input)
}

func GetRawUserInput(conn net.Conn, prompt string) string {
	for {
		io.WriteString(conn, prompt)
		input, err := readLine(conn)
		PanicIfError(err)

		if input == "x" || input == "X" {
			return ""
		} else if input != "" {
			return input
		}
	}

	panic("Unexpected code path")
	return ""
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

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

// vim: nocindent
