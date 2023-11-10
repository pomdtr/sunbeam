package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/tailscale/hujson"
	"mvdan.cc/sh/shell"
)

type Config struct {
	Schema     string                  `json:"$schema,omitempty"`
	Oneliners  []Oneliner              `json:"oneliners,omitempty"`
	Extensions map[string]ExtensionRef `json:"extensions,omitempty"`
}

type ExtensionRef struct {
	Origin      string         `json:"origin"`
	Preferences map[string]any `json:"preferences,omitempty"`
}

func (e *ExtensionRef) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		e.Origin = s
		return nil
	}

	type Alias ExtensionRef
	var a Alias
	if err := json.Unmarshal(b, &a); err == nil {
		e.Origin = a.Origin
		e.Preferences = a.Preferences

		return nil
	}

	return fmt.Errorf("invalid extension ref: %s", string(b))
}

var DefaultConfig = Config{
	Schema: "https://github.com/pomdtr/sunbeam/releases/latest/download/config.schema.json",
	Oneliners: []Oneliner{
		{
			Title:   "Open Sunbeam Docs",
			Command: "sunbeam open https://pomdtr.github.io/sunbeam/introduction",
		},
		{
			Title:   "Edit Config",
			Command: "sunbeam edit --config",
		},
	},
	Extensions: map[string]ExtensionRef{
		"devdocs": {
			Origin: "https://raw.githubusercontent.com/pomdtr/sunbeam/main/extensions/devdocs.sh",
		},
		"google": {
			Origin: "https://raw.githubusercontent.com/pomdtr/sunbeam/main/extensions/google.sh",
		},
	},
}

type Oneliner struct {
	Title   string `json:"title"`
	Command string `json:"command"`
}

func Path() string {
	if env, ok := os.LookupEnv("SUNBEAM_CONFIG"); ok {
		return env
	}

	if _, err := os.Stat(filepath.Join(utils.ConfigHome(), "config.jsonc")); err == nil {
		return filepath.Join(utils.ConfigHome(), "config.jsonc")
	}

	return filepath.Join(utils.ConfigHome(), "config.json")

}

func Load() (Config, error) {
	configPath := Path()
	if _, err := os.Stat(configPath); err != nil {
		configBytes, err := json.MarshalIndent(DefaultConfig, "", "  ")
		if err != nil {
			return Config{}, err
		}

		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return Config{}, err
		}

		f, err := os.Create(configPath)
		if err != nil {
			return Config{}, err
		}
		defer f.Close()

		if _, err := f.Write(configBytes); err != nil {
			return Config{}, err
		}

		return DefaultConfig, nil
	}

	var configBytes []byte
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	if filepath.Ext(configPath) == ".jsonc" {
		bts, err := hujson.Standardize(configBytes)
		if err != nil {
			return Config{}, err
		}

		configBytes = bts
	}

	if err := schemas.ValidateConfig(configBytes); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func (c Config) RootItem(item Oneliner, extensionMap map[string]extensions.Extension) (types.ListItem, error) {
	// extract args from the command
	args, err := shell.Fields(item.Command, os.Getenv)
	if err != nil {
		return types.ListItem{
			Id:          fmt.Sprintf("root - %s", item.Title),
			Title:       item.Title,
			Accessories: []string{"Oneliner"},
			Actions: []types.Action{
				{
					Title: item.Title,
					Type:  types.ActionTypeExec,
					Args:  []string{"sh", "-c", item.Command},
					Exit:  true,
				},
				{
					Title: "Copy Command",
					Key:   "c",
					Type:  types.ActionTypeCopy,
					Text:  item.Command,
				},
			},
		}, nil
	}

	if len(args) == 0 {
		return types.ListItem{}, fmt.Errorf("invalid command: %s", item.Command)
	}

	if args[0] != "sunbeam" {
		return types.ListItem{
			Id:          fmt.Sprintf("root - %s", item.Title),
			Title:       item.Title,
			Accessories: []string{"Oneliner"},
			Actions: []types.Action{
				{
					Title: item.Title,
					Type:  types.ActionTypeExec,
					Args:  []string{"sh", "-c", item.Command},
					Exit:  true,
				},
				{
					Title: "Copy Command",
					Key:   "c",
					Type:  types.ActionTypeCopy,
					Text:  item.Command,
				},
			},
		}, nil
	}

	switch args[1] {
	case "open", "edit":
		return types.ListItem{
			Id:          fmt.Sprintf("root - %s", item.Title),
			Title:       item.Title,
			Accessories: []string{"Oneliner"},
			Actions: []types.Action{
				{
					Title: "Run",
					Type:  types.ActionTypeExec,
					Args:  args,
					Exit:  true,
				},
				{
					Title: "Copy Command",
					Key:   "c",
					Type:  types.ActionTypeCopy,
					Text:  item.Command,
				},
			},
		}, nil
	default:
		if len(args) < 3 {
			return types.ListItem{}, fmt.Errorf("invalid command: %s", item.Command)
		}

		alias := args[1]
		extension, ok := extensionMap[alias]
		if !ok {
			return types.ListItem{}, fmt.Errorf("extension %s not found", alias)
		}

		command, ok := extension.Command(args[2])
		if !ok {
			return types.ListItem{}, fmt.Errorf("command %s not found", args[2])
		}

		params, err := ExtractParams(args[3:], command)
		if err != nil {
			return types.ListItem{}, err
		}

		return types.ListItem{
			Id:          fmt.Sprintf("%s - %s", alias, item.Title),
			Title:       item.Title,
			Accessories: []string{extension.Title},
			Actions: []types.Action{
				{
					Title:     item.Title,
					Type:      types.ActionTypeRun,
					Extension: args[1],
					Command:   command.Name,
					Params:    params,
					Exit:      true,
				},
				{
					Title: "Copy Command",
					Key:   "c",
					Type:  types.ActionTypeCopy,
					Text:  item.Command,
				},
			},
		}, nil
	}
}

func ExtractParams(args []string, command types.CommandSpec) (map[string]any, error) {
	params := make(map[string]any)
	for len(args) > 0 {
		if !strings.HasPrefix(args[0], "--") {
			return nil, fmt.Errorf("invalid argument: %s", args[0])
		}

		parts := strings.SplitN(args[0][2:], "=", 2)
		if len(parts) == 1 {
			spec, ok := CommandParam(command, parts[0])
			if !ok {
				return nil, fmt.Errorf("unknown parameter: %s", parts[0])
			}

			switch spec.Type {
			case types.ParamTypeBoolean:
				params[parts[0]] = true
				args = args[1:]
			case types.ParamTypeString:
				if len(args) < 2 {
					return nil, fmt.Errorf("missing value for parameter: %s", parts[0])
				}

				params[parts[0]] = args[1]
				args = args[2:]
			case types.ParamTypeNumber:
				if len(args) < 2 {
					return nil, fmt.Errorf("missing value for parameter: %s", parts[0])
				}

				value, err := strconv.Atoi(args[1])
				if err != nil {
					return nil, err
				}

				params[parts[0]] = value
				args = args[2:]
			}

			continue
		}

		spec, ok := CommandParam(command, parts[0])
		if !ok {
			return nil, fmt.Errorf("unknown parameter: %s", parts[0])
		}

		switch spec.Type {
		case types.ParamTypeString:
			params[parts[0]] = parts[1]
		case types.ParamTypeBoolean:
			value, err := strconv.ParseBool(parts[1])
			if err != nil {
				return nil, err
			}
			params[parts[0]] = value
		case types.ParamTypeNumber:
			value, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, err
			}
			params[parts[0]] = value
		}

		args = args[1:]
	}

	return params, nil
}

func CommandParam(command types.CommandSpec, name string) (types.Param, bool) {
	for _, param := range command.Params {
		if param.Name == name {
			return param, true
		}
	}

	return types.Param{}, false
}
