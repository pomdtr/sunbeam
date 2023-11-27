package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/types"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	Version = "dev"
)

const (
	CommandGroupCore      = "core"
	CommandGroupExtension = "extension"
)

func IsSunbeamRunning() bool {
	return len(os.Getenv("SUNBEAM")) > 0
}

//go:embed embed/sunbeam.json
var configBytes []byte

func NewRootCmd() (*cobra.Command, error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupCore,
		Title: "Core Commands:",
	})
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewValidateCmd())
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

			fmt.Print(heredoc.Docf(`---
			outline: 2
			---

			# Cli

			%s
			`, doc))
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

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of sunbeam",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(Version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	if IsSunbeamRunning() {
		return rootCmd, nil
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupExtension,
		Title: "Extension Commands:",
	})

	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		if _, ok := os.LookupEnv("SUNBEAM_CONFIG"); ok {
			return nil, fmt.Errorf("config file not found: %s", config.Path)
		}

		if err := os.MkdirAll(filepath.Dir(config.Path), 0755); err != nil {
			return nil, err
		}

		if err := os.WriteFile(config.Path, []byte(configBytes), 0644); err != nil {
			return nil, err
		}
	}

	cfg, err := config.Load(config.Path)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(NewCmdExtension(cfg))

	extensionMap := make(map[string]extensions.Extension)
	for alias, extensionConfig := range cfg.Extensions {
		extension, err := extensions.LoadExtension(extensionConfig.Origin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading extension %s: %s\n", alias, err)
			continue
		}
		extensionMap[alias] = extension

		command, err := NewCmdCustom(alias, extension, extensionConfig)
		if err != nil {
			return nil, err
		}
		rootCmd.AddCommand(command)
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		history, err := history.Load(history.Path)
		if err != nil {
			return err
		}

		rootList := tui.NewRootList("Sunbeam", history, func() (config.Config, []types.ListItem, error) {
			cfg, err := config.Load(config.Path)
			if err != nil {
				return config.Config{}, nil, err
			}

			var items []types.ListItem
			items = append(items, onelinerListItems(cfg.Oneliners)...)

			extensionMap := make(map[string]extensions.Extension)
			for alias, extensionConfig := range cfg.Extensions {
				extension, err := extensions.LoadExtension(extensionConfig.Origin)
				if err != nil {
					continue
				}
				extensionMap[alias] = extension
				items = append(items, extensionListItems(alias, extension, extensionConfig)...)
			}

			return cfg, items, nil
		})
		return tui.Draw(rootList)

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

func onelinerListItems(oneliners []config.Oneliner) []types.ListItem {
	var items []types.ListItem
	for _, oneliner := range oneliners {
		item := types.ListItem{
			Id:          fmt.Sprintf("oneliner - %s", oneliner.Command),
			Title:       oneliner.Title,
			Accessories: []string{"Oneliner"},
			Actions: []types.Action{
				{
					Title:   "Run",
					Type:    types.ActionTypeExec,
					Command: oneliner.Command,
					Dir:     oneliner.Dir,
					Exit:    oneliner.Exit,
				},
				{
					Title: "Copy Command",
					Key:   "c",
					Type:  types.ActionTypeCopy,
					Text:  oneliner.Command,
				},
			},
		}

		items = append(items, item)
	}

	return items
}

func extensionListItems(alias string, extension extensions.Extension, extensionConfig config.ExtensionConfig) []types.ListItem {
	var items []types.ListItem

	var rootItems []types.RootItem
	if extensionConfig.Root != nil {
		for _, name := range extensionConfig.Root {
			command, ok := extension.Command(name)
			if !ok {
				continue
			}
			rootItems = append(rootItems, types.RootItem{
				Title:   command.Title,
				Command: command.Name,
			})
		}
	} else {
		rootItems = append(rootItems, extension.RootItems()...)
	}
	rootItems = append(rootItems, extensionConfig.Items...)

	for _, rootItem := range rootItems {
		item := types.ListItem{
			Id:          fmt.Sprintf("%s - %s", alias, rootItem.Title),
			Title:       rootItem.Title,
			Subtitle:    extension.Manifest.Title,
			Accessories: []string{"Command"},
			Actions: []types.Action{
				{
					Title:     "Run",
					Type:      types.ActionTypeRun,
					Extension: alias,
					Command:   rootItem.Command,
					Params:    rootItem.Params,
					Exit:      true,
				},
			},
		}

		if !extensions.IsRemote(extensionConfig.Origin) {
			item.Actions = append(item.Actions, types.Action{
				Title:  "Edit Extension",
				Key:    "e",
				Type:   types.ActionTypeEdit,
				Path:   extension.Entrypoint,
				Reload: true,
			})
		} else {
			item.Actions = append(item.Actions, types.Action{
				Title:   "View Source",
				Key:     "c",
				Type:    types.ActionTypeExec,
				Command: fmt.Sprintf("curl %s | %s", extensionConfig.Origin, utils.FindPager()),
			})
		}

		if len(extensionConfig.Preferences) > 0 {
			item.Actions = append(item.Actions, types.Action{
				Title:     "Configure Extension",
				Key:       "s",
				Type:      types.ActionTypeConfig,
				Extension: alias,
			})
		}

		items = append(items, item)
	}

	return items
}
