package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	Version = "dev"
	Date    = "unknown"
)

var (
	MaxHeigth = LookupIntEnv("SUNBEAM_HEIGHT", 0)
)

type ExtensionCache map[string]types.Manifest

type Config struct {
	RootItems map[string]string `json:"root"`
}

func LoadConfig(fp string) (Config, error) {
	f, err := os.Open(fp)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func dataHome() string {
	if env, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	return filepath.Join(os.Getenv("HOME"), ".local", "share", "sunbeam")
}

func configPath() string {
	if env, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		return filepath.Join(env, "sunbeam", "config.json")
	}

	return filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "config.json")
}

type History struct {
	entries map[string]int64
	path    string
}

func LoadHistory(fp string) (History, error) {
	f, err := os.Open(fp)
	if err != nil {
		return History{}, err
	}

	var entries map[string]int64
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return History{}, err
	}

	return History{
		entries: entries,
		path:    fp,
	}, nil
}

func (h History) Save() error {
	f, err := os.OpenFile(h.path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(h.path), 0755); err != nil {
			return err
		}

		f, err = os.Create(h.path)
		if err != nil {
			return err
		}
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(h.entries); err != nil {
		return err
	}

	return nil
}

func NewRootCmd() (*cobra.Command, error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:     "sunbeam",
		Short:   "Command Line Launcher",
		Version: fmt.Sprintf("%s (%s)", Version, Date),
		Args:    cobra.ArbitraryArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			extensions, err := FindExtensions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			if len(args) == 0 {
				var completions []string
				for alias := range extensions {
					completions = append(completions, fmt.Sprintf("%s\tExtension command", alias))
				}

				return completions, cobra.ShellCompDirectiveNoFileComp
			}

			if len(args) == 1 {
				extensionPath, ok := extensions[args[0]]
				if !ok {
					return nil, cobra.ShellCompDirectiveDefault
				}

				extension, err := tui.LoadExtension(extensionPath)
				if err != nil {
					return nil, cobra.ShellCompDirectiveDefault
				}

				completions := make([]string, 0)
				for _, command := range extension.Commands {
					if extension.Root == command.Name {
						continue
					}
					completions = append(completions, fmt.Sprintf("%s\t%s", command.Name, command.Title))
				}

				return completions, cobra.ShellCompDirectiveNoFileComp
			}

			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		SilenceUsage:       true,
		DisableFlagParsing: true,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if args[0] == "--help" || args[0] == "-h" {
					return cmd.Help()
				}

				extensions, err := FindExtensions()
				if err != nil {
					return err
				}

				commandPath, ok := extensions[args[0]]
				if !ok {
					return cmd.Help()
				}

				command, err := NewCmdCustom(commandPath)
				if err != nil {
					return err
				}

				command.SetArgs(args[1:])
				command.Use = args[0]
				return command.Execute()
			}

			configPath := configPath()
			config, err := LoadConfig(configPath)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			if len(config.RootItems) == 0 {
				return cmd.Usage()
			}

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

			items := make([]types.ListItem, 0)
			for title, command := range config.RootItems {
				ref, err := ExtractCommand(command)
				if err != nil {
					continue
				}

				items = append(items, types.ListItem{
					Title:       title,
					Id:          title,
					Accessories: []string{strings.TrimPrefix(command, "sunbeam ")},
					Actions: []types.Action{
						{
							Title: "Run Command",
							OnAction: types.Command{
								Type:    types.CommandTypeRun,
								Script:  ref.Script,
								Command: ref.Command,
								Params:  ref.Params,
							},
						},
						{
							Title: "Copy Command",
							Key:   "c",
							OnAction: types.Command{
								Type: types.CommandTypeCopy,
								Text: fmt.Sprintf("%s %s", os.Args[0], command),
								Exit: true,
							},
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

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(items); err != nil {
					return err
				}

				return nil
			}

			rootList := tui.NewRootList(tui.Extensions{}, items...)
			rootList.OnSelect = func(id string) {
				history.entries[id] = time.Now().Unix()
				_ = history.Save()
			}

			return tui.Draw(rootList, MaxHeigth)

		},
	}

	rootCmd.AddCommand(NewCmdRun())
	// rootCmd.AddCommand(NewCmdServe())
	rootCmd.AddCommand(NewCmdFetch())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewCmdExtension())

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

	return rootCmd, nil
}
