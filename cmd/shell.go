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
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
			if err != nil {
				return err
			}

			if flags.Command != "" {
				return run(r, strings.NewReader(flags.Command), "command", args)
			}

			if len(args) == 0 {
				parser := syntax.NewParser()
				fmt.Printf("$ ")
				var runErr error
				fn := func(stmts []*syntax.Stmt) bool {
					if parser.Incomplete() {
						fmt.Printf("> ")
						return true
					}
					ctx := context.Background()
					for _, stmt := range stmts {
						runErr = r.Run(ctx, stmt)
						if r.Exited() {
							return false
						}
					}
					fmt.Printf("$ ")
					return true
				}
				if err := parser.Interactive(os.Stdin, fn); err != nil {
					return err
				}
				return runErr
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
