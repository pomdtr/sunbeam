package cmd

import (
	"fmt"

	"github.com/pomdtr/sunbeam/app"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(app.Version)
	},
}
