package sunbeam

import (
	"encoding/json"
	"fmt"
)

type Action struct {
	Title string     `json:"title,omitempty"`
	Key   string     `json:"key,omitempty"`
	Type  ActionType `json:"type,omitempty"`

	Open   *OpenAction   `json:"open,omitempty"`
	Copy   *CopyAction   `json:"copy,omitempty"`
	Run    *RunAction    `json:"run,omitempty"`
	Exec   *ExecAction   `json:"exec,omitempty"`
	Edit   *EditAction   `json:"edit,omitempty"`
	Config *ConfigAction `json:"config,omitempty"`
	Reload *ReloadAction `json:"reload,omitempty"`
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
	Params map[string]Param `json:"params,omitempty"`
}

type RunAction struct {
	Extension string           `json:"extension,omitempty"`
	Command   string           `json:"command,omitempty"`
	Params    map[string]Param `json:"params,omitempty"`
	Reload    bool             `json:"reload,omitempty"`
	Exit      bool             `json:"exit,omitempty"`
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

type Param struct {
	Value   any `json:"value,omitempty"`
	Default any `json:"default,omitempty"`
}

func (p *Param) UnmarshalJSON(bts []byte) error {
	var s string
	if err := json.Unmarshal(bts, &s); err == nil {
		p.Value = s
		return nil
	}

	var b bool
	if err := json.Unmarshal(bts, &b); err == nil {
		p.Value = b
		return nil
	}

	var n int
	if err := json.Unmarshal(bts, &n); err == nil {
		p.Value = n
		return nil
	}

	var param struct {
		Default  any  `json:"default,omitempty"`
		Optional bool `json:"optional,omitempty"`
	}

	if err := json.Unmarshal(bts, &param); err == nil {
		p.Default = param.Default
		return nil
	}

	return fmt.Errorf("invalid param: %s", string(bts))
}

func (p Param) MarshalJSON() ([]byte, error) {
	if p.Value != nil {
		return json.Marshal(p.Value)
	}

	return json.Marshal(struct {
		Default  any  `json:"default,omitempty"`
		Required bool `json:"required,omitempty"`
	}{
		Default: p.Default,
	})
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
