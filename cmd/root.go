package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/cli/browser"
	"github.com/google/shlex"
	"github.com/pomdtr/sunbeam/internal/tui"
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

type RooItems map[string]string

func getRootItems() (RooItems, error) {
	var candidates []string
	if env, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		candidates = append(candidates, filepath.Join(env, "sunbeam", "root.json"))
	}

	candidates = append(candidates, filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "root.json"))

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}

		f, err := os.Open(candidate)
		if err != nil {
			return nil, err
		}

		decoder := json.NewDecoder(f)
		var rootItems RooItems
		if err := decoder.Decode(&rootItems); err != nil {
			return nil, err
		}

		return rootItems, nil
	}

	return nil, nil
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
				for alias, extension := range extensions {
					completions = append(completions, fmt.Sprintf("%s\t%s", alias, extension))
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

			return nil, cobra.ShellCompDirectiveDefault
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

				commandPath, err := exec.LookPath(fmt.Sprintf("sunbeam-%s", args[0]))
				if err != nil {
					return fmt.Errorf("command %s not found", args[0])
				}

				command, err := NewExtensionCommand(commandPath)
				if err != nil {
					return err
				}

				command.SetArgs(args[1:])
				command.Use = args[0]
				return command.Execute()
			}

			rootItems, err := getRootItems()
			if err != nil {
				return err
			}
			if len(rootItems) == 0 {
				return cmd.Help()
			}

			items := make([]types.ListItem, 0)
			for title, command := range rootItems {
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
								Origin:  ref.Path,
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

			return tui.Draw(tui.NewRootList(tui.Extensions{}, items...), MaxHeigth)

		},
	}

	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(NewCmdServe())
	rootCmd.AddCommand(NewCmdFetch())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewCmdList())

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
	manifest, err := extensions.Get(extensionpath)
	if err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Draw(tui.NewRunner(extensions, tui.CommandRef{
				Path: extensionpath,
			}), MaxHeigth)
		},
		SilenceErrors: true,
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	for _, subcommand := range manifest.Commands {
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
					return tui.Draw(tui.NewRunner(extensions, tui.CommandRef{
						Path:    extensionpath,
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
					return browser.OpenURL(command.Url)
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
				subcmd.MarkFlagRequired(param.Name)
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

	path, err := exec.LookPath(fmt.Sprintf("sunbeam-%s", args[0]))
	if err != nil {
		return ref, fmt.Errorf("command %s not found", args[0])
	}

	ref.Path = path
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
