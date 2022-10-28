package cmd

import (
	"fmt"

	"github.com/pomdtr/sunbeam/api"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available sunbeam commands",
	Args:  cobra.NoArgs,
	Run:   sunbeamList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func sunbeamList(cmd *cobra.Command, args []string) {
	for _, manifest := range api.Sunbeam.Extensions {
		for commandName := range manifest.Commands {
			fmt.Printf("%s.%s\n", manifest.Name, commandName)
		}
	}
}
