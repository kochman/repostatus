package main

import (
	"github.com/kochman/repostatus/server"
	"github.com/kochman/repostatus/travis"
	"os"
)

func main() {
	ghat := os.Getenv("GHAT")
	redisURL := os.Getenv("REDIS_URL")

	updater := travis.Updater{GitHubAccessToken: ghat, RedisURL: redisURL}
	go updater.Run()

	server.Serve(ghat, redisURL)
}
