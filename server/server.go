package server

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/kochman/repostatus/travis"
	"log"
	"net/http"
	"time"
)

type wsHandler struct {
	GitHubAccessToken string
	RedisURL          string
}

type wsMessage struct {
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "status.html")
}

func (ws wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// dump all read messages into channel
	readCh := make(chan []byte)
	go func(ch chan []byte, conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if websocket.IsUnexpectedCloseError(err) {
				close(ch)
				return
			} else if err != nil {
				log.Fatal(err)
			}
			ch <- msg
		}
	}(readCh, conn)

	writeCh := make(chan []byte)
	stopCh := make(chan []struct{})
	go func() {
		for {
			select {
			case msg := <-writeCh:
				conn.WriteMessage(websocket.TextMessage, msg)
			case msg, ok := <-readCh:
				if !ok {
					// channel closed
					readCh = nil
					close(stopCh)
					continue
				}
				wsm := &wsMessage{}
				json.Unmarshal(msg, wsm)

				if wsm.Command == "subscribe" {
					org, ok := wsm.Data["org"].(string)
					if !ok {
						return
					}
					repo, ok := wsm.Data["repo"].(string)
					if !ok {
						return
					}
					repoSlug := org + "/" + repo
					log.Println("subscribe " + repoSlug)
					updater := travis.Updater{GitHubAccessToken: ws.GitHubAccessToken, RedisURL: ws.RedisURL}
					updater.SubscribeRepo(repoSlug)
					go func(ch chan []byte, repo string) {
						ticker := time.Tick(time.Second * 5)

						// first update
						repository, err := updater.GetRepo(repoSlug)
						if err != nil {
							log.Println(err)
							return
						}
						b, err := json.Marshal(repository)
						if err != nil {
							log.Println(err)
							return
						}
						ch <- b

						for {
							select {
							case <-ticker:
								repository, err := updater.GetRepo(repoSlug)
								updater.SubscribeRepo(repoSlug)
								if err != nil {
									log.Println(err)
									return
								}
								newB, err := json.Marshal(repository)
								if err != nil {
									log.Println(err)
									return
								}
								if !bytes.Equal(b, newB) {
									b = newB
									ch <- b
								}
								log.Println("update " + org + "/" + repo)
							case _, open := <-stopCh:
								if !open {
									return
								}
							}
						}
					}(writeCh, repo)
				}
			}
		}
	}()
}

func Serve(ghat string, redisURL string) {
	wsh := wsHandler{GitHubAccessToken: ghat, RedisURL: redisURL}

	http.HandleFunc("/", indexHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/ws", wsh)
	http.ListenAndServe(":5000", nil)
}
