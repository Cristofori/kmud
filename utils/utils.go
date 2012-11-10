package utils

import (
	"bufio"
	"io"
	"log"
	"strings"
	"unicode"
)

func WriteLine(conn io.Writer, line string) (int, error) {
	return io.WriteString(conn, line+"\r\n")
}

func Simplify(str string) string {
	simpleStr := strings.TrimSpace(str)
	simpleStr = strings.ToLower(simpleStr)
	return simpleStr
}

func GetRawUserInput(conn io.ReadWriter, prompt string) string {
	reader := bufio.NewReader(conn)

	for {
		io.WriteString(conn, prompt)

		bytes, _, err := reader.ReadLine()
		input := string(bytes)

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

func GetUserInput(conn io.ReadWriter, prompt string) string {
	input := GetRawUserInput(conn, prompt)
	return Simplify(input)
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

func Argify(data string) (string, []string) {
	fields := strings.Fields(data)

	if len(fields) == 0 {
		return "", []string{}
	}

	arg1 := fields[0]
	args := fields[1:]

	return arg1, args
}

// vim: nocindent
