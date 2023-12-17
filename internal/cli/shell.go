package cli

import (
	"context"
	"os"

	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func NewCmdShell() *cobra.Command {
	var flags struct {
		Command string
	}

	cmd := &cobra.Command{
		Use:     "shell",
		Short:   "Open a shell",
		GroupID: CommandGroupCore,
		Hidden:  true,
		Args:    cobra.ArbitraryArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && flags.Command != "" {
				return cmd.Usage()
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.Command != "" {
				return utils.RunCommand(flags.Command, "")
			}

			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			file, err := syntax.NewParser().Parse(f, "")
			if err != nil {
				return err
			}

			sh, err := interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr), interp.Params(args[1:]...))
			if err != nil {
				return err
			}

			return sh.Run(context.Background(), file)
		},
	}

	cmd.Flags().StringVarP(&flags.Command, "command", "c", "", "Command to run")

	return cmd
}
