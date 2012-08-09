package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func readLine(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
    bytes, _, err := reader.ReadLine()
    line := string(bytes)
    line = strings.TrimSpace(line)
    line = strings.ToLower(line)
    return line, err
}

func handleError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

type Menu struct {
    Actions map[string]func( net.Conn ) error
    Text string
}

func NewMenu() Menu {
    var menu Menu
    menu.Actions = map[string]func(net.Conn)error{}
    return menu
}

func (self *Menu) Exec(conn net.Conn) error {

    for {
        io.WriteString(conn, self.Text)
        input, err := readLine(conn)

        if err != nil {
            return err
        }

        function, ok := self.Actions[input]

        if ok {
            return function(conn)
            break
        }
    }

    return nil
}

func login( conn net.Conn ) error {
    io.WriteString(conn, "What's your name? ")
    line, err := readLine(conn)

    if err != nil {
        return err
    }

    io.WriteString(conn, "Logging in as: " + line )

    return nil
}

func mainMenu() Menu {

    menu := NewMenu()

    menu.Text = `
----- MUD ------
[L]ogin
[N]ew user
[A]bout
> `

    menu.Actions["l"] = login;

    return menu
}

func handleConnection(conn net.Conn) {

    defer conn.Close()

    menu := mainMenu()
    err := menu.Exec(conn)

    if err != nil {
        return
    }

	for {
        io.WriteString(conn, "\n> ")

		line, err := readLine(conn)

        if err != nil {
            fmt.Printf( "Lost connection to client\n" )
            break
        }

		if line == "quit" || line == "exit" {
			io.WriteString(conn, "Goodbye!\n")
			break
		}

		io.WriteString(conn, line)
	}
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
