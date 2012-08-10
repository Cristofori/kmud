package utils

import (
    "io"
    "log"
    "net"
    "bufio"
    "strings"
)

func WriteLine(conn net.Conn, line string) (int, error) {
	return io.WriteString(conn, line+"\n")
}

func readLine(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	bytes, _, err := reader.ReadLine()
	line := string(bytes)
	line = strings.TrimSpace(line)
	line = strings.ToLower(line)
	return line, err
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

// vim: nocindent
