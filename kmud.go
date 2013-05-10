package main

import "kmud/server"
import "runtime"

func main() {
	runtime.GOMAXPROCS(8)

	var s server.Server
	s.Exec()
}
