package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/pomdtr/sunbeam/app"
)

func NewServer(extensions map[string]app.Extension, addr string) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/extensions", func(w http.ResponseWriter, r *http.Request) {
		extensionNames := make([]string, 0, len(extensions))
		for name := range extensions {
			extensionNames = append(extensionNames, name)
		}
		json.NewEncoder(w).Encode(extensionNames)
	})

	r.Get("/extensions/{extension}", func(w http.ResponseWriter, r *http.Request) {
		extensionName := chi.URLParam(r, "extension")
		extension, ok := extensions[extensionName]

		extension.PostInstall = ""
		extension.Requirements = nil

		for name, command := range extension.Commands {
			command.Exec = ""
			command.Url = fmt.Sprintf("/run/%s/%s", extensionName, name)
			extension.Commands[name] = command
		}

		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Extension %s not found", extensionName)))
			return
		}

		json.NewEncoder(w).Encode(extension)
	})

	r.Post("/extensions/{extension}/{command}", func(w http.ResponseWriter, r *http.Request) {
		extensionName := chi.URLParam(r, "extension")
		extension, ok := extensions[extensionName]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Extension %s not found", extensionName)))
			return
		}

		commandName := chi.URLParam(r, "command")
		command, ok := extension.Commands[commandName]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Command %s not found", commandName)))
			return
		}

		var input app.CommandParams
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil && err != io.EOF {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error decoding input params: %s", err)))
			return
		}

		cmd, err := command.Cmd(input, extension.Root)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error running command: %s", err)))
			return
		}

		output, err := cmd.Output()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error running command: %s", err)))
			return
		}

		w.Write(output)
	})

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
