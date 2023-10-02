package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/cobra"
)

func NewCmdEdit() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit an extension",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			extensionPath, ok := extensions[args[0]]
			if !ok {
				return cmd.Help()
			}

			if mime, err := mimetype.DetectFile(extensionPath); err != nil || mime.String() == "application/octet-stream" {
				return fmt.Errorf("extension %s is not a text file", args[0])
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

func NewCmdList() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List installed extensions",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			for alias, extension := range extensions {
				fmt.Printf("%s\t%s\n", alias, extension)
			}

			return nil
		},
	}
}

func FindExtensions() (map[string]string, error) {
	var dirs []string
	if env, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		dirs = append(dirs, filepath.Join(env, "sunbeam", "extensions"))
	} else {
		dirs = append(dirs, filepath.Join(os.Getenv("HOME"), ".local", "share", "sunbeam", "extensions"))
	}

	dirs = append(dirs, filepath.SplitList(os.Getenv("PATH"))...)
	extensions := make(map[string]string)
	for _, dir := range dirs {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}

		dir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			if !strings.HasPrefix(entry.Name(), "sunbeam-") {
				continue
			}

			alias := strings.TrimPrefix(entry.Name(), "sunbeam-")
			if _, ok := extensions[alias]; ok {
				continue
			}

			extensions[alias] = filepath.Join(dir, entry.Name())
		}
	}

	return extensions, nil
}
