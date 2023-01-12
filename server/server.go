package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

func New(address string) *http.Server {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Sunbeam Server is running..."))
	})

	http.HandleFunc("/ws", WebsocketHandle)

	return &http.Server{
		Addr: address,
	}
}

func extractArgs(r *http.Request) (arguments []string) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	extension, ok := r.Form["extension"]
	if !ok {
		return arguments
	}
	arguments = append(arguments, "run", extension[0])

	script, ok := r.Form["script"]
	if !ok {
		return arguments
	}
	arguments = append(arguments, script[0])

	args, ok := r.Form["arg"]
	if !ok {
		return arguments
	}

	arguments = append(arguments, args...)
	return arguments
}

func WebsocketHandle(w http.ResponseWriter, r *http.Request) {

	arguments := extractArgs(r)

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

	mu := sync.Mutex{}
	send := func(messageType int, data []byte) error {
		mu.Lock()
		defer mu.Unlock()
		return ws.WriteMessage(messageType, data)
	}

	sunbeamPath := os.Args[0]

	command := exec.Command(sunbeamPath, arguments...)
	f, err := os.CreateTemp("", "sunbeam-command-*")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	defer f.Close()

	command.Env = []string{
		fmt.Sprintf("SUNBEAM_COMMAND_OUTPUT=%s", f.Name()),
	}
	command.Env = append(command.Env, os.Environ()...)

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

	// Send a ping message every 5 seconds to keep the connection alive
	keepAliveTimeout := 10 * time.Second
	go func() {
		defer waiter.Done()
		defer ws.Close()

		for {
			if err := send(websocket.PingMessage, []byte("keepalive")); err != nil {
				log.Printf("error writing ping message: %v", err)
				return
			}
			time.Sleep(keepAliveTimeout / 2)
			if time.Since(lastPongTime) > keepAliveTimeout {
				log.Printf("no pong message received for %v, closing connection", keepAliveTimeout)
				return
			}

		}
	}()

	var ptySize pty.Winsize
	// this is a loop that reads from the websocket and writes to the pty
	go func() {
		defer waiter.Done()
		defer ws.Close()

		for {
			messageType, message, err := ws.ReadMessage()
			if err != nil {
				log.Printf("error reading from websocket: %v", err)
				return
			}

			if messageType == websocket.BinaryMessage {

				if err = json.Unmarshal(message, &ptySize); err != nil {
					log.Printf("error unmarshalling pty size: %v", err)
					continue
				}

				if err = pty.Setsize(tty, &ptySize); err != nil {
					log.Printf("error setting pty size: %v", err)
					continue
				}
				continue
			}

			if _, err := tty.Write(message); err != nil {
				log.Printf("error writing to pty: %v", err)
				return
			}
		}
	}()

	// this is a loop that reads from the pty and writes to the websocket
	go func() {
		defer waiter.Done()
		defer ws.Close()

		for {
			buf := make([]byte, 1024)
			n, err := tty.Read(buf)
			if err != nil {
				buf := bytes.Buffer{}
				_, err := buf.ReadFrom(f)
				if err != nil {
					log.Println(err)
					return
				}

				if buf.Len() > 0 {
					send(websocket.TextMessage, buf.Bytes())
				}

				log.Printf("error reading from pty: %v", err)
				return
			}

			if err := send(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Printf("error writing to websocket: %v", err)
				return
			}
		}

	}()

	waiter.Wait()
	log.Println("Killing process.")
	command.Process.Kill()
}
