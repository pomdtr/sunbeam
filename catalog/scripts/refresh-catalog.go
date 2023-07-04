package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/pomdtr/sunbeam/catalog"
)

func main() {
	jsonOutput := flag.Bool("json", false, "output as json")
	flag.Parse()

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	input = bytes.TrimSpace(input)

	rows := bytes.Split(input, []byte("\n"))
	items := make([]*catalog.CatalogItem, len(rows))
	for i, row := range rows {
		target := string(row)
		log.Println("fetching", target)
		resp, err := http.Get(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}

		metadata, err := catalog.ExtractCommandMetadata(body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}

		items[i] = &catalog.CatalogItem{
			Metadata: metadata,
			Origin:   string(row),
		}
	}

	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(items); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	fmt.Println("# Command Catalog")
	fmt.Println()
	fmt.Println("Use `sunbeam comman browse` to list commands from your terminal.")

	for _, item := range items {
		fmt.Println()
		fmt.Printf("## %s\n", item.Metadata.Title)
		fmt.Println()
		fmt.Printf("> %s\n", item.Metadata.Description)
		fmt.Println()
		fmt.Printf("```bash\nsunbeam command add <name> '%s'\n```\n", item.Origin)
	}
}
