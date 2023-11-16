package extensions

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type ExtensionMap map[string]Extension

func (e ExtensionMap) List() []Extension {
	extensions := make([]Extension, 0)
	for _, extension := range e {
		extensions = append(extensions, extension)
	}
	return extensions
}

type Extension struct {
	Manifest   types.Manifest
	Entrypoint string `json:"entrypoint"`
	Config     Config `json:"config"`
}

type Config struct {
	Origin      string            `json:"origin,omitempty"`
	Preferences types.Preferences `json:"preferences,omitempty"`
	Items       []types.RootItem  `json:"items,omitempty"`
	Hooks       Hooks             `json:"hooks,omitempty"`
}

func (c Config) ExtensionDir() string {
	return filepath.Join(utils.CacheHome(), "extensions", SHA1(c.Origin))
}

func (c Config) Entrypoint() (string, error) {
	if IsRemoteExtension(c.Origin) {
		return filepath.Join(c.ExtensionDir(), "entrypoint"), nil
	}

	entrypoint := c.Origin
	if strings.HasPrefix(entrypoint, "~") {
		entrypoint = strings.Replace(entrypoint, "~", os.Getenv("HOME"), 1)
	}

	entrypoint, err := filepath.Abs(entrypoint)
	if err != nil {
		return "", err
	}

	return entrypoint, nil
}

type Hooks struct {
	Install string `json:"install,omitempty"`
	Upgrade string `json:"upgrade,omitempty"`
}

func (e *Config) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		e.Origin = s
		return nil
	}

	type Alias Config

	var alias Alias
	if err := json.Unmarshal(b, &alias); err == nil {
		e.Origin = alias.Origin
		e.Preferences = alias.Preferences
		e.Items = alias.Items
		e.Hooks = alias.Hooks
		return nil
	}

	return fmt.Errorf("invalid extension ref: %s", string(b))
}

type Preferences map[string]any

type Metadata struct {
	Type       ExtensionType `json:"type"`
	Origin     string        `json:"origin"`
	Entrypoint string        `json:"entrypoint"`
}

type ExtensionType string

const (
	ExtensionTypeLocal ExtensionType = "local"
	ExtensionTypeHttp  ExtensionType = "http"
)

func (e Extension) Command(name string) (types.CommandSpec, bool) {
	for _, command := range e.Manifest.Commands {
		if command.Name == name {
			return command, true
		}
	}
	return types.CommandSpec{}, false
}

func (e Extension) Root() []types.RootItem {
	rootItems := make([]types.RootItem, 0)
	var items []types.RootItem
	items = append(items, e.Manifest.Items...)
	items = append(items, e.Config.Items...)

	for _, rootItem := range items {
		command, ok := e.Command(rootItem.Command)
		if !ok {
			continue
		}

		if rootItem.Title == "" {
			rootItem.Title = command.Title
		}

		rootItems = append(rootItems, rootItem)
	}

	return rootItems
}

func (e Extension) Run(input types.Payload) error {
	_, err := e.Output(input)
	return err
}

func (ext Extension) CheckRequirements() error {
	for _, requirement := range ext.Manifest.Requirements {
		if _, err := exec.LookPath(requirement.Name); err != nil {
			return fmt.Errorf("missing requirement %s", requirement.Name)
		}
	}

	return nil
}

func (ext Extension) Output(input types.Payload) ([]byte, error) {
	cmd, err := ext.Cmd(input)
	if err != nil {
		return nil, err
	}

	var exitErr *exec.ExitError
	if output, err := cmd.Output(); err == nil {
		return output, nil
	} else if errors.As(err, &exitErr) {
		return nil, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
	} else {
		return nil, err
	}
}

func (e Extension) Cmd(input types.Payload) (*exec.Cmd, error) {
	return e.CmdContext(context.Background(), input)
}

func (e Extension) CmdContext(ctx context.Context, input types.Payload) (*exec.Cmd, error) {
	if input.Params == nil {
		input.Params = make(map[string]any)
	}

	input.Preferences = e.Config.Preferences
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	input.Cwd = cwd

	command, ok := e.Command(input.Command)
	if !ok {
		return nil, fmt.Errorf("command %s not found", input.Command)
	}

	for _, spec := range command.Inputs {
		if !spec.Required {
			if spec.Default != nil {
				input.Params[spec.Name] = spec.Default
			}

			continue
		}
		_, ok := input.Params[spec.Name]
		if !ok {
			return nil, fmt.Errorf("missing required parameter %s", spec.Name)
		}
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, e.Entrypoint, string(inputBytes))
	cmd.Dir = filepath.Dir(e.Entrypoint)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SUNBEAM=1")
	return cmd, nil
}

func SHA1(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func IsRemoteExtension(origin string) bool {
	return strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://")
}

func LoadExtension(config Config) (Extension, error) {
	entrypoint, err := config.Entrypoint()
	if err != nil {
		return Extension{}, err
	}

	entrypointInfo, err := os.Stat(entrypoint)
	if err != nil {
		return InstallExtension(config)
	}

	manifestPath := filepath.Join(config.ExtensionDir(), "manifest.json")
	manifestInfo, err := os.Stat(manifestPath)
	if err != nil {
		return InstallExtension(config)
	}

	if entrypointInfo.ModTime().After(manifestInfo.ModTime()) {
		manifest, err := ExtractManifest(entrypoint)
		if err != nil {
			return Extension{}, err
		}

		f, err := os.Create(manifestPath)
		if err != nil {
			return Extension{}, err
		}
		defer f.Close()

		if err := json.NewEncoder(f).Encode(manifest); err != nil {
			return Extension{}, err
		}

		return Extension{
			Manifest:   manifest,
			Entrypoint: entrypoint,
			Config:     config,
		}, nil
	}

	var manifest types.Manifest
	f, err := os.Open(manifestPath)
	if err != nil {
		return Extension{}, err
	}

	if err := json.NewDecoder(f).Decode(&manifest); err != nil {
		return Extension{}, err
	}

	return Extension{
		Manifest:   manifest,
		Entrypoint: entrypoint,
		Config:     config,
	}, nil
}

func InstallExtension(config Config) (Extension, error) {
	extensionDir := config.ExtensionDir()
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return Extension{}, err
	}

	entrypoint, err := config.Entrypoint()
	if err != nil {
		return Extension{}, err
	}

	if config.Hooks.Install != "" {
		cmd := exec.Command("sh", "-c", config.Hooks.Install)
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return Extension{}, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
			}
			return Extension{}, err
		}
	} else if IsRemoteExtension(config.Origin) {
		resp, err := http.Get(config.Origin)
		if err != nil {
			return Extension{}, err
		}

		if resp.StatusCode != http.StatusOK {
			return Extension{}, fmt.Errorf("error downloading extension: %s", resp.Status)
		}

		f, err := os.Create(entrypoint)
		if err != nil {
			return Extension{}, err
		}

		if _, err := io.Copy(f, resp.Body); err != nil {
			return Extension{}, err
		}

		if err := f.Close(); err != nil {
			return Extension{}, err
		}
	}

	if err := os.Chmod(entrypoint, 0755); err != nil {
		return Extension{}, err
	}

	manifestPath := filepath.Join(config.ExtensionDir(), "manifest.json")
	manifest, err := ExtractManifest(entrypoint)
	if err != nil {
		return Extension{}, err
	}

	f, err := os.Create(manifestPath)
	if err != nil {
		return Extension{}, err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(manifest); err != nil {
		return Extension{}, err
	}

	if err := f.Close(); err != nil {
		return Extension{}, err
	}

	return Extension{
		Manifest:   manifest,
		Entrypoint: entrypoint,
		Config:     config,
	}, nil
}

func UpgradeExtension(config Config) (Extension, error) {
	entrypoint, err := config.Entrypoint()
	if err != nil {
		return Extension{}, err
	}

	if config.Hooks.Upgrade != "" {
		cmd := exec.Command("sh", "-c", config.Hooks.Upgrade)
		if err := cmd.Run(); err != nil {
			return Extension{}, fmt.Errorf("failed to run install hook: %s", err)
		}
	} else if IsRemoteExtension(config.Origin) {
		resp, err := http.Get(config.Origin)
		if err != nil {
			return Extension{}, err
		}

		if resp.StatusCode != http.StatusOK {
			return Extension{}, fmt.Errorf("error downloading extension: %s", resp.Status)
		}

		f, err := os.OpenFile(entrypoint, os.O_TRUNC|os.O_WRONLY, 0755)
		if err != nil {
			return Extension{}, err
		}

		if _, err := io.Copy(f, resp.Body); err != nil {
			return Extension{}, err
		}
	}

	manifest, err := ExtractManifest(entrypoint)
	if err != nil {
		return Extension{}, err
	}

	manifestPath := filepath.Join(config.ExtensionDir(), "manifest.json")
	f, err := os.OpenFile(manifestPath, os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return Extension{}, err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(manifest); err != nil {
		return Extension{}, err
	}

	if err := f.Close(); err != nil {
		return Extension{}, err
	}

	return Extension{
		Manifest:   manifest,
		Entrypoint: entrypoint,
		Config:     config,
	}, nil

}

func ExtractManifest(entrypoint string) (types.Manifest, error) {
	if err := os.Chmod(entrypoint, 0755); err != nil {
		return types.Manifest{}, err
	}

	cmd := exec.Command(entrypoint)
	cmd.Dir = filepath.Dir(entrypoint)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SUNBEAM=1")

	manifestBytes, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return types.Manifest{}, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
		}

		return types.Manifest{}, err
	}

	if err := schemas.ValidateManifest(manifestBytes); err != nil {
		return types.Manifest{}, err
	}

	var manifest types.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return types.Manifest{}, err
	}

	return manifest, nil
}
