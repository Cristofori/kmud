package main

import "kmud/server"
import "runtime"

func main() {
	runtime.GOMAXPROCS(8)
	server.Exec()
}
