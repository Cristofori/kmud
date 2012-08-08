package main

import "fmt"
import "net"
import "log"

func handleError( err error ) {
    if err != nil {
        log.Fatal( "Error: %v\n", err )
    }
}

func handleConnection(conn net.Conn) {
    conn.Write( []byte("Hello!\n") )
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
