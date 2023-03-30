package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pomdtr/sunbeam/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type Page struct {
	Command      *cobra.Command
	HeaderOffset int
}

func buildDoc(command *cobra.Command, headerOffset int) (string, error) {
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
		if strings.HasPrefix(line, "#") {
			if headerOffset < 0 {
				line = strings.TrimPrefix(line, strings.Repeat("#", -headerOffset))
			} else {
				line = strings.Repeat("#", headerOffset) + line
			}
		}
		out.WriteString(line + "\n")
	}

	for _, child := range command.Commands() {
		childPage, err := buildDoc(child, headerOffset+1)
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

	doc, err := buildDoc(cmd, -1)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating docs:", err)
		os.Exit(1)
	}

	fmt.Println(doc)
}
