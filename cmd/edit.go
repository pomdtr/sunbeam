package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewCmdEdit(configPath string) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit an extension",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveDefault
			}

			extensions, err := FindExtensions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			completions := make([]string, 0)
			for alias, extension := range extensions {
				completions = append(completions, fmt.Sprintf("%s\t%s", alias, extension))
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return editFile(configPath)
			}
			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			extensionPath, ok := extensions[args[0]]
			if !ok {
				return cmd.Help()
			}

			return editFile(extensionPath)
		},
	}
}

func editFile(p string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	command := exec.Command("sh", "-c", fmt.Sprintf("%s %s", editor, p))
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
