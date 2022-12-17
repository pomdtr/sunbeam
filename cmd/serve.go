package cmd

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

func NewCmdServe() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:  "serve",
		RunE: StartServer,
	}

	return serveCmd
}

func StartServer(cmd *cobra.Command, args []string) error {
	r := mux.NewRouter()

	r.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.HandleFunc("/ws", WebsocketHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/")))

	return http.ListenAndServe(":8080", r)
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,

		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	command := exec.Command("sunbeam")
	tty, err := pty.Start(command)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	waiter := sync.WaitGroup{}
	waiter.Add(3)

	// this is a keep-alive loop that ensures connection does not hang-up itself
	lastPongTime := time.Now()
	ws.SetPongHandler(func(msg string) error {
		lastPongTime = time.Now()
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Send a ping message every 5 seconds to keep the connection alive
	keepAliveTimeout := 10 * time.Second
	go func() {
		defer waiter.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := ws.WriteMessage(websocket.PingMessage, []byte("keepalive")); err != nil {
					log.Printf("error writing ping message: %v", err)
					cancel()
					return
				}
				time.Sleep(keepAliveTimeout / 2)
				if time.Since(lastPongTime) > keepAliveTimeout {
					log.Printf("no pong message received for %v, closing connection", keepAliveTimeout)
					cancel()
					return
				}
			}

		}
	}()

	var ptySize pty.Winsize
	// this is a loop that reads from the websocket and writes to the pty
	go func() {
		defer waiter.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				messageType, message, err := ws.ReadMessage()
				if err != nil {
					log.Printf("error reading from websocket: %v", err)
					cancel()
					return
				}

				if messageType == websocket.BinaryMessage {

					if err = json.Unmarshal(message, &ptySize); err != nil {
						log.Printf("error unmarshalling pty size: %v", err)
						cancel()
						return
					}

					if err = pty.Setsize(tty, &ptySize); err != nil {
						log.Printf("error setting pty size: %v", err)
						cancel()
						return
					}
					continue
				}

				if _, err := tty.Write(message); err != nil {
					log.Printf("error writing to pty: %v", err)
					cancel()
					return
				}
			}
		}
	}()

	// this is a loop that reads from the pty and writes to the websocket
	go func() {
		defer waiter.Done()
		select {
		case <-ctx.Done():
			return
		default:
			for {
				buf := make([]byte, 1024)
				n, err := tty.Read(buf)
				if err != nil {
					// if the pty is closed, restart it
					if err.Error() == "EOF" {
						command = exec.Command("sunbeam")
						tty, _ = pty.Start(command)
						pty.Setsize(tty, &ptySize)
					} else {
						log.Printf("error reading from pty: %v", err)
						cancel()
						return
					}
				}
				if err := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
					log.Printf("error writing to websocket: %v", err)
					cancel()
					return
				}
			}
		}

	}()

	waiter.Wait()
	log.Println("all goroutines exited, closing connection")
}
