package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true).UseEncodedPath()
	router.NewRoute().Name("test").Methods("Post").Path("/test/{name}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, mux.Vars(r))
	})
	http.ListenAndServe(":8081", router)
	// defer server.Close()

	// res, err := http.Get(server.URL + "/test/test%2Ftest")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// greeting, err := ioutil.ReadAll(res.Body)
	// res.Body.Close()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%s", greeting)
}
