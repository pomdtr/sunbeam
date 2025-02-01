package sunbeam

import (
	"encoding/json"
	"fmt"
)

type Action struct {
	Title string     `json:"title,omitempty"`
	Type  ActionType `json:"type,omitempty"`
	Exit  bool       `json:"exit,omitempty"`

	Open   *OpenAction   `json:"-"`
	Copy   *CopyAction   `json:"-"`
	Run    *RunAction    `json:"-"`
	Edit   *EditAction   `json:"-"`
	Reload *ReloadAction `json:"-"`
}

func (a *Action) UnmarshalJSON(bts []byte) error {
	var action struct {
		Title string `json:"title,omitempty"`
		Type  string `json:"type,omitempty"`
	}

	if err := json.Unmarshal(bts, &action); err != nil {
		return err
	}

	a.Title = action.Title
	a.Type = ActionType(action.Type)

	switch a.Type {
	case ActionTypeRun:
		a.Run = &RunAction{
			Params: map[string]any{},
		}
		return json.Unmarshal(bts, a.Run)
	case ActionTypeOpen:
		a.Open = &OpenAction{}
		return json.Unmarshal(bts, a.Open)
	case ActionTypeCopy:
		a.Copy = &CopyAction{}
		return json.Unmarshal(bts, a.Copy)
	case ActionTypeReload:
		a.Reload = &ReloadAction{
			Params: map[string]any{},
		}
		return json.Unmarshal(bts, a.Reload)
	}

	return nil
}

func (a Action) MarshalJSON() ([]byte, error) {
	switch a.Type {
	case ActionTypeRun:
		output := map[string]interface{}{
			"title":   a.Title,
			"type":    a.Type,
			"command": a.Run.Command,
		}

		if a.Run.Params != nil {
			output["params"] = a.Run.Params
		}

		if a.Run.Reload {
			output["reload"] = true
		}

		return json.Marshal(output)
	case ActionTypeOpen:
		return json.Marshal(map[string]interface{}{
			"title": a.Title,
			"type":  a.Type,
			"url":   a.Open.Url,
		})

	case ActionTypeCopy:
		return json.Marshal(map[string]interface{}{
			"title": a.Title,
			"type":  a.Type,
			"text":  a.Copy.Text,
		})
	case ActionTypeReload:
		output := map[string]interface{}{
			"title": a.Title,
			"type":  a.Type,
		}

		if a.Reload.Params != nil {
			output["params"] = a.Reload.Params
		}

		return json.Marshal(output)
	}

	return nil, fmt.Errorf("unknown action type: %s", a.Type)
}

type EditAction struct {
	Path   string `json:"path,omitempty"`
	Reload bool   `json:"reload,omitempty"`
}

type ReloadAction struct {
	Params map[string]any `json:"params,omitempty"`
}

type RunAction struct {
	Extension string         `json:"-"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
	Reload    bool           `json:"reload,omitempty"`
}

type CopyAction struct {
	Text string `json:"text,omitempty"`
}

type ExecAction struct {
	Interactive bool   `json:"interactive,omitempty"`
	Command     string `json:"command,omitempty"`
	Dir         string `json:"dir,omitempty"`
}

type OpenAction struct {
	Url string `json:"url,omitempty"`
}

type ActionType string

const (
	ActionTypeRun    ActionType = "run"
	ActionTypeOpen   ActionType = "open"
	ActionTypeCopy   ActionType = "copy"
	ActionTypeEdit   ActionType = "edit"
	ActionTypeReload ActionType = "reload"
)

type Params map[string]any
