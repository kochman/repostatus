package main

import (
	"github.com/kochman/repostatus/server"
	"os"
)

func main() {
	ghat := os.Getenv("GHAT")
	server.Serve(ghat)
}
