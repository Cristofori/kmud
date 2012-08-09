package main

import (
	"bufio"
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
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
	Actions map[string]func(*mgo.Session, net.Conn) error
	Text    string
}

func NewMenu() Menu {
	var menu Menu
	menu.Actions = map[string]func(*mgo.Session, net.Conn) error{}
	return menu
}

func (self *Menu) Exec(session *mgo.Session, conn net.Conn) error {

	for {
		io.WriteString(conn, self.Text)
		input, err := readLine(conn)

		if err != nil {
			return err
		}

		function, ok := self.Actions[input]

		if ok {
			return function(session, conn)
			break
		}
	}

	return nil
}

func login(session *mgo.Session, conn net.Conn) error {

	for {
		io.WriteString(conn, "Username: ")
		line, err := readLine(conn)

		if err != nil {
			return err
		}

		c := session.DB("mud").C("users")
		q := c.Find(bson.M{"name": line})

		count, err := q.Count()

		if err != nil {
			return err
		}

		if count == 0 {
			io.WriteString(conn, "User not found")
		} else if count == 1 {
			result := map[string]string{}
			err := q.One(&result)

			if err != nil {
				return err
			}

			io.WriteString(conn, "Logging in as: "+line)
			break
		}
	}

	return nil
}

func newUser(session *mgo.Session, conn net.Conn) error {

	for {
		io.WriteString(conn, "Desired username: ")

		line, err := readLine(conn)

		if err != nil {
			return err
		}

		c := session.DB("mud").C("users")
		q := c.Find(bson.M{"name": line})

		count, err := q.Count()

		if err != nil {
			return err
		}

		if count == 0 {
			c.Insert(bson.M{"name": line})
			break
		}

		io.WriteString(conn, "That username is already in use\n")
	}

	return nil
}

func quit(session *mgo.Session, conn net.Conn) error {
	io.WriteString(conn, "Goodbye!\n")
	conn.Close()
	return nil
}

func mainMenu() Menu {

	menu := NewMenu()

	menu.Text = `
----- MUD ------
[L]ogin
[N]ew user
[A]bout
[Q]uit
> `

	menu.Actions["l"] = login
	menu.Actions["n"] = newUser
	menu.Actions["q"] = quit

	return menu
}

func handleConnection(session *mgo.Session, conn net.Conn) {

	defer conn.Close()
	defer session.Close()

	menu := mainMenu()
	err := menu.Exec(session, conn)

	if err != nil {
		return
	}

	for {
		io.WriteString(conn, "\n> ")

		line, err := readLine(conn)

		if err != nil {
			fmt.Printf("Lost connection to client\n")
			break
		}

		if line == "quit" || line == "exit" {
			quit(session, conn)
			break
		}

		// if line == "x" || line == "logout" || line == "logoff" {
		// }

		io.WriteString(conn, line)
	}
}

func main() {

	fmt.Printf("Connecting to database... ")
	session, err := mgo.Dial("localhost")

	handleError(err)

	fmt.Printf("done.\n")

	listener, err := net.Listen("tcp", ":8945")
	handleError(err)

	fmt.Printf("Server listening on port 8945\n")

	for {
		conn, err := listener.Accept()
		handleError(err)
		go handleConnection(session.Copy(), conn)
	}
}

// vim: nocindent
