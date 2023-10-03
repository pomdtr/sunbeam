package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/google/shlex"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"muzzammil.xyz/jsonc"
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
					completions = append(completions, alias)
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

				command, err := NewExtensionCommand(commandPath)
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
					Accessories: []string{command},
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

func NewExtensionCommand(extensionpath string) (*cobra.Command, error) {
	extensions := tui.Extensions{}
	extension, err := extensions.Get(extensionpath)
	if err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{
		SilenceErrors: true,
	}

	if extension.Root != "" {
		rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := extension.Run(tui.CommandInput{
					Command: extension.Root,
				})

				if err != nil {
					return err
				}

				if _, err := os.Stdout.Write(output); err != nil {
					return err
				}

				return nil
			}

			return tui.Draw(tui.NewRunner(extensions, tui.CommandRef{
				Script:  extensionpath,
				Command: extension.Root,
			}), MaxHeigth)
		}
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	for _, subcommand := range extension.Commands {
		subcommand := subcommand
		subcmd := &cobra.Command{
			Use: subcommand.Name,
			RunE: func(cmd *cobra.Command, args []string) error {
				params := make(map[string]any)
				for _, param := range subcommand.Params {
					if !cmd.Flags().Changed(param.Name) {
						continue
					}
					switch param.Type {
					case types.ParamTypeString:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}
						params[param.Name] = value
					case types.ParamTypeBoolean:
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}
						params[param.Name] = value
					}
				}

				extension, err := extensions.Get(extensionpath)
				if err != nil {
					return err
				}

				if subcommand.Mode == types.CommandModeView {
					if !isatty.IsTerminal(os.Stdout.Fd()) {
						output, err := extension.Run(tui.CommandInput{
							Command: subcommand.Name,
							Params:  params,
						})

						if err != nil {
							return err
						}

						if _, err := os.Stdout.Write(output); err != nil {
							return err
						}
					}
					return tui.Draw(tui.NewRunner(extensions, tui.CommandRef{
						Script:  extensionpath,
						Command: subcommand.Name,
						Params:  params,
					}), MaxHeigth)
				}

				out, err := extension.Run(tui.CommandInput{
					Command: subcommand.Name,
					Params:  params,
				})
				if err != nil {
					return err
				}

				if len(out) == 0 {
					return nil
				}

				var command types.Command
				if err := jsonc.Unmarshal(out, &command); err != nil {
					return err
				}

				switch command.Type {
				case types.CommandTypeCopy:
					return clipboard.WriteAll(command.Text)
				case types.CommandTypeOpen:
					return utils.Open(command.Target, command.App)
				default:
					return nil
				}
			},
		}

		if subcommand.Hidden {
			subcmd.Hidden = true
		}

		for _, param := range subcommand.Params {
			switch param.Type {
			case types.ParamTypeString:
				subcmd.Flags().String(param.Name, "", param.Description)
			case types.ParamTypeBoolean:
				subcmd.Flags().Bool(param.Name, false, param.Description)
			}

			if !param.Optional {
				_ = subcmd.MarkFlagRequired(param.Name)
			}
		}

		rootCmd.AddCommand(subcmd)
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
		childPage, err := buildDoc(child)
		if err != nil {
			return "", err
		}
		out.WriteString(childPage)
	}

	return out.String(), nil
}

func LookupIntEnv(key string, fallback int) int {
	env, ok := os.LookupEnv(key)
	if !ok {
		return fallback

	}

	value, err := strconv.Atoi(env)
	if err != nil {
		return fallback
	}

	return value
}

func LookupBoolEnv(key string, fallback bool) bool {
	env, ok := os.LookupEnv(key)
	if !ok {
		return fallback

	}

	b, err := strconv.ParseBool(env)
	if err != nil {
		return fallback
	}

	return b
}

func ExtractCommand(shellCommand string) (tui.CommandRef, error) {
	var ref tui.CommandRef
	args, err := shlex.Split(shellCommand)
	if err != nil {
		return ref, err
	}

	if len(args) == 0 {
		return ref, fmt.Errorf("no command specified")
	}

	extensions, err := FindExtensions()
	if err != nil {
		return ref, err
	}

	path, ok := extensions[args[0]]
	if !ok {
		return ref, fmt.Errorf("extension %s not found", args[0])
	}

	ref.Script = path
	args = args[1:]

	if len(args) == 0 {
		return ref, nil
	}

	ref.Command = args[0]
	args = args[1:]

	if len(args) == 0 {
		return ref, nil
	}

	ref.Params = make(map[string]any)

	for len(args) > 0 {
		if !strings.HasPrefix(args[0], "--") {
			return ref, fmt.Errorf("invalid argument: %s", args[0])
		}

		arg := strings.TrimPrefix(args[0], "--")

		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			ref.Params[parts[0]] = parts[1]
			args = args[1:]
			continue
		}

		if len(args) == 1 {
			ref.Params[arg] = true
			args = args[1:]
			continue
		}

		if strings.HasPrefix(args[1], "--") {
			ref.Params[arg] = true
			args = args[1:]
			continue
		}

		ref.Params[arg] = args[1]
		args = args[2:]
	}

	return ref, nil
}
