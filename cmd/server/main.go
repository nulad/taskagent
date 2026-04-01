package main

import "github.com/nulad/taskagent/internal/config"

func main() {
	_ = config.Load()
	// TODO: initialize and run server
}