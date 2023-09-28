package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

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
	path := os.Getenv("PATH")

	extensions := make(map[string]string)
	for _, dir := range filepath.SplitList(path) {
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
