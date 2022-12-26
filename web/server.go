package web

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

//go:embed frontend/dist
var frontendDist embed.FS

func NewServer(address string) (*http.Server, error) {
	sub, err := fs.Sub(frontendDist, "frontend/dist")
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()

	r.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.HandleFunc("/ws", WebsocketHandle)
	r.HandleFunc("/run/{extension}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		extension := vars["extension"]

		// redirect to the index.html file
		http.Redirect(w, r, fmt.Sprintf("/?extension=%s", extension), http.StatusSeeOther)
	})
	r.HandleFunc("/run/{extension}/{script}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		extension := vars["extension"]
		script := vars["script"]

		var arguments []string
		for name, value := range r.URL.Query() {
			var arg string
			if len(value) == 0 {
				arg = fmt.Sprintf("--%s", name)
			} else {
				arg = fmt.Sprintf("--%s=%s", name, value[0])
			}
			arg = url.QueryEscape(arg)
			arguments = append(arguments, fmt.Sprintf("arg=%s", arg))
		}

		query := strings.Join(arguments, "&")

		http.Redirect(w, r, fmt.Sprintf("/?extension=%s&script=%s&%s", extension, script, query), http.StatusSeeOther)
	})

	tmpl := template.Must(template.ParseFS(sub, "index.html"))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		theme := r.Header.Get("X-Sunbeam-Theme")
		if fs.Stat(sub, path.Join("themes", fmt.Sprintf(theme, ".json"))); errors.Is(err, fs.ErrNotExist) {
			theme = "tomorrow-night"
		}
		tmpl.Execute(w, map[string]string{
			"theme": theme,
		})
	})
	r.PathPrefix("/").Handler(http.FileServer(http.FS(sub)))

	return &http.Server{
		Addr:    address,
		Handler: r,
	}, nil
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

	mu := sync.Mutex{}
	send := func(messageType int, data []byte) error {
		mu.Lock()
		defer mu.Unlock()
		return ws.WriteMessage(messageType, data)
	}

	var sunbeamPath string
	if len(os.Args) == 0 {
		sunbeamPath = "sunbeam"
	} else {
		sunbeamPath = os.Args[0]
	}

	command := exec.Command(sunbeamPath, arguments...)
	command.Env = []string{
		"SUNBEAM_SERVER=1",
	}
	command.Env = append(command.Env, os.Environ()...)

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

	// Send a ping message every 5 seconds to keep the connection alive
	keepAliveTimeout := 10 * time.Second
	go func() {
		defer waiter.Done()
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

		for {
			buf := make([]byte, 1024)
			n, err := tty.Read(buf)
			if err != nil {
				if errBuf.String() != "" {
					send(websocket.TextMessage, errBuf.Bytes())
				} else {
					action, _ := json.Marshal(map[string]any{"action": "exit"})
					send(websocket.TextMessage, action)
				}
				return
			}

			if err := send(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Printf("error writing to websocket: %v", err)
				return
			}
		}

	}()

	waiter.Wait()
	command.Process.Kill()
}
