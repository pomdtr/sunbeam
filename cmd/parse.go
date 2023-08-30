package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func NewParseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "parse",
		RunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return nil
			}

			decoder := json.NewDecoder(os.Stdin)
			var input map[string]any
			if err := decoder.Decode(&input); err != nil {
				return fmt.Errorf("failed to decode input: %w", err)
			}

			shellPath, _ := cmd.Flags().GetString("shell")
			shell := filepath.Base(shellPath)

			for k, v := range input {
				switch shell {
				case "bash", "zsh":
					fmt.Printf("ARG_%s=%v\n", strings.ToUpper(k), v)
				case "fish":
					fmt.Printf("set ARG_%s %v\n", strings.ToUpper(k), v)
				default:
					return fmt.Errorf("unsupported shell: %s", shell)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("shell", "s", os.Getenv("SHELL"), "Shell to generate output for")

	return cmd
}
