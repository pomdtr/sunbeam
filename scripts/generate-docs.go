package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pomdtr/sunbeam/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func buildDoc(command *cobra.Command) (string, error) {
	if command.GroupID == "extension" {
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

func main() {
	cmd, err := cmd.NewRootCmd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	doc, err := buildDoc(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating docs:", err)
		os.Exit(1)
	}

	fmt.Println(doc)
}
