package tui

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
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Config struct {
	Aliases map[string]string `json:"aliases"`
	Root    map[string]string `json:"root"`
	Window  WindowOptions     `json:"-"`
}

type CommandInput struct {
	Command string         `json:"command"`
	Params  map[string]any `json:"params"`
	Inputs  map[string]any `json:"inputs,omitempty"`
	Query   string         `json:"query,omitempty"`
}

func (e Extension) Command(name string) (types.CommandSpec, bool) {
	for _, command := range e.Commands {
		if command.Name == name {
			return command, true
		}
	}
	return types.CommandSpec{}, false
}

func ShellCommand(ref CommandRef) string {
	args := []string{"sunbeam", "run", ref.Origin, ref.Command}
	for name, value := range ref.Params {
		switch value := value.(type) {
		case string:
			args = append(args, fmt.Sprintf("--%s=%s", name, value))
		case bool:
			if value {
				args = append(args, fmt.Sprintf("--%s", name))
			}
		}
	}

	return strings.Join(args, " ")
}

func (e Extension) Run(input CommandInput) ([]byte, error) {
	if input.Params == nil {
		input.Params = make(map[string]any)
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	switch e.Origin.Scheme {
	case "file":
		command := exec.Command(e.Origin.Path, string(inputBytes))
		command.Env = os.Environ()
		command.Env = append(command.Env, "NO_COLOR=1")

		var exitErr *exec.ExitError
		if output, err := command.Output(); err == nil {
			return output, nil
		} else if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
		} else {
			return nil, err
		}
	case "http", "https":
		var bearerToken string
		if e.Origin.User != nil {
			if _, ok := e.Origin.User.Password(); !ok {
				bearerToken = e.Origin.User.Username()
				e.Origin.User = nil
			}
		}

		req, err := http.NewRequest("POST", e.Origin.String(), bytes.NewReader(inputBytes))
		if err != nil {
			return nil, err
		}

		if bearerToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("command failed: %s", resp.Status)
		}

		return io.ReadAll(resp.Body)
	default:
		return nil, fmt.Errorf("unsupported origin scheme: %s", e.Origin.Scheme)
	}
}

func ParseOrigin(origin string) (*url.URL, error) {
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
		if strings.HasPrefix(url.Path, "~") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}

			url.Path = filepath.Join(homeDir, strings.TrimPrefix(url.Path, "~"))
		}

		abs, err := filepath.Abs(url.Path)
		if err != nil {
			return nil, err
		}

		url.Path = abs
	}

	return url, nil
}

type Extension struct {
	types.Manifest
	Origin *url.URL
}

type Extensions struct {
	Aliases    map[string]string
	Extensions map[string]Extension
}

func NewExtensions(aliases map[string]string) Extensions {
	return Extensions{
		Aliases:    aliases,
		Extensions: make(map[string]Extension),
	}
}

func (e Extensions) Get(rawOrigin string) (Extension, error) {
	origin, err := ParseOrigin(rawOrigin)
	if err != nil {
		return Extension{}, fmt.Errorf("invalid origin: %s", rawOrigin)
	}

	if extension, ok := e.Extensions[origin.String()]; ok {
		return extension, nil
	}

	extension, err := LoadExtension(origin)
	if err != nil {
		return extension, err
	}

	if e.Extensions == nil {
		e.Extensions = make(map[string]Extension)
	}
	e.Extensions[origin.String()] = extension

	return extension, nil
}

func LoadExtension(origin *url.URL) (Extension, error) {
	var extension Extension

	switch origin.Scheme {
	case "file":
		command := exec.Command(origin.Path)
		b, err := command.Output()
		if err != nil {
			return extension, err
		}

		if err := schemas.ValidateManifest(b); err != nil {
			return extension, err
		}

		var manifest types.Manifest
		if err := json.Unmarshal(b, &manifest); err != nil {
			return extension, err
		}

		return Extension{
			Manifest: manifest,
			Origin:   origin,
		}, nil
	case "http", "https":
		resp, err := http.Get(origin.String())
		if err != nil {
			return extension, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return extension, fmt.Errorf("failed to fetch extension manifest: %s", resp.Status)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return extension, err
		}

		if err := schemas.ValidateManifest(b); err != nil {
			return extension, err
		}

		var manifest types.Manifest
		if err := json.Unmarshal(b, &manifest); err != nil {
			return extension, err
		}

		return Extension{
			Manifest: manifest,
			Origin:   origin,
		}, nil
	default:
		return extension, fmt.Errorf("unsupported origin scheme: %s", origin.Scheme)
	}
}
