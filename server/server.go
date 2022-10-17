package server

import (
	"encoding/json"
	"net/http"

	"github.com/pomdtr/sunbeam/scripts"
)

// func Serve(address string, port int) error {
// 	scriptList, err := scripts.ScanDir(scripts.CommandDir)
// 	if err != nil {
// 		log.Fatalf("Error while scanning scripts directory: %s", err)
// 	}

// 	routeMap := make(map[string]string)
// 	for _, script := range scriptList {
// 		route := Route(script)
// 		routeMap[route] = script.Title()
// 		log.Printf("Serving %s at %s", script.Path(), route)
// 		http.HandleFunc(Route(script), serveScript(script))
// 	}

// 	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
// 		res.WriteHeader(http.StatusOK)
// 		encoder := json.NewEncoder(res)
// 		encoder.SetIndent("", "  ")
// 		_ = encoder.Encode(routeMap)
// 	})

// 	log.Println("Starting server on", fmt.Sprintf("%s:%d", address, port))
// 	return http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), nil)
// }

// func Route(s scripts.Command) string {
// 	return strings.TrimPrefix(s.Path(), scripts.CommandDir)
// }

func serveScript(cmd scripts.Command) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var input scripts.CommandInput
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
