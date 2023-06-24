package main

import (
	"fmt"

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

	for _, item := range catalog {
		fmt.Println()
		fmt.Printf("## %s\n", item.Title)
		fmt.Println()
		fmt.Printf("> %s\n", item.Description)
		fmt.Println()
		fmt.Printf("```bash\nsunbeam extension install <alias> '%s'\n```\n", item.Origin)
	}
}
