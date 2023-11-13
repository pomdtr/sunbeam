package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func NewCmdShell() *cobra.Command {
	flags := struct {
		Command string
	}{}

	run := func(r *interp.Runner, reader io.Reader, name string, args []string) error {
		prog, err := syntax.NewParser().Parse(reader, name)
		if err != nil {
			return err
		}
		r.Reset()
		r.Params = args
		ctx := context.Background()
		return r.Run(ctx, prog)
	}

	cmd := &cobra.Command{
		Use:     "shell [script]",
		Short:   `Execute a command or script in a shell`,
		GroupID: CommandGroupCore,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.Command == "" && len(args) == 0 {
				return fmt.Errorf("must specify either a command or a file")
			}

			return nil
		},
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
			if err != nil {
				return err
			}

			if flags.Command != "" {
				return run(r, strings.NewReader(flags.Command), "command", args)
			}

			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()
			return run(r, f, args[0], args[1:])
		},
	}

	cmd.Flags().StringVarP(&flags.Command, "command", "c", "", "command to run")

	return cmd
}
