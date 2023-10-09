package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

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

func NewRootCmd() (*cobra.Command, error) {
	extensions, err := FindExtensions()
	if err != nil {
		return nil, err
	}

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		Version:      fmt.Sprintf("%s (%s)", Version, Date),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				return err
			}
			historyPath := filepath.Join(cacheDir, "sunbeam", "history.json")
			history, err := LoadHistory(historyPath)
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}

				history = History{
					entries: make(map[string]int64),
					path:    historyPath,
				}
			}

			generator := func() (map[string]tui.Extension, []types.ListItem, error) {
				config, err := LoadConfig()
				if err != nil && !os.IsNotExist(err) {
					return nil, nil, err
				}

				items := make([]types.ListItem, 0)
				for alias, extension := range extensions {
					for _, command := range extension.Commands {
						if !IsRootCommand(command) {
							continue
						}

						if command.Hidden {
							continue
						}

						items = append(items, types.ListItem{
							Id:       fmt.Sprintf("extensions/%s/%s", alias, command.Name),
							Title:    command.Title,
							Subtitle: extension.Title,
							Actions: []types.Action{
								{
									Title: "Run",
									OnAction: types.Command{
										Type:      types.CommandTypeRun,
										Extension: alias,
										Command:   command.Name,
									},
								},
							},
						})
					}
				}

				for title, command := range config.Root {
					items = append(items, types.ListItem{
						Id:       fmt.Sprintf("commands/%s", title),
						Title:    title,
						Subtitle: "Command",
						Actions: []types.Action{
							{
								Title:    "Run Action",
								OnAction: command,
							},
						},
					})
				}

				sort.Slice(items, func(i, j int) bool {
					timestampA, ok := history.entries[items[i].Id]
					if !ok {
						return false
					}

					timestampB, ok := history.entries[items[j].Id]
					if !ok {
						return true
					}

					return timestampA > timestampB
				})

				return extensions, items, nil
			}

			rootList := tui.NewRootList("Sunbeam", generator)
			rootList.OnSelect = func(id string) {
				history.entries[id] = time.Now().Unix()
				_ = history.Save()
			}

			return tui.Draw(rootList, MaxHeight)
		},
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupCore,
		Title: "Core Commands:",
	})
	if len(extensions) > 0 {
		rootCmd.AddGroup(&cobra.Group{
			ID:    CommandGroupExtension,
			Title: "Extension Commands:",
		})
	}

	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(NewCmdFetch())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewCmdExtension())
	rootCmd.AddCommand(NewCmdServe())

	docCmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation for sunbeam",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := buildDoc(rootCmd)
			if err != nil {
				return err
			}

			fmt.Println(doc)
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

	for alias := range extensions {
		command, err := NewCmdCustom(extensions, alias)
		if err != nil {
			return nil, err
		}

		rootCmd.AddCommand(command)
	}

	return rootCmd, nil
}

func IsRootCommand(command types.CommandSpec) bool {
	if command.Hidden {
		return false
	}

	for _, param := range command.Params {
		if !param.Optional {
			return false
		}
	}

	return true
}
