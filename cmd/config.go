package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/shlex"
	"github.com/spf13/cobra"
	"github.com/tailscale/hujson"
)

func NewCmdConfig(configPath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage sunbeam config",
	}

	cmd.AddCommand(NewCmdConfigDump(configPath))
	cmd.AddCommand(NewCmdConfigEdit(configPath))

	return cmd
}

func NewCmdConfigDump(configPath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump config as JSON",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			raw, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("error reading config: %w", err)
			}

			standardized, err := hujson.Standardize(raw)
			if err != nil {
				return fmt.Errorf("error standardizing config: %w", err)
			}

			var config any
			if err := json.Unmarshal(standardized, &config); err != nil {
				return fmt.Errorf("error unmarshaling config: %w", err)
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(config); err != nil {
				return fmt.Errorf("error encoding config: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func NewCmdConfigEdit(configPath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit sunbeam config",
		Long:  "Edit sunbeam config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			editor, _ := cmd.Flags().GetString("editor")

			args, err := shlex.Split(editor)
			if err != nil {
				return fmt.Errorf("error splitting editor command: %w", err)
			}
			args = append(args, configPath)

			editorCmd := exec.Command(args[0], args[1:]...)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr

			if err := editorCmd.Run(); err != nil {
				return fmt.Errorf("error running editor: %w", err)
			}

			return nil
		},
	}

	defaultEditor := os.Getenv("EDITOR")
	if defaultEditor == "" {
		defaultEditor = "vim"
	}

	cmd.Flags().String("editor", defaultEditor, "editor to use")

	return cmd
}
