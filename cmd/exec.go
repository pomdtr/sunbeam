package cmd

import (
	"github.com/spf13/cobra"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func NewCmdExec() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <script>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Exec a script",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			i := interp.New(interp.Options{
				Args:         args,
				Unrestricted: true,
			})
			i.Use(stdlib.Symbols)

			imports, err := cmd.Flags().GetStringSlice("include")
			if err != nil {
				return err
			}

			for _, pkg := range imports {
				if _, err := i.EvalPath(pkg); err != nil {
					return err
				}
			}

			if _, err := i.EvalPath(args[0]); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringSliceP("include", "i", []string{}, "Include another script")

	return cmd
}
