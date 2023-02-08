package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pomdtr/sunbeam/app"
)

func NewServer(extensions []*app.Extension, addr string) *http.Server {
	extensionMap := make(map[string]*app.Extension)
	for _, extension := range extensions {
		extensionMap[extension.Name()] = extension
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: return a page with all extensions
		extensionNames := make([]string, 0, len(extensions))
		for _, extension := range extensions {
			extensionNames = append(extensionNames, extension.Name())
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(extensionNames)
	})

	r.Get("/{extension}", func(w http.ResponseWriter, r *http.Request) {
		extensionName := chi.URLParam(r, "extension")
		extension, ok := extensionMap[extensionName]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Extension %s not found", extensionName)))
			return
		}

		commands := make([]app.Command, len(extension.Commands))
		for i, command := range extension.Commands {
			command.Exec = buildExec(command)
			command.Params = append(command.Params, app.Param{
				Name: "sunbeam-remote",
				Type: "string",
				Env:  "SUNBEAM_REMOTE",
			})
			commands[i] = command
		}

		w.Header().Set("Content-Type", "application/json")
		// TODO: return a page with only commands from the extension
		json.NewEncoder(w).Encode(app.Extension{
			Version:     extension.Version,
			Title:       extension.Title,
			Description: extension.Description,
			RootItems:   extension.RootItems,
			Commands:    commands,
		})
	})

	r.Post("/{extension}/{command}", func(w http.ResponseWriter, r *http.Request) {
		var params map[string]any
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil && err != io.EOF {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error decoding input params: %s", err)))
			return
		}

		extensionName := chi.URLParam(r, "extension")
		extension, ok := extensionMap[extensionName]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Extension %s not found", extensionName)))
			return
		}

		commandName := chi.URLParam(r, "command")
		command, ok := extension.GetCommand(commandName)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Command %s not found", commandName)))
			return
		}

		cmd, err := command.Cmd(params, extension.Root)
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

func buildExec(command app.Command) string {
	args := []string{"sunbeam", "http", "--ignore-stdin", "POST", "{{ sunbeam_remote }}"}

	for _, param := range command.Params {
		args = append(args, fmt.Sprintf("%s={{%s}}", param.Name, param.Name))
	}

	return strings.Join(args, " ")
}
