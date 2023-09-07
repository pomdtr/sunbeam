package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <origin>",
		Short:              "Run a command",
		GroupID:            coreGroupID,
		DisableFlagParsing: true,
		RunE: func(_ *cobra.Command, args []string) error {
			origin := args[0]
			var commandName string
			var manifest ExtensionManifest
			if strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://") {
				commandName = origin

				resp, err := http.Get(origin)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != 200 {
					return errors.New("manifest not found")
				}

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				err = json.Unmarshal(body, &manifest)
				if err != nil {
					return err
				}

				manifest.Entrypoint = []string{"sunbeam", "fetch", origin}
			} else {
				origin, err := filepath.Abs(origin)
				if err != nil {
					return err
				}

				info, err := os.Stat(origin)
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("manifest not found: %s", origin)
				} else if err != nil {
					return err
				}

				var manifestPath string
				if info.IsDir() {
					manifestPath = filepath.Join(origin, "sunbeam.json")
					if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
						return fmt.Errorf("manifest not found: %s", manifestPath)
					}

					commandName = filepath.Base(origin)
				} else if info.Name() == "sunbeam.json" {
					manifestPath = origin
					commandName = filepath.Base(filepath.Dir(origin))
				} else {
					return fmt.Errorf("manifest must be named sunbeam.json")
				}

				manifestBytes, err := os.ReadFile(manifestPath)
				if err != nil {
					return err
				}

				err = json.Unmarshal(manifestBytes, &manifest)
				if err != nil {
					return err
				}

			}

			cmd, err := NewCustomCmd(commandName, Extension{
				Manifest: manifest,
			})
			if err != nil {
				return err
			}

			cmd.SetArgs(args[1:])
			return cmd.Execute()
		},
	}

	return cmd
}
