package main

import (
	"fmt"
	"strings"

	"github.com/pomdtr/sunbeam/catalog"
)

func main() {
	catalog, err := catalog.GetCatalog()
	if err != nil {
		panic(err)
	}

	fmt.Println("# Sunbeam Extensions")
	fmt.Println()
	fmt.Println("Use `sunbeam extension browse` to list extensions from the command line.")
	fmt.Println()

	for _, item := range catalog {
		normalizedTitle := strings.ToLower(item.Title)
		normalizedTitle = strings.ReplaceAll(normalizedTitle, " ", "-")
		fmt.Printf("- [%s](#%s)\n", item.Title, normalizedTitle)
	}

	for _, item := range catalog {
		fmt.Printf("### %s - %s\n", item.Title, item.Author)
		fmt.Println()
		fmt.Printf("> %s\n", item.Description)

		fmt.Printf("```bash\nsunbeam extension install <alias> '%s'\n```\n", item.Origin)
	}
}
