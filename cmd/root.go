package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
		Use:                "sunbeam",
		Short:              "Command Line Launcher",
		SilenceUsage:       true,
		DisableFlagParsing: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveDefault
			}

			entrypoint, err := filepath.Abs(args[0])
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			extension, err := extensions.ExtractManifest(entrypoint)
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for _, command := range extension.Commands {
				completions = append(completions, fmt.Sprintf("%s\t%s", command.Name, command.Title))
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		Args: cobra.ArbitraryArgs,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupCore,
		Title: "Core Commands:",
	})
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewCmdFetch())
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

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of sunbeam",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Sunbeam %s (%s)\n", Version, Date)
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

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	rootCmd.AddCommand(NewCmdUpgrade(cfg))
	extensionMap := make(map[string]extensions.Extension)
	for alias, extensionConfig := range cfg.Extensions {
		extension, err := extensions.LoadExtension(extensionConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading extension %s: %s\n", alias, err)
			continue
		}
		extensionMap[alias] = extension

		command, err := NewCmdCustom(alias, extension)
		if err != nil {
			return nil, err
		}
		rootCmd.AddCommand(command)
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
			return cmd.Help()
		}

		if len(args) == 0 {
			if len(LoadRootItems(cfg.Oneliners, extensionMap)) == 0 {
				return cmd.Usage()
			}

			rootList := tui.NewRootList("Sunbeam", func() (extensions.ExtensionMap, []types.ListItem, error) {
				cfg, err := config.Load()
				if err != nil {
					return nil, nil, err
				}

				extensionMap := make(map[string]extensions.Extension)
				for alias, extensionConfig := range cfg.Extensions {
					extension, err := extensions.LoadExtension(extensionConfig)
					if err != nil {
						continue
					}
					extensionMap[alias] = extension
				}

				return extensionMap, LoadRootItems(cfg.Oneliners, extensionMap), nil
			})
			return tui.Draw(rootList)
		}

		var entrypoint string
		if args[0] == "-" {
			tempfile, err := os.CreateTemp("", "entrypoint-*%s")
			if err != nil {
				return err
			}
			defer os.Remove(tempfile.Name())

			if _, err := io.Copy(tempfile, os.Stdin); err != nil {
				return err
			}

			if err := tempfile.Close(); err != nil {
				return err
			}

			entrypoint = tempfile.Name()
		} else if extensions.IsRemote(args[0]) {
			tempfile, err := os.CreateTemp("", "entrypoint-*%s")
			if err != nil {
				return err
			}
			defer os.Remove(tempfile.Name())

			if err := extensions.DownloadEntrypoint(args[0], tempfile.Name()); err != nil {
				return err
			}

			entrypoint = tempfile.Name()
		} else {
			e, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			if _, err := os.Stat(e); err != nil {
				return fmt.Errorf("error loading extension: %w", err)
			}

			entrypoint = e
		}

		if err := os.Chmod(entrypoint, 0755); err != nil {
			return err
		}

		manifest, err := extensions.ExtractManifest(entrypoint)
		if err != nil {
			return fmt.Errorf("error loading extension: %w", err)
		}

		extension := extensions.Extension{
			Manifest:   manifest,
			Entrypoint: entrypoint,
			Config: extensions.Config{
				Origin: entrypoint,
			},
		}

		rootCmd, err := NewCmdCustom(filepath.Base(entrypoint), extension)
		if err != nil {
			return fmt.Errorf("error loading extension: %w", err)
		}

		rootCmd.Use = "extension"
		rootCmd.SetArgs(args[1:])
		return rootCmd.Execute()
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

func LoadRootItems(oneliners map[string]config.Oneliner, extensionMap map[string]extensions.Extension) []types.ListItem {
	var items []types.ListItem
	for title, oneliner := range oneliners {
		item := types.ListItem{
			Id:          fmt.Sprintf("root - %s", title),
			Title:       title,
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
					Exit:  true,
				},
			},
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
	for _, rootItem := range extension.Root() {
		listItem := types.ListItem{
			Id:          fmt.Sprintf("%s - %s", alias, rootItem.Title),
			Title:       rootItem.Title,
			Accessories: []string{extension.Manifest.Title},
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

		listItem.Actions = append(listItem.Actions, types.Action{
			Title:  "Copy Origin",
			Key:    "c",
			Type:   types.ActionTypeCopy,
			Target: extension.Config.Origin,
			Exit:   true,
		})

		items = append(items, listItem)
	}

	return items
}
