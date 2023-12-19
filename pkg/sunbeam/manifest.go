package sunbeam

import (
	"fmt"
)

type Manifest struct {
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Preferences []Input       `json:"preferences,omitempty"`
	Commands    []CommandSpec `json:"commands"`
}

type CommandSpec struct {
	Name   string      `json:"name"`
	Title  string      `json:"title"`
	Hidden bool        `json:"hidden,omitempty"`
	Params []Input     `json:"params,omitempty"`
	Mode   CommandMode `json:"mode,omitempty"`
}

type Platfom string

const (
	PlatformLinux Platfom = "linux"
	PlatformMac   Platfom = "macos"
)

type Requirement struct {
	Name string `json:"name"`
	Link string `json:"link,omitempty"`
}

type CommandMode string

const (
	CommandModeSearch CommandMode = "search"
	CommandModeFilter CommandMode = "filter"
	CommandModeDetail CommandMode = "detail"
	CommandModeTTY    CommandMode = "tty"
	CommandModeSilent CommandMode = "silent"
)

type InputType string

const (
	InputText     InputType = "text"
	InputTextArea InputType = "textarea"
	InputPassword InputType = "password"
	InputCheckbox InputType = "checkbox"
	InputNumber   InputType = "number"
)

type Input struct {
	Type     InputType `json:"type"`
	Name     string    `json:"name"`
	Label    string    `json:"label"`
	Optional bool      `json:"optional,omitempty"`

	TextInput     *TextInput     `json:"text,omitempty"`
	PasswordInput *PasswordInput `json:"password,omitempty"`
	TextAreaInput *TextAreaInput `json:"textarea,omitempty"`
	NumberInput   *NumberInput   `json:"number,omitempty"`
	Checkbox      *CheckboxInput `json:"checkbox,omitempty"`
}

func (i Input) Default() any {
	switch i.Type {
	case InputText:
		return i.TextInput.Default
	case InputPassword:
		return i.PasswordInput.Default
	case InputTextArea:
		return i.TextAreaInput.Default
	case InputNumber:
		return i.NumberInput.Default
	case InputCheckbox:
		return i.Checkbox.Default
	default:
		return nil
	}
}

func (i Input) SetDefault(v any) error {
	switch i.Type {
	case InputText:
		defaultValue, ok := v.(string)
		if !ok {
			return fmt.Errorf("invalid default value: %v", v)
		}
		i.TextInput.Default = defaultValue
		return nil
	case InputPassword:
		defaultValue, ok := v.(string)
		if !ok {
			return fmt.Errorf("invalid default value: %v", v)
		}
		i.PasswordInput.Default = defaultValue
		return nil
	case InputTextArea:
		defaultValue, ok := v.(string)
		if !ok {
			return fmt.Errorf("invalid default value: %v", v)
		}
		i.TextAreaInput.Default = defaultValue
		return nil
	case InputNumber:
		defaultValue, ok := v.(int)
		if !ok {
			return fmt.Errorf("invalid default value: %v", v)
		}
		i.NumberInput.Default = defaultValue
		return nil
	case InputCheckbox:
		defaultValue, ok := v.(bool)
		if !ok {
			return fmt.Errorf("invalid default value: %v", v)
		}
		i.Checkbox.Default = defaultValue
		return nil
	default:
		return fmt.Errorf("invalid input type: %s", i.Type)
	}

}

type TextInput struct {
	Placeholder string `json:"placeholder,omitempty"`
	Default     string `json:"default,omitempty"`
}

type PasswordInput struct {
	Placeholder string `json:"placeholder,omitempty"`
	Default     string `json:"default,omitempty"`
}

type TextAreaInput struct {
	Placeholder string `json:"placeholder,omitempty"`
	Default     string `json:"default,omitempty"`
}

type NumberInput struct {
	Placeholder string `json:"placeholder,omitempty"`
	Default     int    `json:"default,omitempty"`
}

type CheckboxInput struct {
	Label   string `json:"label,omitempty"`
	Default bool   `json:"default,omitempty"`
}
