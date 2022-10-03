package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pomdtr/sunbeam/commands"
)

func Serve(address string, port int) error {
	scripts, err := commands.ScanDir(commands.CommandDir)
	if err != nil {
		log.Fatalf("Error while scanning commands directory: %s", err)
	}

	routeMap := make(map[string]commands.ScriptMetadatas)
	for _, script := range scripts {
		route := Route(script)
		routeMap[route] = script.Metadatas
		log.Printf("Serving %s at %s", script.Url.Path, route)
		http.HandleFunc(Route(script), serveScript(script))
	}

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(res)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(routeMap)
	})

	log.Println("Starting server on", fmt.Sprintf("%s:%d", address, port))
	return http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), nil)
}

func Route(s commands.Script) string {
	return strings.TrimPrefix(s.Url.Path, commands.CommandDir)
}

func serveScript(s commands.Script) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		command := commands.Command{}
		command.Script = s
		_ = json.NewDecoder(req.Body).Decode(&command.CommandInput)

		scriptResponse, err := command.Run()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			_, _ = res.Write([]byte(err.Error()))
			return
		}

		res.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(res).Encode(scriptResponse)
	}
}
