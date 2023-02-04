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
			command.Exec = ""
			commands[i] = command
		}

		w.Header().Set("Content-Type", "application/json")
		// TODO: return a page with only commands from the extension
		json.NewEncoder(w).Encode(app.Extension{
			Version:     extension.Version,
			Description: extension.Description,
			Preferences: extension.Preferences,
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
		environ := make(map[string]string)
		for _, env := range r.Header.Values("X-Sunbeam-Env") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				environ[parts[0]] = parts[1]
			}
		}

		output, err := RunCommand(command, extension.Root, params, environ)
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

	r.Get("/{extension}/{command}", func(w http.ResponseWriter, r *http.Request) {
		params := make(map[string]any)
		for name, param := range r.URL.Query() {
			params[name] = param[0]
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

		environ := make(map[string]string)
		for _, env := range r.Header.Values("X-Sunbeam-Env") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				environ[parts[0]] = parts[1]
			}
		}

		output, err := RunCommand(command, extension.Root, params, environ)
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

func RunCommand(command app.Command, dir string, params map[string]any, env map[string]string) ([]byte, error) {
	cmd, err := command.Cmd(params, env, dir)
	if err != nil {
		return nil, fmt.Errorf("error running command: %s", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running command: %s", err)
	}

	return output, nil
}
