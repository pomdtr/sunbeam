package server

import (
	"encoding/json"
	"fmt"
	"html"
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
		encoder.Encode(routeMap)
	})

	log.Println("Starting server on", fmt.Sprintf("%s:%d", address, port))
	return http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), nil)
}

func Route(s commands.Script) string {
	route := strings.Builder{}
	prefix := strings.TrimPrefix(s.Url.Path, commands.CommandDir)
	if prefix[0] != '/' {
		prefix = "/" + prefix
	}
	route.WriteString(fmt.Sprintf("%s", prefix))

	return route.String()
}

func serveScript(s commands.Script) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		params := make([]string, 0)

		// Add required arguments
		if args, ok := req.URL.Query()["_"]; ok {
			args := strings.Split(",", args[0])
			for arg := range args {
				params = append(params, html.UnescapeString(args[arg]))
			}
		}

		// Add options
		for key, value := range req.URL.Query() {
			params = append(params, fmt.Sprintf("--%s=%s", key, html.UnescapeString(value[0])))
		}

		command := commands.Command{}
		command.Script = s
		command.Args = params
		json.NewDecoder(req.Body).Decode(&command.Input)

		scriptResponse := command.Run()

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(scriptResponse)
	}
}
