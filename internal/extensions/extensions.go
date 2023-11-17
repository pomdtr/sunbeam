package extensions

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	Install     string            `json:"install,omitempty"`
	Upgrade     string            `json:"upgrade,omitempty"`
	Items       []types.RootItem  `json:"items,omitempty"`
}

func ExtensionsDir() string {
	if env, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		return filepath.Join(env, "sunbeam", "extensions")
	}

	return filepath.Join(os.Getenv("HOME"), ".cache", "sunbeam", "extensions")
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
	if input.Preferences == nil {
		input.Preferences = e.Manifest.Preferences
	}

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

func IsRemote(origin string) bool {
	return strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://")
}

func DownloadEntrypoint(origin string, target string) error {
	resp, err := http.Get(origin)
	if err != nil {
		return fmt.Errorf("failed to download extension: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download extension: %s", resp.Status)
	}

	f, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create entrypoint: %w", err)
	}

	if _, err := f.ReadFrom(resp.Body); err != nil {
		return fmt.Errorf("failed to write entrypoint: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close entrypoint: %w", err)
	}

	if err := os.Chmod(target, 0755); err != nil {
		return fmt.Errorf("failed to chmod entrypoint: %w", err)
	}

	return nil
}

func LoadRemoteExtension(config Config) (Extension, error) {
	extensionDir := filepath.Join(ExtensionsDir(), SHA1(config.Origin))

	entrypoint := filepath.Join(extensionDir, "entrypoint")
	entrypointInfo, err := os.Stat(entrypoint)
	if err != nil {
		if err := os.MkdirAll(extensionDir, 0755); err != nil {
			return Extension{}, fmt.Errorf("failed to create directory: %w", err)
		}
		if err := DownloadEntrypoint(config.Origin, entrypoint); err != nil {
			return Extension{}, err
		}
	}

	manifestPath := filepath.Join(extensionDir, "manifest.json")
	manifestInfo, err := os.Stat(manifestPath)
	if err != nil || entrypointInfo.ModTime().After(manifestInfo.ModTime()) {
		manifest, err := cacheManifest(entrypoint, manifestPath)
		if err != nil {
			return Extension{}, err
		}

		return Extension{
			Manifest:   manifest,
			Entrypoint: entrypoint,
			Config:     config,
		}, nil
	}

	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return Extension{}, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest types.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return Extension{}, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return Extension{
		Manifest:   manifest,
		Entrypoint: entrypoint,
		Config:     config,
	}, nil
}

func LoadLocalExtension(config Config) (Extension, error) {
	entrypoint := config.Origin
	if strings.HasPrefix(entrypoint, "~") {
		entrypoint = strings.Replace(entrypoint, "~", os.Getenv("HOME"), 1)
	} else if !filepath.IsAbs(entrypoint) {
		entrypoint = filepath.Join(utils.ConfigDir(), entrypoint)
	}

	entrypoint, err := filepath.Abs(entrypoint)
	if err != nil {
		return Extension{}, fmt.Errorf("failed to get absolute path: %w", err)
	}

	entrypointInfo, err := os.Stat(entrypoint)
	if err != nil && config.Install != "" {
		if err := exec.Command("sh", "-c", config.Install).Run(); err != nil {
			return Extension{}, fmt.Errorf("failed to install extension: %w", err)
		}

		info, err := os.Stat(entrypoint)
		if err != nil {
			return Extension{}, fmt.Errorf("failed to stat entrypoint: %w", err)
		}
		entrypointInfo = info
	}

	sha := SHA1(config.Origin)
	manifestPath := filepath.Join(ExtensionsDir(), sha, "manifest.json")
	manifestInfo, err := os.Stat(manifestPath)
	if err != nil || entrypointInfo.ModTime().After(manifestInfo.ModTime()) {
		manifest, err := cacheManifest(entrypoint, manifestPath)
		if err != nil {
			return Extension{}, err
		}

		return Extension{
			Manifest:   manifest,
			Entrypoint: entrypoint,
			Config:     config,
		}, nil
	}

	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return Extension{}, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest types.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return Extension{}, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return Extension{
		Manifest:   manifest,
		Entrypoint: entrypoint,
		Config:     config,
	}, nil
}

func LoadExtension(config Config) (Extension, error) {
	if IsRemote(config.Origin) {
		return LoadRemoteExtension(config)
	}

	return LoadLocalExtension(config)
}

func cacheManifest(entrypoint string, manifestPath string) (types.Manifest, error) {
	manifest, err := ExtractManifest(entrypoint)
	if err != nil {
		return types.Manifest{}, fmt.Errorf("failed to extract manifest: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return types.Manifest{}, fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(manifestPath)
	if err != nil {
		return types.Manifest{}, fmt.Errorf("failed to create manifest: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(manifest); err != nil {
		return types.Manifest{}, fmt.Errorf("failed to write manifest: %w", err)
	}

	return manifest, nil
}

func Upgrade(extensionConfig Config) error {
	manifestPath := filepath.Join(ExtensionsDir(), SHA1(extensionConfig.Origin), "manifest.json")
	if IsRemote(extensionConfig.Origin) {
		entrypoint := filepath.Join(ExtensionsDir(), SHA1(extensionConfig.Origin), "entrypoint")
		if err := DownloadEntrypoint(extensionConfig.Origin, entrypoint); err != nil {
			return err
		}

		if _, err := cacheManifest(entrypoint, manifestPath); err != nil {
			return err
		}

		return nil
	}

	if extensionConfig.Upgrade != "" {
		if err := exec.Command("sh", "-c", extensionConfig.Upgrade).Run(); err != nil {
			return fmt.Errorf("failed to upgrade extension: %w", err)
		}
	} else if extensionConfig.Install != "" {
		if err := exec.Command("sh", "-c", extensionConfig.Install).Run(); err != nil {
			return fmt.Errorf("failed to install extension: %w", err)
		}
	}

	entrypoint := extensionConfig.Origin
	if strings.HasPrefix(entrypoint, "~") {
		entrypoint = strings.Replace(entrypoint, "~", os.Getenv("HOME"), 1)
	} else if !filepath.IsAbs(entrypoint) {
		entrypoint = filepath.Join(utils.ConfigDir(), entrypoint)
	}

	if _, err := cacheManifest(entrypoint, manifestPath); err != nil {
		return err
	}
	return nil
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
