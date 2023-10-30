package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
)

func NewCmdConfig() *cobra.Command {
	return &cobra.Command{
		Use:     "config",
		Short:   "Edit the Sunbeam config",
		GroupID: CommandGroupCore,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			editor := utils.FindEditor()

			editCmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", editor, config.Path))
			editCmd.Stdin = os.Stdin
			editCmd.Stdout = os.Stdout
			editCmd.Stderr = os.Stderr

			return editCmd.Run()
		},
	}
}
