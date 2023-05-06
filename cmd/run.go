package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewCmdRun(extensionDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Generate a page from a command or a script, and push it's output",
		GroupID: coreGroupID,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			extension, err := ListExtensions(extensionDir)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			completions := make([]string, 0, len(extension))
			for extension := range extension {
				completions = append(completions, extension)
			}

			return completions, cobra.ShellCompDirectiveDefault
		},
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input string
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				input = string(b)
			}

			if args[0] == "." {
				if _, err := os.Stat(extensionBinaryName); err != nil {
					return fmt.Errorf("no extension found in current directory")
				}

				cwd, err := os.Getwd()
				if err != nil {
					return err
				}

				return runExtension(filepath.Join(cwd, extensionBinaryName), args[1:], input)
			}

			if strings.HasPrefix(args[0], ".") {
				if _, err := os.Stat(args[0]); err != nil {
					return fmt.Errorf("%s: no such file or directory", args[0])
				}

				return Draw(internal.NewCommandGenerator(&types.Command{
					Name:  args[0],
					Args:  args[1:],
					Input: input,
				}))
			}

			if _, err := exec.LookPath(args[0]); err == nil {
				return Draw(internal.NewCommandGenerator(&types.Command{
					Name:  args[0],
					Args:  args[1:],
					Input: input,
				}))
			}

			url, err := url.Parse(args[0])
			if err == nil && url.Host == "gist.github.com" {
				rawUrl := fmt.Sprintf("%s/raw/sunbeam-extension", url.String())

				tempfile, err := os.CreateTemp("", "sunbeam-extension")
				if err != nil {
					return err
				}
				defer os.Remove(tempfile.Name())

				res, err := http.Get(rawUrl)
				if err != nil {
					return err
				}
				defer res.Body.Close()

				body, err := io.ReadAll(res.Body)
				if err != nil {
					return err
				}

				if _, err := tempfile.Write(body); err != nil {
					return err
				}

				if err := os.Chmod(tempfile.Name(), 0755); err != nil {
					return err
				}

				return Draw(internal.NewCommandGenerator(&types.Command{
					Name:  tempfile.Name(),
					Args:  args[1:],
					Input: input,
				}))
			}

			return fmt.Errorf("file or command not found: %s", args[0])
		},
	}

	return cmd
}
