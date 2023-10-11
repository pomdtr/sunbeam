package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewCmdEdit() *cobra.Command {
	return &cobra.Command{
		Use:     "edit <file>",
		Short:   "Open a file in your editor",
		GroupID: CommandGroupDev,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open("/dev/tty")
			if err != nil {
				return err
			}

			editor := findEditor()
			editCmd := exec.Command("sh", "-c", editor+" "+args[0])
			editCmd.Stdin = f
			editCmd.Stdout = os.Stderr

			return editCmd.Run()
		},
	}
}

func findEditor() string {
	if editor, ok := os.LookupEnv("VISUAL"); ok {
		return editor
	}

	if editor, ok := os.LookupEnv("EDITOR"); ok {
		return editor
	}

	return "vim"
}
