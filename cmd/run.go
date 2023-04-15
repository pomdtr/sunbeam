package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

func NewCmdRun(extensionDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use: "run <page>",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			extension, err := ListExtensions(extensionDir)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			return extension, cobra.ShellCompDirectiveDefault
		},
		Short: "Run page from file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions, err := ListExtensions(extensionDir)
			if err != nil {
				return err
			}

			for _, extension := range extensions {
				if extension != args[0] {
					continue
				}
				bin := path.Join(extensionDir, extension, extensionBinaryName)
				command := exec.Command(bin, args[1:]...)
				return Draw(command.Output)
			}

			if repository, err := utils.RepositoryFromString(args[0]); err == nil {
				page := types.Page{
					Type: types.FormPage,
					SubmitAction: &types.Action{
						Type: types.RunAction,
						Command: &types.Command{
							Args: []string{os.Args[0], "extension", "install", "--open", repository.String()},
						},
						Inputs: []types.Input{
							{
								Type:  types.CheckboxInput,
								Title: "Confirm",
								Label: fmt.Sprintf("Install extension %s?", repository.FullName()),
							},
						},
					},
				}
				return Draw(func() ([]byte, error) {
					return json.Marshal(page)
				})
			}

			if _, err := os.Stat(args[0]); err == nil {
				command := exec.Command(args[0], args[1:]...)
				return Draw(command.Output)
			}

			if _, err := exec.LookPath(args[0]); err == nil {
				command := exec.Command(args[0], args[1:]...)
				return Draw(command.Output)
			}

			return fmt.Errorf("file or command not found: %s", args[0])
		},
	}

	cmd.Flags().String("on-success", "push", "action to trigger when the command is successful")

	return cmd
}
