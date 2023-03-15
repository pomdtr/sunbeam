package main

import (
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/pomdtr/sunbeam/cmd"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed docs/schema.json
var schema string

var version = "dev"

func main() {
	var pageShema *jsonschema.Schema

	compiler := jsonschema.NewCompiler()

	if err := compiler.AddResource("https://pomdtr.github.io/sunbeam/schemas/page.json", strings.NewReader(schema)); err != nil {
		fmt.Fprintln(os.Stderr, "failed to add schema to compiler: ", err)
		os.Exit(1)
	}
	pageShema, err := compiler.Compile("https://pomdtr.github.io/sunbeam/schemas/page.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to compile schema: ", err)
	}

	if err := cmd.Execute(version, pageShema); err != nil {
		os.Exit(1)
	}
}
