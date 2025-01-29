package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
)

func NewCmdEdit() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit [file]",
		Short:   "Open a file in your editor",
		GroupID: CommandGroupCore,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				editCmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", utils.FindEditor(), args[0]))
				editCmd.Stdin = os.Stdin
				editCmd.Stdout = os.Stdout
				editCmd.Stderr = os.Stderr
				return editCmd.Run()
			}

			tempfile, err := os.CreateTemp("", "sunbeam-*")
			if err != nil {
				return err
			}
			defer os.Remove(tempfile.Name())

			if !isatty.IsTerminal(os.Stdin.Fd()) {
				f, err := os.OpenFile(tempfile.Name(), os.O_RDWR, 0644)
				if err != nil {
					return err
				}

				if _, err := io.Copy(f, os.Stdin); err != nil {
					return err
				}
				if err := f.Close(); err != nil {
					return err
				}
			}

			tty, err := os.Open("/dev/tty")
			if err != nil {
				return err
			}
			editor := utils.FindEditor()
			editCmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", editor, tempfile.Name()))
			editCmd.Stdin = tty
			editCmd.Stdout = os.Stderr
			if err := editCmd.Run(); err != nil {
				return err
			}

			f, err := os.Open(tempfile.Name())
			if err != nil {
				return err
			}

			if _, err := io.Copy(os.Stdout, f); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd

}
