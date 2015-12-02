package main

import (
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/Cristofori/kmud/server"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())

	go signalHandler()

	var s server.Server
	s.Exec()
}

func signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		<-c
		// stack := make([]byte, 1024*10)
		// runtime.Stack(stack, true)
		// os.Stderr.Write(stack)
		os.Exit(0)
	}
}
