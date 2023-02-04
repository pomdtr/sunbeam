package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewCmdRun(config *tui.Config) *cobra.Command {
	runCmd := &cobra.Command{
		Use:     "run <extension-root>",
		Short:   "Run an extension from a directory",
		GroupID: "core",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionRoot := args[0]
			if fi, err := os.Stat(extensionRoot); err != nil || !fi.IsDir() {
				return remoteRun(extensionRoot)
			}

			return localRun(extensionRoot)
		},
	}

	return runCmd
}

func localRun(extensionRoot string) error {
	manifestPath := filepath.Join(extensionRoot, "sunbeam.yml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("directory %s is not a sunbeam extension", extensionRoot)
	}

	extension, err := app.ParseManifest(manifestPath)
	extension.Root = extensionRoot
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %s", err)
	}

	rootList := tui.NewRootList(&extension)
	model := tui.NewModel(rootList)

	return tui.Draw(model)
}

func remoteRun(rawUrl string) (err error) {
	resp, err := http.Get(rawUrl)
	if err != nil {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch extension: %s", resp.Status)
	}

	var extension app.Extension
	if err := json.NewDecoder(resp.Body).Decode(&extension); err != nil {
		return err
	}

	for i, command := range extension.Commands {
		command.Exec = buildExec(command, rawUrl)
		extension.Commands[i] = command
	}

	tempdir, err := os.MkdirTemp("", "sunbeam")
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(tempdir, "sunbeam.yml")
	file, err := os.OpenFile(manifestPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if err := yaml.NewEncoder(file).Encode(extension); err != nil {
		return err
	}

	extension.Root = tempdir
	list := tui.NewRootList(&extension)
	model := tui.NewModel(list)
	return tui.Draw(model)
}

func buildExec(command app.Command, extensionUrl string) string {
	args := []string{"sunbeam", "http", "--ignore-stdin", "POST", fmt.Sprintf("%s/%s", extensionUrl, command.Name)}

	for _, param := range command.Params {
		args = append(args, fmt.Sprintf("%s=${{%s}}", param.Name, param.Name))
	}

	for _, env := range command.Env {
		args = append(args, fmt.Sprintf(`"X-Sunbeam-Env:%s=$%s"`, env.Name, env.Name))
	}

	return strings.Join(args, " ")
}
