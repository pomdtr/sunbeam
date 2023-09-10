package internal

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

	"github.com/acarl005/stripansi"
	"github.com/pomdtr/sunbeam/pkg"
)

type Extension struct {
	Origin string
	pkg.Manifest
}

func (ext Extension) Run(commandName string, input pkg.CommandInput) ([]byte, error) {

	origin, err := url.Parse(ext.Origin)
	if err != nil {
		return nil, err
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	if origin.Scheme == "file" {
		command := exec.Command(origin.Path, commandName)
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
	if origin.User != nil {
		if _, ok := origin.User.Password(); !ok {
			bearerToken = origin.User.Username()
			origin.User = nil
		}
	}

	commandUrl, err := url.JoinPath(origin.String(), commandName)
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

func (e Extension) Command(name string) (pkg.Command, bool) {
	for _, command := range e.Commands {
		if command.Name == name {
			return command, true
		}
	}

	return pkg.Command{}, false
}

type Extensions struct {
	filepath  string                  `json:"-"`
	Commands  map[string]string       `json:"commands"`
	Manifests map[string]pkg.Manifest `json:"manifests"`
}

func (ext Extensions) Get(name string) (Extension, error) {
	var extension Extension
	origin, ok := ext.Commands[name]
	if !ok {
		return extension, fmt.Errorf("extension %s does not exist", name)
	}
	extension.Origin = origin

	manifest, ok := ext.Manifests[origin]
	if !ok {
		return extension, fmt.Errorf("manifest for extension %s does not exist", name)
	}
	extension.Manifest = manifest

	return extension, nil
}

func (ext Extensions) Update(name string, extension Extension) error {
	if _, ok := ext.Commands[name]; !ok {
		return fmt.Errorf("extension %s does not exist", name)
	}

	ext.Commands[name] = extension.Origin
	ext.Manifests[extension.Origin] = extension.Manifest

	return nil
}

func (ext Extensions) Add(name string, extension Extension) error {
	if _, ok := ext.Commands[name]; ok {
		return fmt.Errorf("extension %s already exists", name)
	}

	ext.Commands[name] = extension.Origin
	if _, ok := ext.Manifests[extension.Origin]; !ok {
		ext.Manifests[extension.Origin] = extension.Manifest
	}

	return nil
}

func (ext Extensions) Remove(name string) error {
	origin, ok := ext.Commands[name]
	if !ok {
		return fmt.Errorf("extension %s does not exist", name)
	}
	delete(ext.Commands, name)

	for _, commandOrigin := range ext.Commands {
		if commandOrigin == origin {
			return nil
		}
	}

	delete(ext.Manifests, origin)
	return nil
}

func (ext Extensions) Rename(oldName, newName string) error {
	if _, ok := ext.Commands[oldName]; !ok {
		return fmt.Errorf("extension %s does not exist", oldName)
	}

	if _, ok := ext.Commands[newName]; ok {
		return fmt.Errorf("extension %s already exists", newName)
	}

	ext.Commands[newName] = ext.Commands[oldName]
	delete(ext.Commands, oldName)

	return nil
}

func (ext Extensions) List() []string {
	var names []string
	for name := range ext.Commands {
		names = append(names, name)
	}

	return names
}

func (ext Extensions) Map() map[string]Extension {
	var extensions = make(map[string]Extension)
	for name, origin := range ext.Commands {
		extensions[name] = Extension{
			Origin:   origin,
			Manifest: ext.Manifests[origin],
		}
	}

	return extensions
}

func LoadExtensions(manifestPath string) (Extensions, error) {
	extensions := Extensions{
		filepath:  manifestPath,
		Commands:  make(map[string]string),
		Manifests: make(map[string]pkg.Manifest),
	}

	f, err := os.Open(manifestPath)
	if os.IsNotExist(err) {
		return extensions, nil
	} else if err != nil {
		return extensions, err
	}

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&extensions); err != nil {
		return extensions, err
	}

	return extensions, nil
}

func (e Extensions) Save() error {
	if err := os.MkdirAll(filepath.Dir(e.filepath), 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(e.filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(e)
}

func LoadManifest(origin *url.URL) (pkg.Manifest, error) {
	var manifest pkg.Manifest

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
