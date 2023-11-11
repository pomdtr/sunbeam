package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"mvdan.cc/sh/shell"
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
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		Version:      fmt.Sprintf("%s (%s)", Version, Date),
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootList := tui.NewRootList("Sunbeam", func() (extensions.ExtensionMap, []types.ListItem, error) {
				cfg, err := config.Load()
				if err != nil {
					return nil, nil, err
				}

				extensionMap := make(map[string]extensions.Extension)
				for alias, origin := range cfg.Extensions {
					extension, err := LoadExtension(alias, origin)
					if err != nil {
						continue
					}

					extensionMap[alias] = extension
				}

				return extensionMap, RootItems(cfg, extensionMap), nil
			})
			return tui.Draw(rootList)
		},
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
	rootCmd.AddCommand(NewCmdEdit())
	rootCmd.AddCommand(NewCmdCopy())
	rootCmd.AddCommand(NewCmdPaste())
	rootCmd.AddCommand(NewCmdUpgrade(cfg))
	rootCmd.AddCommand(NewCmdOpen())
	rootCmd.AddCommand(NewCmdConfigure(cfg))

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

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupExtension,
		Title: "Extension Commands:",
	})

	for alias, origin := range cfg.Extensions {
		extension, err := LoadExtension(alias, origin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading extension %s: %s\n", alias, err)
			continue
		}

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
		item, err := RootItem(rootItem, extensionMap)
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

func RootItem(item config.Oneliner, extensionMap map[string]extensions.Extension) (types.ListItem, error) {
	// extract args from the command
	args, err := shell.Fields(item.Command, os.Getenv)
	if err != nil {
		return types.ListItem{
			Id:          fmt.Sprintf("root - %s", item.Title),
			Title:       item.Title,
			Accessories: []string{"Oneliner"},
			Actions: []types.Action{
				{
					Title: item.Title,
					Type:  types.ActionTypeExec,
					Args:  []string{"sh", "-c", item.Command},
					Exit:  true,
				},
				{
					Title: "Copy Command",
					Key:   "c",
					Type:  types.ActionTypeCopy,
					Text:  item.Command,
				},
			},
		}, nil
	}

	if len(args) == 0 {
		return types.ListItem{}, fmt.Errorf("invalid command: %s", item.Command)
	}

	if args[0] != "sunbeam" {
		return types.ListItem{
			Id:          fmt.Sprintf("root - %s", item.Title),
			Title:       item.Title,
			Accessories: []string{"Oneliner"},
			Actions: []types.Action{
				{
					Title: item.Title,
					Type:  types.ActionTypeExec,
					Args:  []string{"sh", "-c", item.Command},
					Exit:  true,
				},
				{
					Title: "Copy Command",
					Key:   "c",
					Type:  types.ActionTypeCopy,
					Text:  item.Command,
				},
			},
		}, nil
	}

	switch args[1] {
	case "open", "edit":
		return types.ListItem{
			Id:          fmt.Sprintf("root - %s", item.Title),
			Title:       item.Title,
			Accessories: []string{"Oneliner"},
			Actions: []types.Action{
				{
					Title: "Run",
					Type:  types.ActionTypeExec,
					Args:  args,
					Exit:  true,
				},
				{
					Title: "Copy Command",
					Key:   "c",
					Type:  types.ActionTypeCopy,
					Text:  item.Command,
				},
			},
		}, nil
	default:
		if len(args) < 3 {
			return types.ListItem{}, fmt.Errorf("invalid command: %s", item.Command)
		}

		alias := args[1]
		extension, ok := extensionMap[alias]
		if !ok {
			return types.ListItem{}, fmt.Errorf("extension %s not found", alias)
		}

		command, ok := extension.Command(args[2])
		if !ok {
			return types.ListItem{}, fmt.Errorf("command %s not found", args[2])
		}

		params, err := ExtractParams(args[3:], command)
		if err != nil {
			return types.ListItem{}, err
		}

		return types.ListItem{
			Id:          fmt.Sprintf("%s - %s", alias, item.Title),
			Title:       item.Title,
			Accessories: []string{extension.Title},
			Actions: []types.Action{
				{
					Title:     item.Title,
					Type:      types.ActionTypeRun,
					Extension: args[1],
					Command:   command.Name,
					Params:    params,
					Exit:      true,
				},
				{
					Title: "Copy Command",
					Key:   "c",
					Type:  types.ActionTypeCopy,
					Text:  item.Command,
				},
			},
		}, nil
	}
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

func ExtractParams(args []string, command types.CommandSpec) (map[string]any, error) {
	params := make(map[string]any)
	for len(args) > 0 {
		if !strings.HasPrefix(args[0], "--") {
			return nil, fmt.Errorf("invalid argument: %s", args[0])
		}

		parts := strings.SplitN(args[0][2:], "=", 2)
		if len(parts) == 1 {
			input, ok := CommandParam(command, parts[0])
			if !ok {
				return nil, fmt.Errorf("unknown parameter: %s", parts[0])
			}

			switch input.Type {
			case types.InputCheckbox:
				params[parts[0]] = true
				args = args[1:]
			case types.InputTextField, types.InputPassword:
				if len(args) < 2 {
					return nil, fmt.Errorf("missing value for parameter: %s", parts[0])
				}

				params[parts[0]] = args[1]
				args = args[2:]
			}

			continue
		}

		spec, ok := CommandParam(command, parts[0])
		if !ok {
			return nil, fmt.Errorf("unknown parameter: %s", parts[0])
		}

		switch spec.Type {
		case types.InputTextField, types.InputPassword:
			params[parts[0]] = parts[1]
		case types.InputCheckbox:
			value, err := strconv.ParseBool(parts[1])
			if err != nil {
				return nil, err
			}
			params[parts[0]] = value
		}

		args = args[1:]
	}

	return params, nil
}

func CommandParam(command types.CommandSpec, name string) (types.Input, bool) {
	for _, param := range command.Inputs {
		if param.Name == name {
			return param, true
		}
	}

	return types.Input{}, false
}
