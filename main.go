package main

import (
	"flag"
	"github.com/kochman/buildstatus/server"
)

func main() {
	var ghat = flag.String("ghat", "", "GitHub access token")
	flag.Parse()
	server.Serve(*ghat)
}
