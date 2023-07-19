package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

type ExitCodeError struct {
	ExitCode int
}

func (e ExitCodeError) Error() string {
	return fmt.Sprintf("exit code %d", e.ExitCode)
}

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run <origin> [args...]",
		Short:   "read a sunbeam command",
		GroupID: coreGroupID,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input string
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				bs, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				input = string(bs)
			}

			if originUrl, err := url.Parse(args[0]); err == nil && originUrl.Scheme == "https" {
				resp, err := http.Get(args[0])
				if err != nil {
					return fmt.Errorf("could not fetch script: %w", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != 200 {
					return fmt.Errorf("could not fetch script: %s", resp.Status)
				}

				tmpDir, err := os.MkdirTemp("", "sunbeam")
				if err != nil {
					return err
				}
				defer os.RemoveAll(tmpDir)

				scriptPath := filepath.Join(tmpDir, "sunbeam-command")
				f, err := os.OpenFile(scriptPath, os.O_CREATE|os.O_WRONLY, 0755)
				if err != nil {
					return err
				}
				defer f.Close()

				if _, err := io.Copy(f, resp.Body); err != nil {
					return err
				}
				if err := f.Close(); err != nil {
					return err
				}

				return runCommand(types.Command{
					Name:  scriptPath,
					Args:  args[1:],
					Input: input,
				})
			}

			root := args[0]
			if info, err := os.Stat(root); err != nil {
				return fmt.Errorf("could not find script: %w", err)
			} else if !info.IsDir() {
				root = filepath.Join(root, "sunbeam.json")
			}

			var entrypoint string
			if filepath.Base(root) == "sunbeam.json" {
				f, err := os.Open(root)
				if err != nil {
					return err
				}
				defer f.Close()

				var metadata Metadata
				if err := json.NewDecoder(f).Decode(&metadata); err != nil {
					return err
				}
				entrypoint = filepath.Join(filepath.Dir(root), metadata.Entrypoint)
			} else {
				entrypoint = root
			}

			return runCommand(types.Command{
				Name:  entrypoint,
				Args:  args[1:],
				Input: input,
			})
		},
	}

	return cmd
}
