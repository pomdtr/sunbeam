package api

import (
	"strings"
)

type ListItem struct {
	Icon     string         `json:"icon"`
	Title    string         `json:"title"`
	Subtitle string         `json:"subtitle"`
	Detail   DetailData     `json:"detail"`
	Fill     string         `json:"fill"`
	Actions  []ScriptAction `json:"actions"`
}

type ScriptAction struct {
	Type     string            `json:"type"`
	RawTitle string            `json:"title"`
	Path     string            `json:"path"`
	Keybind  string            `json:"keybind"`
	Params   map[string]string `json:"params"`
	Target   string            `json:"target,omitempty"`
	CommandData
	Application string `json:"application,omitempty"`
	Url         string `json:"url,omitempty"`
	Content     string `json:"content,omitempty"`
}

func (a ScriptAction) Title() string {
	if a.RawTitle != "" {
		return a.RawTitle
	}

	return strings.ToTitle(a.Type)
}
