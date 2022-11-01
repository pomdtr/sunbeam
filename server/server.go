package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pomdtr/sunbeam/api"
)

func Serve(address string, port int) error {
	for _, manifest := range api.Sunbeam.Extensions {
		for route, command := range manifest.Scripts {
			http.HandleFunc(route, serveCommand(command))
		}
	}

	log.Println("Starting server on", fmt.Sprintf("%s:%d", address, port))
	return http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), nil)
}

func serveCommand(cmd api.Script) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var input api.ScriptInput
		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			_, _ = res.Write([]byte(err.Error()))
			return
		}

		// scriptResponse, err := cmd.Run(input)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			_, _ = res.Write([]byte(err.Error()))
			return
		}

		res.WriteHeader(http.StatusOK)
		// _ = json.NewEncoder(res).Encode(scriptResponse)
	}
}
