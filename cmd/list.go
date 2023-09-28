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
		Use:   "list",
		Short: "List installed extensions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions, err := ListExtensions()
			if err != nil {
				return err
			}

			for _, extension := range extensions {
				name := strings.TrimPrefix(filepath.Base(extension), "sunbeam-")
				fmt.Printf("%s\t%s\n", name, extension)
			}

			return nil
		},
	}
}

func ListExtensions() ([]string, error) {
	path := os.Getenv("PATH")
	var extensions []string
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

			extensions = append(extensions, filepath.Join(dir, entry.Name()))
		}
	}

	return extensions, nil
}
