package cmd

import (
	"fmt"

	"github.com/pomdtr/sunbeam/app"
	"github.com/spf13/cobra"
)

func NewCmdVersion() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(app.Version)
		},
	}
}
