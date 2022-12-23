package main

type Action struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Text  string `json:"text"`
}

type Detail struct {
	Preview string   `json:"preview"`
	Actions []Action `json:"actions"`
}
