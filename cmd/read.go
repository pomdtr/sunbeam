package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run <command> [args...]",
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

			info, err := os.Stat(args[0])
			if err != nil {
				return err
			}

			scriptPath := args[0]
			if info.IsDir() {
				scriptPath = filepath.Join(scriptPath, manifestName)
			}

			if filepath.Base(scriptPath) != manifestName {
				return runCommand(types.Command{
					Name:  scriptPath,
					Args:  args[1:],
					Input: input,
				})
			}

			command, err := ParseCommand(filepath.Dir(scriptPath), "")
			if err != nil {
				return fmt.Errorf("could not parse command: %w", err)
			}

			if command.Entrypoint != "" {
				return runCommand(types.Command{
					Name:  filepath.Join(filepath.Dir(scriptPath), command.Entrypoint),
					Args:  args[1:],
					Input: input,
				})
			}

			if len(args) < 2 {
				var listitems []types.ListItem
				for name, command := range command.SubCommands {
					listitems = append(listitems, types.ListItem{
						Title:       command.Title,
						Subtitle:    command.Description,
						Accessories: []string{name},
						Actions: []types.Action{
							{
								Title: "Run",
								Type:  types.PushAction,
								Command: &types.Command{
									Name: filepath.Join(filepath.Dir(scriptPath), command.Entrypoint),
									Args: args,
								},
							},
						},
					})
				}

				return Run(func() (*types.Page, error) {
					return &types.Page{
						Type:  types.ListPage,
						Title: "Sunbeam",
						Items: listitems,
					}, nil
				})
			}

			subcommand := args[1]
			for name, c := range command.SubCommands {
				if subcommand != name {
					continue
				}

				return runCommand(types.Command{
					Name:  filepath.Join(filepath.Dir(scriptPath), c.Entrypoint),
					Args:  args[2:],
					Input: input,
				})
			}

			return fmt.Errorf("subcommand not found: %s", subcommand)
		},
	}

	return cmd
}
