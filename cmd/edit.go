package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func NewCmdEdit() *cobra.Command {
	flags := struct {
		extension string
	}{}
	cmd := &cobra.Command{
		Use:     "edit [file]",
		Short:   "Open a file in your editor",
		GroupID: CommandGroupDev,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && flags.extension != "" {
				return fmt.Errorf("cannot specify both file and extension")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				tty, err := os.Open("/dev/tty")
				if err != nil {
					return err
				}
				editor := findEditor()
				editCmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", editor, args[0]))
				editCmd.Stdin = tty
				editCmd.Stdout = os.Stderr
				return editCmd.Run()
			}

			tempfile, err := os.CreateTemp("", "sunbeam-edit-")
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
			editor := findEditor()
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

	cmd.Flags().StringVarP(&flags.extension, "extension", "e", "", "File extension to use for temporary file")
	return cmd

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
