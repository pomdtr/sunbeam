package server

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pomdtr/sunbeam/commands"
)

func Serve(address string, port int) error {
	scripts, err := commands.ScanDir(commands.CommandDir)
	router := mux.NewRouter().UseEncodedPath().StrictSlash(true)
	if err != nil {
		log.Fatalf("Error while scanning commands directory: %s", err)
	}

	routeMap := make(map[string]commands.ScriptMetadatas)
	for _, script := range scripts {
		route := Route(script)
		routeMap[route] = script.Metadatas
		log.Printf("Serving %s at %s", script.Url.Path, route)
		router.HandleFunc(Route(script), serveScript(script))
	}

	router.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(res)
		encoder.SetIndent("", "  ")
		encoder.Encode(routeMap)
	})

	log.Println("Starting server on", fmt.Sprintf("%s:%d", address, port))
	return http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), router)
}

func Route(s commands.Script) string {
	route := strings.Builder{}
	prefix := strings.TrimPrefix(s.Url.Path, commands.CommandDir)
	if prefix[0] != '/' {
		prefix = "/" + prefix
	}
	route.WriteString(fmt.Sprintf("%s", prefix))
	if s.Metadatas.Argument1 != nil {
		route.WriteString(fmt.Sprintf("/{%s}", s.Metadatas.Argument1.Name))
	}
	if s.Metadatas.Argument2 != nil {
		route.WriteString(fmt.Sprintf("/{%s}", s.Metadatas.Argument2.Name))
	}
	if s.Metadatas.Argument3 != nil {
		route.WriteString(fmt.Sprintf("/{%s}", s.Metadatas.Argument3.Name))
	}

	return route.String()
}

func serveScript(s commands.Script) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		args := make([]string, 0)
		vars := mux.Vars(req)
		if s.Metadatas.Argument1 != nil {
			arg, _ := url.QueryUnescape(vars[s.Metadatas.Argument1.Name])
			args = append(args, arg)
		}
		if s.Metadatas.Argument2 != nil {
			args = append(args, vars[s.Metadatas.Argument1.Name])
		}
		if s.Metadatas.Argument3 != nil {
			args = append(args, vars[s.Metadatas.Argument1.Name])
		}

		for key, value := range req.URL.Query() {
			args = append(args, fmt.Sprintf("--%s=%s", key, html.UnescapeString(value[0])))
		}

		command := commands.Command{}
		command.Script = s
		command.Args = args
		json.NewDecoder(req.Body).Decode(&command.Input)

		scriptResponse, err := command.Run()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(err.Error()))
			return
		}

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(scriptResponse)
	}
}
