package main

import (
	"encoding/json"
	"os"
)

type ListItem struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

func main() {
	jsonEncoder := json.NewEncoder(os.Stdout)
	for _, item := range []ListItem{
		{Title: "Hello", Subtitle: "World"},
		{Title: "Goodbye", Subtitle: "World"},
	} {
		jsonEncoder.Encode(item)
	}
}
