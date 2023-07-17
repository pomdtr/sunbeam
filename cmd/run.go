package cmd

import (
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

			scriptDir := args[0]
			if info, err := os.Stat(scriptDir); err != nil {
				return fmt.Errorf("could not find script: %w", err)
			} else if !info.IsDir() {
				base := filepath.Base(scriptDir)
				if base != "sunbeam.json" && base != "sunbeam-command" {
					return fmt.Errorf("not a sunbeam script")
				}

				scriptDir = filepath.Dir(scriptDir)
			}

			command, err := ExtractDirMetadata(scriptDir)
			if err != nil {
				return fmt.Errorf("could not parse command: %w", err)
			}

			return runCommand(types.Command{
				Name:  filepath.Join(scriptDir, command.Entrypoint),
				Args:  args[1:],
				Input: input,
			})
		},
	}

	return cmd
}
