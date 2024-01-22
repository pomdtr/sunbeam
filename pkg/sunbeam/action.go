package sunbeam

import (
	"encoding/json"
)

type Action struct {
	Title string     `json:"title,omitempty"`
	Key   string     `json:"key,omitempty"`
	Type  ActionType `json:"type,omitempty"`

	Open   *OpenAction   `json:"-"`
	Copy   *CopyAction   `json:"-"`
	Run    *RunAction    `json:"-"`
	Exec   *ExecAction   `json:"-"`
	Edit   *EditAction   `json:"-"`
	Config *ConfigAction `json:"-"`
	Reload *ReloadAction `json:"-"`
}

func (a *Action) UnmarshalJSON(bts []byte) error {
	var action struct {
		Title string `json:"title,omitempty"`
		Key   string `json:"key,omitempty"`
		Type  string `json:"type,omitempty"`
	}

	if err := json.Unmarshal(bts, &action); err != nil {
		return err
	}

	a.Title = action.Title
	a.Key = action.Key
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
	case ActionTypeExec:
		a.Exec = &ExecAction{}
		return json.Unmarshal(bts, a.Exec)
	case ActionTypeExit:
		return nil
	case ActionTypeReload:
		a.Reload = &ReloadAction{}
		return json.Unmarshal(bts, a.Reload)
	case ActionTypeConfig:
		a.Config = &ConfigAction{}
		return json.Unmarshal(bts, a.Config)
	}

	return nil
}

type ConfigAction struct {
	Extension string `json:"extension,omitempty"`
}

type EditAction struct {
	Path   string `json:"path,omitempty"`
	Exit   bool   `json:"exit,omitempty"`
	Reload bool   `json:"reload,omitempty"`
}

type ReloadAction struct {
	Params map[string]any `json:"params,omitempty"`
}

type RunAction struct {
	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
	Reload    bool           `json:"reload,omitempty"`
	Exit      bool           `json:"exit,omitempty"`
}

type CopyAction struct {
	Text string `json:"text,omitempty"`
	Exit bool   `json:"exit,omitempty"`
}

type ExecAction struct {
	Interactive bool   `json:"interactive,omitempty"`
	Command     string `json:"command,omitempty"`
	Dir         string `json:"dir,omitempty"`
	Exit        bool   `json:"exit,omitempty"`
}

type OpenAction struct {
	Url  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
}

type ActionType string

const (
	ActionTypeRun    ActionType = "run"
	ActionTypeOpen   ActionType = "open"
	ActionTypeCopy   ActionType = "copy"
	ActionTypeEdit   ActionType = "edit"
	ActionTypeExec   ActionType = "exec"
	ActionTypeExit   ActionType = "exit"
	ActionTypeReload ActionType = "reload"
	ActionTypeConfig ActionType = "config"
)

type Payload struct {
	Command     string         `json:"command"`
	Preferences map[string]any `json:"preferences"`
	Params      map[string]any `json:"params"`
	Cwd         string         `json:"cwd"`
	Query       string         `json:"query,omitempty"`
}
