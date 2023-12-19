package extensions

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/schemas"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
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
	Manifest   sunbeam.Manifest
	Entrypoint string `json:"entrypoint"`
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

func (e Extension) Command(name string) (sunbeam.CommandSpec, bool) {
	for _, command := range e.Manifest.Commands {
		if command.Name == name {
			return command, true
		}
	}
	return sunbeam.CommandSpec{}, false
}

func (e Extension) RootCommands() []sunbeam.CommandSpec {
	rootCommands := make([]sunbeam.CommandSpec, 0)
	for _, command := range e.Manifest.Commands {
		if command.Hidden {
			continue
		}

		rootCommands = append(rootCommands, command)
	}

	return rootCommands
}

func (e Extension) Run(input sunbeam.Payload) error {
	_, err := e.Output(input)
	return err
}

func (ext Extension) Output(input sunbeam.Payload) ([]byte, error) {
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

func (e Extension) Cmd(input sunbeam.Payload) (*exec.Cmd, error) {
	return e.CmdContext(context.Background(), input)
}

func (e Extension) CmdContext(ctx context.Context, input sunbeam.Payload) (*exec.Cmd, error) {
	if input.Preferences == nil {
		input.Preferences = make(map[string]any)
	}

	for _, spec := range e.Manifest.Preferences {
		if _, ok := input.Preferences[spec.Name]; ok {
			continue
		}

		if !spec.Optional {
			return nil, fmt.Errorf("missing required preference %s", spec.Name)
		}

		input.Preferences[spec.Name] = spec.Default
	}

	command, ok := e.Command(input.Command)
	if !ok {
		return nil, fmt.Errorf("command %s not found", input.Command)
	}

	if input.Params == nil {
		input.Params = make(map[string]any)
	}

	for _, spec := range command.Params {
		if _, ok := input.Params[spec.Name]; ok {
			continue
		}

		if !spec.Optional {
			return nil, fmt.Errorf("missing required parameter %s", spec.Name)
		}

		input.Params[spec.Name] = spec.Default
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	input.Cwd = cwd

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

func Hash(origin string) (string, error) {
	if !IsRemote(origin) {
		abs, err := filepath.Abs(origin)
		if err != nil {
			return "", err
		}

		origin = abs
	}

	h := sha1.New()
	h.Write([]byte(origin))
	return hex.EncodeToString(h.Sum(nil)), nil
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

	return nil
}

func LoadEntrypoint(origin string, extensionDir string) (string, error) {
	if IsRemote(origin) {
		originUrl, err := url.Parse(origin)
		if err != nil {
			return "", fmt.Errorf("failed to parse origin: %w", err)
		}

		entrypoint := filepath.Join(extensionDir, filepath.Base(originUrl.Path))
		if _, err := os.Stat(entrypoint); err == nil {
			return entrypoint, nil
		}

		if err := os.MkdirAll(extensionDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		if err := DownloadEntrypoint(origin, entrypoint); err != nil {
			return "", err
		}

		if err := os.Chmod(entrypoint, 0755); err != nil {
			return "", fmt.Errorf("failed to chmod entrypoint: %w", err)
		}

		return entrypoint, nil
	}

	entrypoint := origin
	if strings.HasPrefix(entrypoint, "~") {
		entrypoint = strings.Replace(entrypoint, "~", os.Getenv("HOME"), 1)
	} else if !filepath.IsAbs(entrypoint) {
		entrypoint = filepath.Join(filepath.Dir(config.Path), entrypoint)
	}

	return filepath.Abs(entrypoint)
}

func LoadExtension(origin string) (Extension, error) {
	hash, err := Hash(origin)
	if err != nil {
		return Extension{}, err
	}
	extensionDir := filepath.Join(utils.CacheDir(), "extensions", hash)
	entrypoint, err := LoadEntrypoint(origin, extensionDir)
	if err != nil {
		return Extension{}, err
	}

	entrypointInfo, err := os.Stat(entrypoint)
	if err != nil {
		return Extension{}, err
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
		}, nil
	}

	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return Extension{}, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest sunbeam.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return Extension{}, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return Extension{
		Manifest:   manifest,
		Entrypoint: entrypoint,
	}, nil
}

func cacheManifest(entrypoint string, manifestPath string) (sunbeam.Manifest, error) {
	manifest, err := ExtractManifest(entrypoint)
	if err != nil {
		return sunbeam.Manifest{}, fmt.Errorf("failed to extract manifest: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return sunbeam.Manifest{}, fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(manifestPath)
	if err != nil {
		return sunbeam.Manifest{}, fmt.Errorf("failed to create manifest: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(manifest); err != nil {
		return sunbeam.Manifest{}, fmt.Errorf("failed to write manifest: %w", err)
	}

	return manifest, nil
}

func Upgrade(extensionConfig config.ExtensionConfig) error {
	hash, err := Hash(extensionConfig.Origin)
	if err != nil {
		return err
	}

	extensionDir := filepath.Join(utils.CacheDir(), "extensions", hash)
	manifestPath := filepath.Join(extensionDir, "manifest.json")
	if IsRemote(extensionConfig.Origin) {
		originUrl, err := url.Parse(extensionConfig.Origin)
		if err != nil {
			return fmt.Errorf("failed to parse origin: %w", err)
		}

		entrypoint := filepath.Join(extensionDir, filepath.Base(originUrl.Path))
		if err := DownloadEntrypoint(extensionConfig.Origin, entrypoint); err != nil {
			return err
		}

		if _, err := cacheManifest(entrypoint, manifestPath); err != nil {
			return err
		}

		return nil
	}

	entrypoint := extensionConfig.Origin
	if strings.HasPrefix(entrypoint, "~") {
		entrypoint = strings.Replace(entrypoint, "~", os.Getenv("HOME"), 1)
	} else if !filepath.IsAbs(entrypoint) {
		entrypoint = filepath.Join(filepath.Dir(config.Path), entrypoint)
	}

	if _, err := cacheManifest(entrypoint, manifestPath); err != nil {
		return err
	}
	return nil
}

func ExtractManifest(entrypoint string) (sunbeam.Manifest, error) {
	entrypoint, err := filepath.Abs(entrypoint)
	if err != nil {
		return sunbeam.Manifest{}, err
	}

	if err := os.Chmod(entrypoint, 0755); err != nil {
		return sunbeam.Manifest{}, err
	}

	cmd := exec.Command(entrypoint)
	cmd.Dir = filepath.Dir(entrypoint)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SUNBEAM=1")

	manifestBytes, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return sunbeam.Manifest{}, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
		}

		return sunbeam.Manifest{}, err
	}

	if err := schemas.ValidateManifest(manifestBytes); err != nil {
		return sunbeam.Manifest{}, err
	}

	var manifest sunbeam.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return sunbeam.Manifest{}, err
	}

	return manifest, nil
}
