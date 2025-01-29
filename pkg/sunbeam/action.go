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
		a.Run = &RunAction{}
		return json.Unmarshal(bts, a.Run)
	case ActionTypeOpen:
		a.Open = &OpenAction{}
		return json.Unmarshal(bts, a.Open)
	case ActionTypeCopy:
		a.Copy = &CopyAction{}
		return json.Unmarshal(bts, a.Copy)
	case ActionTypeEdit:
		a.Edit = &EditAction{}
		return json.Unmarshal(bts, a.Edit)
	case ActionTypeReload:
		a.Reload = &ReloadAction{}
		return json.Unmarshal(bts, a.Reload)
	}

	return nil
}

func (a Action) MarshalJSON() ([]byte, error) {
	switch a.Type {
	case ActionTypeRun:
		return json.Marshal(map[string]interface{}{
			"title":   a.Title,
			"type":    a.Type,
			"command": a.Run.Command,
			"params":  a.Run.Params,
			"reload":  a.Run.Reload,
		})
	case ActionTypeOpen:
		return json.Marshal(map[string]interface{}{
			"title":  a.Title,
			"type":   a.Type,
			"target": a.Open.Target,
		})

	case ActionTypeCopy:
		return json.Marshal(map[string]interface{}{
			"title": a.Title,
			"type":  a.Type,
			"text":  a.Copy.Text,
		})
	case ActionTypeEdit:
		return json.Marshal(map[string]interface{}{
			"title":  a.Title,
			"type":   a.Type,
			"path":   a.Edit.Path,
			"reload": a.Edit.Reload,
		})
	case ActionTypeReload:
		return json.Marshal(map[string]interface{}{
			"title":  a.Title,
			"type":   a.Type,
			"params": a.Reload.Params,
		})
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
	Target string `json:"target,omitempty"`
}

type ActionType string

const (
	ActionTypeRun    ActionType = "run"
	ActionTypeOpen   ActionType = "open"
	ActionTypeCopy   ActionType = "copy"
	ActionTypeEdit   ActionType = "edit"
	ActionTypeReload ActionType = "reload"
)

type Payload struct {
	Command string         `json:"command"`
	Params  map[string]any `json:"params"`
	Cwd     string         `json:"cwd"`
	Query   string         `json:"query,omitempty"`
}
