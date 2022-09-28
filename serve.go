package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pomdtr/sunbeam/commands"
)

func serve() error {
	scripts, err := commands.ScanDir(commands.CommandDir)
	if err != nil {
		log.Fatalf("Error while scanning commands directory: %s", err)
	}
	router := mux.NewRouter()
	for _, script := range scripts {
		router.HandleFunc(Route(script), serveScript(script))
	}

	router.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("OK"))
	})

	return http.ListenAndServe(":8001", router)
}

func Route(s commands.Script) string {
	route := strings.Builder{}
	prefix := strings.TrimPrefix(s.Path, commands.CommandDir)
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
		vars := mux.Vars(req)

		args := make([]string, 0)
		if s.Metadatas.Argument1 != nil {
			args = append(args, vars[s.Metadatas.Argument1.Name])
		}
		if s.Metadatas.Argument2 != nil {
			args = append(args, vars[s.Metadatas.Argument1.Name])
		}
		if s.Metadatas.Argument3 != nil {
			args = append(args, vars[s.Metadatas.Argument1.Name])
		}
		for key, value := range req.URL.Query() {
			args = append(args, fmt.Sprintf("--%s=%s", key, value[0]))
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
