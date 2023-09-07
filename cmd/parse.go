package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func NewParseCmd() *cobra.Command {
	flags := struct {
		shell string
	}{}

	cmd := &cobra.Command{
		Use:     "parse",
		Short:   "Parse command input from stdin and print environment variables for the command",
		GroupID: coreGroupID,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			shell := path.Base(flags.shell)

			if shell != "bash" && shell != "zsh" && shell != "fish" {
				return fmt.Errorf("invalid shell %s", flags.shell)
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return nil
			}

			decoder := json.NewDecoder(os.Stdin)
			var input CommandInput
			if err := decoder.Decode(&input); err != nil {
				return fmt.Errorf("failed to decode input: %w", err)
			}

			shell := filepath.Base(flags.shell)

			printEnv(shell, "COMMAND", input.Command)
			if input.Query != "" {
				printEnv(shell, "QUERY", input.Query)
			}
			for k, v := range input.Params {
				printEnv(shell, fmt.Sprintf("ARG_%s", strings.ToUpper(k)), v)
			}

			return nil
		},
	}

	cmd.Flags().StringP("shell", "s", os.Getenv("SHELL"), "Shell to generate output for")

	return cmd
}

func printEnv(shell string, key string, value any) {
	switch shell {
	case "bash", "zsh":
		fmt.Printf("%s=%v\n", key, value)
	case "fish":
		fmt.Printf("set %s %v\n", key, value)
	}
}
