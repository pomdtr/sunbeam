package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

type Extension struct {
	Origin   string            `json:"origin"`
	Manifest ExtensionManifest `json:"manifest"`
}

type ExtensionManifest struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Origin      string     `json:"origin,omitempty"`
	Entrypoint  Entrypoint `json:"entrypoint,omitempty"`
	Commands    []Command  `json:"commands"`
}

type Entrypoint []string

func (e *Entrypoint) UnmarshalJSON(b []byte) error {
	var entrypoint string
	if err := json.Unmarshal(b, &entrypoint); err == nil {
		*e = Entrypoint{entrypoint}
		return nil
	}

	var entrypoints []string
	if err := json.Unmarshal(b, &entrypoints); err == nil {
		*e = Entrypoint(entrypoints)
		return nil
	}

	return fmt.Errorf("invalid entrypoint: %s", string(b))
}

type Command struct {
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Mode        CommandMode     `json:"mode"`
	Hidden      bool            `json:"hidden,omitempty"`
	Description string          `json:"description,omitempty"`
	Params      []CommandParams `json:"params,omitempty"`
}

type CommandMode string

const (
	CommandModeList      CommandMode = "filter"
	CommandModeGenerator CommandMode = "generator"
	CommandModeDetail    CommandMode = "detail"
	CommandModeText      CommandMode = "text"
	CommandModeSilent    CommandMode = "silent"
)

type CommandParams struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Default     any       `json:"default,omitempty"`
	Optional    bool      `json:"optional,omitempty"`
	Description string    `json:"description,omitempty"`
}

type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeBoolean ParamType = "boolean"
)

type CommandInput struct {
	Command string         `json:"command"`
	Query   string         `json:"query"`
	Params  map[string]any `json:"params"`
}

func (ext Extension) Run(input CommandInput) ([]byte, error) {
	origin, err := url.Parse(ext.Origin)
	if err != nil {
		return nil, err
	}
	if origin.Scheme == "file" {
		inputBytes, err := json.Marshal(input.Params)
		if err != nil {
			return nil, err
		}

		command := exec.Command(origin.Path, input.Command)
		command.Stdin = bytes.NewReader(inputBytes)
		command.Env = os.Environ()
		command.Env = append(command.Env, fmt.Sprintf("SUNBEAM_QUERY=%s", input.Query))

		var exitErr *exec.ExitError
		if output, err := command.Output(); err != nil && errors.As(err, &exitErr) {
			return output, fmt.Errorf("command failed: %s", string(exitErr.Stderr))
		} else {
			return output, err
		}
	}
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(ext.Origin, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("command failed: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)

}

type ExtensionMap map[string]Extension

func LoadExtensions() (ExtensionMap, error) {
	extensions := make(ExtensionMap)
	f, err := os.Open(filepath.Join(dataDir, "extensions.json"))
	if os.IsNotExist(err) {
		return extensions, nil
	}

	if err != nil {
		return extensions, err
	}

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&extensions); err != nil {
		return extensions, err
	}

	return extensions, nil
}

func (e ExtensionMap) Save() error {
	f, err := os.Create(filepath.Join(dataDir, "extensions.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	return encoder.Encode(e)
}

func LoadManifest(origin *url.URL) (ExtensionManifest, error) {
	var manifest ExtensionManifest

	if origin.Scheme == "file" {
		command := exec.Command(origin.Path)
		b, err := command.Output()
		if err != nil {
			return manifest, err
		}

		if err := json.Unmarshal(b, &manifest); err != nil {
			return manifest, err
		}
		return manifest, nil

	}

	resp, err := http.Get(origin.String())
	if err != nil {
		return manifest, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return manifest, fmt.Errorf("failed to fetch extension manifest: %s", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return manifest, err
	}

	if err := json.Unmarshal(b, &manifest); err != nil {
		return manifest, err
	}

	return manifest, nil

}

func NewExtensionCmd(extensionMap ExtensionMap) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension",
		Short:   "Manage extensions",
		GroupID: coreGroupID,
	}

	cmd.AddCommand(NewExtensionAddCmd(extensionMap))

	return cmd
}

func parseOrigin(origin string) (*url.URL, error) {
	url, err := url.Parse(origin)
	if err != nil {
		return nil, err
	}

	if url.Scheme == "" {
		url.Scheme = "file"
	}

	if url.Scheme != "file" && url.Scheme != "http" && url.Scheme != "https" {
		return nil, fmt.Errorf("invalid origin: %s", origin)
	}

	if url.Scheme == "file" && !filepath.IsAbs(url.Path) {
		abs, err := filepath.Abs(url.Path)
		if err != nil {
			return nil, err
		}

		url.Path = abs
	}

	return url, nil
}

func NewExtensionAddCmd(extensionMap ExtensionMap) *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <origin>",
		Short: "Add an extension",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := extensionMap[args[0]]; ok {
				return fmt.Errorf("extension %s already exists", args[0])
			}

			return nil
		},
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			originUrl, err := parseOrigin(args[1])
			if err != nil {
				return err
			}

			manifest, err := LoadManifest(originUrl)
			if err != nil {
				return err
			}

			extensionMap[name] = Extension{
				Origin:   originUrl.String(),
				Manifest: manifest,
			}

			if err := extensionMap.Save(); err != nil {
				return err
			}

			return nil
		},
	}
}

func NewExtensionCmdRemove(extensions ExtensionMap) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an extension",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := extensions[args[0]]; !ok {
				return fmt.Errorf("extension %s does not exist", args[0])
			}

			return nil
		},
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			delete(extensions, name)

			if err := extensions.Save(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func NewExtensionUpdateCmd(extensions ExtensionMap) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update an extension",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := extensions[args[0]]; !ok {
				return fmt.Errorf("extension %s does not exist", args[0])
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			extension := extensions[name]

			origin, err := parseOrigin(extension.Origin)
			if err != nil {
				return err
			}

			manifest, err := LoadManifest(origin)
			if err != nil {
				return err
			}

			extension.Manifest = manifest
			extensions[name] = extension

			if err := extensions.Save(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func NewExtensionRenameCmd(extensions ExtensionMap) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename an extension",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := extensions[args[0]]; !ok {
				return fmt.Errorf("extension %s does not exist", args[0])
			}

			if _, ok := extensions[args[1]]; ok {
				return fmt.Errorf("extension %s already exists", args[1])
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName := args[0]
			newName := args[1]

			extension := extensions[oldName]
			delete(extensions, oldName)
			extensions[newName] = extension

			if err := extensions.Save(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
