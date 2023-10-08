package types

import (
	"encoding/json"
	"fmt"
)

type CommandRef struct {
	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
}

type Command struct {
	Type CommandType `json:"type,omitempty"`

	Text string `json:"text,omitempty"`

	App    Applications `json:"app,omitempty"`
	Target string       `json:"target,omitempty"`

	Exit bool `json:"exit,omitempty"`

	Reload bool `json:"reload,omitempty"`

	CommandRef

	Exec string `json:"exec,omitempty"`
}

type CommandType string

const (
	CommandTypeRun    CommandType = "run"
	CommandTypeOpen   CommandType = "open"
	CommandTypeCopy   CommandType = "copy"
	CommandTypeReload CommandType = "reload"
	CommandTypeExit   CommandType = "exit"
	CommandTypePop    CommandType = "pop"
	CommandTypePass   CommandType = "pass"
)

type Applications []Application

type Application struct {
	Name     string   `json:"name"`
	Platform Platform `json:"platform"`
}

type Platform string

func (a *Applications) UnmarshalJSON(b []byte) error {
	var app Application
	if err := json.Unmarshal(b, &app); err == nil {
		*a = []Application{app}
		return nil
	}

	var apps []Application
	if err := json.Unmarshal(b, &apps); err == nil {
		*a = apps
		return nil
	}

	return fmt.Errorf("invalid application")
}

var (
	PlatformWindows = "windows"
	PlatformMac     = "mac"
	PlatformLinux   = "linux"
)

type CommandInput struct {
	Params   map[string]any `json:"params"`
	FormData map[string]any `json:"formData,omitempty"`
	Query    string         `json:"query,omitempty"`
}
