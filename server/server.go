package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pomdtr/sunbeam/app"
)

func NewServer(extensions map[string]app.Extension, addr string) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: return a sunbeam.yml file
		http.Redirect(w, r, "/extensions", http.StatusFound)
	})

	r.Get("/extensions", func(w http.ResponseWriter, r *http.Request) {
		listItems := make([]app.ListItem, 0)
		for extensionName, extension := range extensions {
			for _, rootItem := range extension.RootItems {
				rootItem.Extension = extensionName
				listItems = append(listItems, app.ListItem{
					Id:          fmt.Sprintf("%s:%s", extensionName, rootItem.Title),
					Title:       rootItem.Title,
					Subtitle:    extension.Title,
					Accessories: []string{extensionName},
					Actions: []app.Action{
						{
							Title: "Run Command",
							Type:  "run-command",
							Url:   fmt.Sprintf("/extensions/%s/%s", extensionName, rootItem.Command),
						},
					},
				})
			}
		}

		page := app.Page{
			Type:  "list",
			Title: "Sunbeam",
			List: app.List{
				Items: listItems,
			},
		}

		// TODO: return a page with all extensions
		json.NewEncoder(w).Encode(page)
	})

	r.Get("/extensions/{extension}", func(w http.ResponseWriter, r *http.Request) {
		extensionName := chi.URLParam(r, "extension")
		extension, ok := extensions[extensionName]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Extension %s not found", extensionName)))
			return
		}

		listItems := make([]app.ListItem, 0)
		for _, rootItem := range extension.RootItems {
			rootItem.Extension = extensionName
			listItems = append(listItems, app.ListItem{
				Id:       rootItem.Title,
				Title:    rootItem.Title,
				Subtitle: extension.Title,
				Actions: []app.Action{
					{
						Title: "Run Command",
						Type:  "run-command",
						Url:   fmt.Sprintf("/extensions/%s/%s", extensionName, rootItem.Command),
					},
				},
			})
		}

		// TODO: return a page with only commands from the extension
		json.NewEncoder(w).Encode(app.Page{
			Type:  "list",
			Title: extension.Title,
			List: app.List{
				Items: listItems,
			},
		})
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
		command, ok := extension.GetCommand(commandName)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Command %s not found", commandName)))
			return
		}

		var input app.CommandPayload
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil && err != io.EOF {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error decoding input params: %s", err)))
			return
		}

		output, err := command.Run(input, extension.Root)
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
