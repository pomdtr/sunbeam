package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	bytes, _ := json.Marshal(Detail{
		Preview: "preview",
		Actions: []Action{
			{
				Title: "title",
				Type:  "copy-text",
				Text:  "text",
			},
		},
	})

	fmt.Println(string(bytes))
}
