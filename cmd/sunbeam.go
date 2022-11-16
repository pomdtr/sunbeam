package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/tui"
)

var globalOptions tui.SunbeamOptions

var rootOptions struct {
	With []string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sunbeam",
	Short: "Command Line Launcher",
	Run:   Sunbeam,
	Args:  cobra.ArbitraryArgs,
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&globalOptions.Width, "width", "W", 0, "width of the window")
	rootCmd.PersistentFlags().IntVarP(&globalOptions.Height, "height", "H", 0, "height of the window")
	rootCmd.Flags().StringArrayVarP(&rootOptions.With, "with", "w", nil, "script parameters in the form of name=value")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// TODO: handle arbitrary args (https://github.com/cli/cli/blob/trunk/cmd/gh/main.go)
func Sunbeam(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		manifests := make([]api.Manifest, 0)
		for _, manifest := range api.Sunbeam.Extensions {
			manifests = append(manifests, manifest)
		}

		rootList := tui.RootList(manifests...)
		err := tui.Draw(rootList, globalOptions)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	extensionName := args[0]

	manifest, ok := api.Sunbeam.Extensions[extensionName]
	if !ok {
		log.Fatalf("Extension %s not found", extensionName)
	}

	if len(args) < 2 {
		model := tui.RootList(manifest)
		err := tui.Draw(model, globalOptions)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	scriptName := args[1]

	script, ok := manifest.Scripts[scriptName]
	if !ok {
		log.Fatalf("Page not found: %s", scriptName)
	}

	itemMap := make(map[string]api.ScriptInput)
	for _, formItem := range script.Inputs {
		itemMap[formItem.Name] = formItem
	}

	scriptParams := make(map[string]any)
	for _, param := range rootOptions.With {
		tokens := strings.SplitN(param, "=", 2)
		if len(tokens) != 2 {
			log.Fatalf("Invalid parameter: %s", param)
		}

		name, value := tokens[0], tokens[1]
		formItem, ok := itemMap[name]
		if !ok {
			log.Fatalf("Params %s does not exists in script %s", name, tokens[1])
		}
		switch formItem.Type {
		case "textfield":
			scriptParams[name] = value
		case "checkbox":
			scriptParams[name] = value == "true"
		}
	}

	container := tui.NewRunContainer(manifest, script, scriptParams)
	err := tui.Draw(container, globalOptions)
	if err != nil {
		log.Fatalf("could not run script: %v", err)
	}
}
