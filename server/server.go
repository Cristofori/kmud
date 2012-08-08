package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
    // "labix.org/v2/mgo"
    // "labix.org/v2/mgo/bson"
)

func handleError(err error) {
	if err != nil {
		log.Fatal("Error: %v\n", err)
	}
}

func handleConnection(conn net.Conn) {
	io.WriteString(conn, "Welcome!\n")

	reader := bufio.NewReader(conn)

	for {
		io.WriteString(conn, "> ")
		bytes, _, err := reader.ReadLine()
		handleError(err)

		line := string(bytes)
		line = strings.TrimSpace(line)
		line = strings.ToLower(line)

		if line == "quit" || line == "exit" {
			io.WriteString(conn, "Goodbye!\n")
			break
		}

		io.WriteString(conn, line+"\n")
	}

	conn.Close()
}

func main() {
	fmt.Printf("Here's a server!\n")

	listener, err := net.Listen("tcp", ":8945")
	handleError(err)

	for {
		conn, err := listener.Accept()
		handleError(err)
		go handleConnection(conn)
	}
}

// vim: nocindent
