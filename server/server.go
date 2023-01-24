package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pomdtr/sunbeam/app"
	"gopkg.in/yaml.v3"
)

func NewServer(extensions map[string]app.Extension, addr string) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Sunbeam server is running")
	})

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
		extension.Root = nil

		for name, command := range extension.Commands {
			command.Exec = ""
			extension.Commands[name] = command
		}

		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Extension %s not found", extensionName)))
			return
		}

		yaml.NewEncoder(w).Encode(extension)
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

		cmd, err := command.Cmd(input, extension.Root.Path)
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

		_, err = w.Write(output)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error writing response: %s", err)))
			return
		}
	})

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
