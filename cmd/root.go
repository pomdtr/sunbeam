package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/tailscale/hujson"
)

var (
	Version = "dev"
	Date    = "unknown"
)

var (
	MaxHeigth = LookupIntEnv("SUNBEAM_HEIGHT", 0)
)

type Config struct {
	Commands map[string]types.Command `json:"commands"`
}

func LoadConfig() (Config, error) {
	var configPath string
	if env, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		if _, err := os.Stat(filepath.Join(env, "sunbeam", "config.json")); err == nil {
			configPath = filepath.Join(env, "sunbeam", "config.json")
		} else if _, err := os.Stat(filepath.Join(env, "sunbeam", "config.jsonc")); err == nil {
			configPath = filepath.Join(env, "sunbeam", "config.jsonc")
		} else {
			return Config{}, fmt.Errorf("config file not found")
		}
	} else {
		if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "config.jsonc")); err == nil {
			configPath = filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "config.jsonc")
		} else if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "config.json")); err == nil {
			configPath = filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "config.json")
		} else {
			return Config{}, fmt.Errorf("config file not found")
		}
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	jsonBytes, err := hujson.Standardize(configBytes)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

type ExtensionCache map[string]types.Manifest

func dataHome() string {
	if env, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	return filepath.Join(os.Getenv("HOME"), ".local", "share", "sunbeam")
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
				extension, ok := extensions[args[0]]
				if !ok {
					return nil, cobra.ShellCompDirectiveDefault
				}

				completions := make([]string, 0)
				for _, command := range extension.Commands {
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
			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				if args[0] == "--help" || args[0] == "-h" {
					return cmd.Help()
				}

				command, err := NewCmdCustom(extensions, args[0])
				if err != nil {
					return err
				}

				command.SetArgs(args[1:])
				command.Use = args[0]
				return command.Execute()
			}

			config, err := LoadConfig()
			if err != nil && !os.IsNotExist(err) {
				return err
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

			rootList := tui.NewRootList(extensions, config.Commands, history.entries)
			rootList.OnSelect = func(id string) {
				history.entries[id] = time.Now().Unix()
				_ = history.Save()
			}

			return tui.Draw(rootList, MaxHeigth)
		},
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

	return rootCmd, nil
}
