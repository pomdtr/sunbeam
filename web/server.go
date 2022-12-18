package web

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/creack/pty"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

//go:embed frontend/dist
var frontendDist embed.FS

func NewRouter(theme string) (http.Handler, error) {
	r := mux.NewRouter()

	r.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.HandleFunc("/ws", WebsocketHandle)
	r.HandleFunc("/theme.json", func(w http.ResponseWriter, r *http.Request) {
		themePath := fmt.Sprintf("./frontend/dist/themes/%s.json", theme)
		http.ServeFile(w, r, themePath)
	})
	r.HandleFunc("/run/{extension}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		extension := vars["extension"]

		// redirect to the index.html file
		http.Redirect(w, r, fmt.Sprintf("/index.html?extension=%s", extension), http.StatusFound)
	})
	r.HandleFunc("/run/{extension}/{script}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		extension := vars["extension"]
		script := vars["script"]

		var arguments []string
		for name, value := range r.URL.Query() {
			arg := fmt.Sprintf("--%s=%s", name, value[0])
			arg = url.QueryEscape(arg)
			arguments = append(arguments, fmt.Sprintf("arg=%s", arg))
		}

		query := strings.Join(arguments, "&")

		http.Redirect(w, r, fmt.Sprintf("/index.html?extension=%s&script=%s&%s", extension, script, query), http.StatusFound)
	})

	sub, err := fs.Sub(frontendDist, "frontend/dist")
	if err != nil {
		return nil, err
	}
	r.PathPrefix("/").Handler(http.FileServer(http.FS(sub)))

	return r, nil
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
	defer ws.Close()

	command := exec.Command("sunbeam", arguments...)
	var errBuf bytes.Buffer
	command.Stderr = &errBuf
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
					if errBuf.String() != "" {
						log.Printf("text from stderr: %v", errBuf.String())
						// ws.WriteMessage(websocket.TextMessage, errBuf.Bytes())
					}
					// if the pty is closed, restart it
					if err.Error() == "EOF" {
						log.Println("pty closed, restarting")
						command = exec.Command("sunbeam")
						errBuf.Reset()
						command.Stderr = &errBuf
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
