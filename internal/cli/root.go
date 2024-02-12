package cli

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	Version = "dev"
)

func IsSunbeamRunning() bool {
	return len(os.Getenv("SUNBEAM")) > 0
}

//go:embed embed/sunbeam.json
var configBytes []byte

func NewRootCmd() (*cobra.Command, error) {
	if IsSunbeamRunning() {
		return NewCmdStd(), nil
	}

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
	}

	rootCmd.AddCommand(NewCmdValidate())
	rootCmd.AddCommand(NewCmdStd())
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

	rootCmd.AddCommand(NewCmdInstall(cfg))
	rootCmd.AddCommand(NewCmdRemove(cfg))
	rootCmd.AddCommand(NewCmdUpgrade(cfg))
	rootCmd.AddCommand(NewCmdRun(cfg))

	extensionMap := make(map[string]extensions.Extension)
	for alias, extensionConfig := range cfg.Extensions {
		extension, err := extensions.LoadExtension(extensionConfig.Origin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading extension %s: %s\n", alias, err)
			continue
		}

		extension.Env = cfg.Env
		for k, v := range extensionConfig.Env {
			extension.Env[k] = v
		}

		extensionMap[alias] = extension

		command, err := NewCmdCustom(alias, extension)
		if err != nil {
			return nil, err
		}
		rootCmd.AddCommand(command)
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			encoder.SetEscapeHTML(false)

			return encoder.Encode(cfg)
		}
		history, err := history.Load(history.Path)
		if err != nil {
			return err
		}

		rootList := tui.NewRootList("Sunbeam", &history, RootListGenerator())
		return tui.Draw(rootList)
	}

	return rootCmd, nil
}

func RootListGenerator() func() (config.Config, []sunbeam.ListItem, error) {
	return func() (config.Config, []sunbeam.ListItem, error) {
		cfg, err := config.Load(config.Path)
		if err != nil {
			return config.Config{}, nil, err
		}

		var items []sunbeam.ListItem
		var root []sunbeam.Action
		root = append(root, cfg.Root...)

		for alias, extensionConfig := range cfg.Extensions {
			extension, err := extensions.LoadExtension(extensionConfig.Origin)
			if err != nil {
				return cfg, nil, err
			}

			for _, action := range extension.Manifest.Root {
				action.Extension = alias
				root = append(root, action)
			}
		}

		for _, commandRef := range root {
			extensionConfig, ok := cfg.Extensions[commandRef.Extension]
			if !ok {
				continue
			}

			extension, err := extensions.LoadExtension(extensionConfig.Origin)
			if err != nil {
				continue
			}

			extension.Env = cfg.Env
			for k, v := range extensionConfig.Env {
				extension.Env[k] = v
			}

			command, ok := extension.Command(commandRef.Command)
			if !ok {
				continue
			}

			title := commandRef.Title
			if title == "" {
				title = command.Title
			}

			item := sunbeam.ListItem{
				Id:          fmt.Sprintf("%s:%s:%s", commandRef.Extension, commandRef.Command, commandRef.Title),
				Title:       title,
				Accessories: []string{commandRef.Extension},
				Actions: []sunbeam.Action{
					{
						Title:     "Run",
						Extension: commandRef.Extension,
						Command:   commandRef.Command,
						Params:    commandRef.Params,
						Reload:    commandRef.Reload,
					},
				},
			}

			if !extensions.IsRemote(extensionConfig.Origin) {
				item.Actions = append(item.Actions, sunbeam.Action{
					Title:     "Edit",
					Extension: "std",
					Command:   "edit",
					Params: map[string]interface{}{
						"path": extensionConfig.Origin,
					},
					Reload: true,
				})
			}

			items = append(items, item)
		}
		return cfg, items, nil
	}
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
