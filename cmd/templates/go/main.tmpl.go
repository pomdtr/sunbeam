package main

import (
	"encoding/json"
	"os"

	sunbeam "github.com/pomdtr/sunbeam/types"
)

func main() {
	json.NewEncoder(os.Stdout).Encode(sunbeam.Page{
		Type: sunbeam.DetailPage,
		Preview: &sunbeam.Preview{
			Text: "Hello, World!",
		},
		Actions: []sunbeam.Action{
			sunbeam.NewCopyAction("Copy", "Hello, World!"),
		},
	})
}
