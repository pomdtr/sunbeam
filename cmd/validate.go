package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/spf13/cobra"
)

func NewValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a Sunbeam schema",
	}

	cmd.AddCommand(NewCmdValidatePage())
	cmd.AddCommand(NewCmdValidateManifest())
	cmd.AddCommand(NewCmdValidateCommand())

	return cmd
}

func NewCmdValidatePage() *cobra.Command {
	return &cobra.Command{
		Use:   "page",
		Short: "Validate a page",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no input provided")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("unable to read stdin: %s", err)
			}

			if err := schemas.ValidatePage(input); err != nil {
				return err
			}

			fmt.Println("✅ Input is valid!")
			return nil
		},
	}

}

func NewCmdValidateManifest() *cobra.Command {
	return &cobra.Command{
		Use:   "manifest",
		Short: "Validate a manifest",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no input provided")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("unable to read stdin: %s", err)
			}

			if err := schemas.ValidateManifest(input); err != nil {
				return err
			}

			fmt.Println("✅ Input is valid!")
			return nil
		},
	}
}

func NewCmdValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "command",
		Short: "Validate a command",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no input provided")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("unable to read stdin: %s", err)
			}

			if err := schemas.ValidateCommand(input); err != nil {
				return err
			}

			fmt.Println("✅ Input is valid!")
			return nil
		},
	}
}
