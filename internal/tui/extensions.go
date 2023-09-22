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
	Extensions map[string]string `json:"extensions"`
	Items      []types.RootItem  `json:"items"`
	Window     WindowOptions     `json:"window"`
}

type Extension struct {
	Origin *url.URL
	Alias  string
	types.Manifest
}

type CommandInput struct {
	Query  string         `json:"query,omitempty"`
	Params map[string]any `json:"params"`
}

func (e Extension) Command(name string) (types.Command, bool) {
	for _, command := range e.Commands {
		if command.Name == name {
			return command, true
		}
	}

	return types.Command{}, false
}

func (ext Extension) Run(commandName string, input CommandInput) ([]byte, error) {
	if input.Params == nil {
		input.Params = make(map[string]any)
	}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	if ext.Origin.Scheme == "file" {
		command := exec.Command(ext.Origin.Path, commandName)
		command.Stdin = bytes.NewReader(inputBytes)
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
	}

	var bearerToken string
	if ext.Origin.User != nil {
		if _, ok := ext.Origin.User.Password(); !ok {
			bearerToken = ext.Origin.User.Username()
			ext.Origin.User = nil
		}
	}

	commandUrl, err := url.JoinPath(ext.Origin.String(), commandName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", commandUrl, bytes.NewReader(inputBytes))
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
}

type Extensions map[string]Extension

func (e Extensions) List() []string {
	aliases := make([]string, 0, len(e))
	for alias := range e {
		aliases = append(aliases, alias)
	}

	return aliases
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

func LoadExtension(originRaw string) (Extension, error) {
	origin, err := ParseOrigin(originRaw)
	if err != nil {
		return Extension{}, err
	}

	var manifest types.Manifest

	if origin.Scheme == "file" {
		command := exec.Command(origin.Path)
		b, err := command.Output()
		if err != nil {
			return Extension{}, err
		}

		if err := schemas.ValidateManifest(b); err != nil {
			return Extension{}, err
		}

		if err := json.Unmarshal(b, &manifest); err != nil {
			return Extension{}, err
		}
		return Extension{
			Origin:   origin,
			Manifest: manifest,
		}, nil

	}

	resp, err := http.Get(origin.String())
	if err != nil {
		return Extension{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Extension{}, fmt.Errorf("failed to fetch extension manifest: %s", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return Extension{}, err
	}

	if err := schemas.ValidateManifest(b); err != nil {
		return Extension{}, err
	}

	if err := json.Unmarshal(b, &manifest); err != nil {
		return Extension{}, err
	}

	return Extension{
		Origin:   origin,
		Manifest: manifest,
	}, nil
}
