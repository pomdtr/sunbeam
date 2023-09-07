package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/schemas"
	"github.com/spf13/cobra"
)

func NewValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate",
		Short:   "Validate a Sunbeam schema",
		GroupID: coreGroupID,
	}

	cmd.AddCommand(NewValidatePageCmd())
	cmd.AddCommand(NewValidateManifestCmd())

	return cmd
}

func NewValidatePageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "page",
		Short: "Validate a page",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input []byte

			if len(args) == 1 {
				c, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("unable to read file: %s", err)
				}
				input = c
			} else if !isatty.IsTerminal(os.Stdin.Fd()) {
				c, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("unable to read stdin: %s", err)
				}

				input = c
			} else {
				return fmt.Errorf("no input provided")
			}

			if err := schemas.PageSchema.Validate(input); err != nil {
				return err
			}

			fmt.Println("✅ Input is valid!")
			return nil
		},
	}

}

func NewValidateManifestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manifest",
		Short: "Validate a manifest",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input []byte

			if len(args) == 1 {
				c, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("unable to read file: %s", err)
				}
				input = c
			} else if !isatty.IsTerminal(os.Stdin.Fd()) {
				c, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("unable to read stdin: %s", err)
				}

				input = c
			} else {
				return fmt.Errorf("no input provided")
			}

			if err := schemas.Validate(schemas.ManifestSchema, input); err != nil {
				return err
			}

			fmt.Println("✅ Input is valid!")
			return nil
		},
	}

}
