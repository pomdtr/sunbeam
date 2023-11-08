package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	Version = "dev"
	Date    = "unknown"
)

const (
	CommandGroupCore      = "core"
	CommandGroupDev       = "dev"
	CommandGroupExtension = "extension"
)

func IsSunbeamRunning() bool {
	return len(os.Getenv("SUNBEAM")) > 0
}

type NonInteractiveOutput struct {
	Extensions []extensions.Extension `json:"extensions"`
	Items      []types.ListItem       `json:"items"`
}

func NewRootCmd() (*cobra.Command, error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		Version:      fmt.Sprintf("%s (%s)", Version, Date),
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupCore,
		Title: "Core Commands:",
	}, &cobra.Group{
		ID:    CommandGroupDev,
		Title: "Development Commands:",
	})
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewCmdFetch())
	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(NewCmdExtension())
	rootCmd.AddCommand(NewCmdEdit())
	rootCmd.AddCommand(NewCmdCopy())
	rootCmd.AddCommand(NewCmdPaste())
	rootCmd.AddCommand(NewCmdOpen())

	docCmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation for sunbeam",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := buildDoc(rootCmd)
			if err != nil {
				return err
			}

			fmt.Printf("# CLI\n\n%s\n", doc)
			return nil
		},
	}
	rootCmd.AddCommand(docCmd)

	manCmd := &cobra.Command{
		Use:    "generate-man-pages [path]",
		Short:  "Generate Man Pages for sunbeam",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			header := &doc.GenManHeader{
				Title:   "MINE",
				Section: "3",
			}
			err := doc.GenManTree(rootCmd, header, args[0])
			if err != nil {
				return err
			}

			return nil
		},
	}
	rootCmd.AddCommand(manCmd)

	if IsSunbeamRunning() {
		return rootCmd, nil
	}

	extensionMap, err := FindExtensions()
	if err != nil {
		return nil, err
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			extensionMap, err := FindExtensions()
			if err != nil {
				return err
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")

			exts := make([]extensions.Extension, 0)
			for _, ext := range extensionMap {
				exts = append(exts, ext)
			}

			return encoder.Encode(NonInteractiveOutput{
				Extensions: exts,
				Items:      RootItems(cfg, extensionMap),
			})
		}
		rootList := tui.NewRootList("Sunbeam", func() (extensions.ExtensionMap, []types.ListItem, map[string]map[string]any, error) {
			extensionMap, err := FindExtensions()
			if err != nil {
				return nil, nil, nil, err
			}

			cfg, err := config.Load()
			if err != nil {
				return nil, nil, nil, err
			}

			return extensionMap, RootItems(cfg, extensionMap), cfg.Preferences, nil
		})
		return tui.Draw(rootList)
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupExtension,
		Title: "Extension Commands:",
	})

	for alias, extension := range extensionMap {
		command, err := NewCmdCustom(alias, extension)
		if err != nil {
			return nil, err
		}

		rootCmd.AddCommand(command)
	}

	return rootCmd, nil
}

func buildDoc(command *cobra.Command) (string, error) {
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
		if child.GroupID == CommandGroupExtension {
			continue
		}

		if child.Hidden {
			continue
		}

		childPage, err := buildDoc(child)
		if err != nil {
			return "", err
		}
		out.WriteString(childPage)
	}

	return out.String(), nil
}

func RootItems(cfg config.Config, extensionMap map[string]extensions.Extension) []types.ListItem {
	var items []types.ListItem
	for _, rootItem := range cfg.Oneliners {
		item, err := cfg.RootItem(rootItem, extensionMap)
		item.Id = fmt.Sprintf("root - %s", item.Title)
		if err != nil {
			continue
		}

		items = append(items, item)
	}

	for alias, extension := range extensionMap {
		items = append(items, ExtensionRootItems(alias, extension)...)
	}

	return items
}

func ExtensionRootItems(alias string, extension extensions.Extension) []types.ListItem {
	var items []types.ListItem
	for _, rootItem := range extension.RootItems() {
		listItem := types.ListItem{
			Id:          fmt.Sprintf("%s - %s", alias, rootItem.Title),
			Title:       rootItem.Title,
			Accessories: []string{extension.Title},
			Actions: []types.Action{
				{
					Title:     "Run",
					Type:      types.ActionTypeRun,
					Extension: alias,
					Command:   rootItem.Command,
					Params:    rootItem.Params,
				},
			},
		}

		if extension.Type == extensions.ExtensionTypeLocal {
			listItem.Actions = append(listItem.Actions, types.Action{
				Title:  "Edit",
				Key:    "e",
				Type:   types.ActionTypeEdit,
				Target: extension.Entrypoint,
			})
		}

		listItem.Actions = append(listItem.Actions, types.Action{
			Title:  "Copy Origin",
			Key:    "c",
			Type:   types.ActionTypeCopy,
			Target: extension.Origin,
		})

		items = append(items, listItem)
	}

	return items
}
