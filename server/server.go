package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pomdtr/sunbeam/app"
	"gopkg.in/yaml.v3"
)

func NewServer(extensions []*app.Extension, addr string) *http.Server {
	extensionMap := make(map[string]app.Extension)
	for _, extension := range extensions {
		extensionMap[extension.Name()] = *extension
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
			command.Exec = buildExec(command, extension, extractUrl(r))
			commands[i] = command
		}

		w.Header().Set("Content-Type", "text/yaml")
		yaml.NewEncoder(w).Encode(app.Extension{
			Version:     extension.Version,
			Preferences: extension.Preferences,
			Title:       extension.Title,
			Description: extension.Description,
			RootItems:   extension.RootItems,
			Commands:    commands,
		})
	})

	r.Post("/{extension}/{command}", func(w http.ResponseWriter, r *http.Request) {
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

		var payload app.CmdPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error decoding payload: %s", err)))
			return
		}

		cmd, err := command.Cmd(payload, extension.Root)
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

func extractUrl(r *http.Request) *url.URL {
	extensionUrl := url.URL{
		Path: r.URL.Path,
	}

	if r.TLS != nil {
		extensionUrl.Scheme = "https"
	} else {
		extensionUrl.Scheme = "http"
	}

	extensionUrl.Host = r.Host

	return &extensionUrl
}

func buildExec(command app.Command, extension app.Extension, extensionUrl *url.URL) string {
	commandUrl := url.URL{
		Scheme: extensionUrl.Scheme,
		Host:   extensionUrl.Host,
		Path:   path.Join(extensionUrl.Path, command.Name),
	}
	args := []string{"curl", "-X", "POST", "-H", "Content-Type:application/json", commandUrl.String(), "-d", "{{ payload }}"}
	return strings.Join(args, " ")
}
