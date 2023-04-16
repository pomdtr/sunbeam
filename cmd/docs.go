package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func buildDoc(command *cobra.Command) (string, error) {
	if command.Hidden {
		return "", nil
	}
	if command.Name() == "help" {
		return "", nil
	}

	var page strings.Builder
	err := doc.GenMarkdown(command, &page)
	if err != nil {
		return "", err
	}

	out := strings.Builder{}
	for _, line := range strings.Split(page.String(), "\n") {
		if strings.Contains(line, "SEE ALSO") {
			break
		}

		out.WriteString(line + "\n")
	}

	for _, child := range command.Commands() {
		childPage, err := buildDoc(child)
		if err != nil {
			return "", err
		}
		out.WriteString(childPage)
	}

	return out.String(), nil
}

func NewGeneratorDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "generate-docs",
		Short:  "Generate documentation for sunbeam",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := buildDoc(cmd.Root())
			if err != nil {
				return err
			}

			fmt.Println(doc)
			return nil
		},
	}

	return cmd
}
