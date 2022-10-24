package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pomdtr/sunbeam/api"
)

func Serve(address string, port int) error {
	routeMap := make(map[string]string)
	for _, command := range api.Commands {
		http.HandleFunc(command.Id, serveCommand(command))
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

func serveCommand(cmd api.Command) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var input api.CommandInput
		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			_, _ = res.Write([]byte(err.Error()))
			return
		}

		scriptResponse, err := cmd.Run(input)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			_, _ = res.Write([]byte(err.Error()))
			return
		}

		res.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(res).Encode(scriptResponse)
	}
}
