package main

import (
	"kmud/server"
	"os"
	"os/signal"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(8)

	go signalHandler()

	var s server.Server
	s.Exec()
}

func signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		<-c
		stack := make([]byte, 1024*10)
		runtime.Stack(stack, true)
		os.Stderr.Write(stack)
		os.Exit(0)
	}
}
