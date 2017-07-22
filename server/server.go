package server

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/kochman/buildstatus/travis"
	"log"
	"net/http"
	"time"
)

type wsHandler struct {
	GitHubAccessToken string
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
					repo := wsm.Data["repo"].(string)
					log.Println(repo)
					go func(ch chan []byte, repo string) {
						tc := travis.Client{RepoSlug: repo, GitHubAccessToken: ws.GitHubAccessToken}
						ticker := time.Tick(time.Minute)

						// first update
						branches, err := tc.Branches()
						if err != nil {
							log.Println(err)
							return
						}
						b, err := json.Marshal(branches)
						if err != nil {
							log.Println(err)
							return
						}
						ch <- b

						for {
							select {
							case <-ticker:
								branches, err := tc.Branches()
								if err != nil {
									log.Println(err)
									return
								}
								newB, err := json.Marshal(branches)
								if err != nil {
									log.Println(err)
									return
								}
								if !bytes.Equal(b, newB) {
									b = newB
									ch <- b
								}
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

/*
func (ws wsHandler) updater() {
	var chans []chan []byte
	tc := travis.Client{}
	ticker := time.Tick(time.Minute)

	// first update
	branches, err := tc.Branches()
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.Marshal(branches)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ticker:
			branches, err := tc.Branches()
			if err != nil {
				log.Fatal(err)
			}
			newB, err := json.Marshal(branches)
			if err != nil {
				log.Fatal(err)
			}
			if !bytes.Equal(b, newB) {
				b = newB
				for _, ch := range chans {
					log.Println("periodic send")
					ch <- b
				}
			}
		case ch := <-ws.newClientCh:
			log.Println("sending")
			ch <- b
			chans = append(chans, ch)
		}
	}
}*/

func Serve(ghat string) {
	wsh := wsHandler{GitHubAccessToken: ghat}

	http.HandleFunc("/", indexHandler)
	http.Handle("/ws", wsh)
	http.ListenAndServe("localhost:5000", nil)
}
