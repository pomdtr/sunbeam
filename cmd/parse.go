package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/pkg"
	"github.com/spf13/cobra"
)

func NewCmdParse() *cobra.Command {
	shells := []string{"bash", "zsh", "fish"}
	cmd := &cobra.Command{
		Use:       "parse",
		Short:     "Parse command input from stdin and print environment variables for the command",
		Args:      cobra.ExactArgs(1),
		ValidArgs: shells,
		GroupID:   coreGroupID,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			for _, shell := range shells {
				if shell == args[0] {
					return nil
				}
			}
			return fmt.Errorf("invalid shell: %s", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return nil
			}

			decoder := json.NewDecoder(os.Stdin)
			var input pkg.CommandInput
			if err := decoder.Decode(&input); err != nil {
				return fmt.Errorf("failed to decode input: %w", err)
			}

			if input.Query != "" {
				printEnv(shell, "QUERY", input.Query)
			}
			for k, v := range input.Params {
				printEnv(shell, strings.ReplaceAll(strings.ToUpper(k), "-", "_"), v)
			}

			return nil
		},
	}

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
