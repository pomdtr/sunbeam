package cmd

import (
	"fmt"
	"io"
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

				if err := os.Chmod(filepath.Join(cwd, extensionBinaryName), 0755); err != nil {
					return err
				}

				os.Setenv("SUNBEAM_KV_PATH", filepath.Join(cwd, ".sunbeam", kvFile))
				os.Setenv("SUNBEAM_EXTENSION_BIN", filepath.Join(cwd, extensionBinaryName))

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

			return fmt.Errorf("file or command not found: %s", args[0])
		},
	}

	return cmd
}
