package cli

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/github"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/types"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewCmdExtension(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension",
		Short:   "Manage sunbeam extensions",
		GroupID: CommandGroupCore,
	}

	cmd.AddCommand(NewCmdExtensionInstall(cfg))
	cmd.AddCommand(NewCmdExtensionUpgrade(cfg))
	cmd.AddCommand(NewCmdExtensionRename(cfg))
	cmd.AddCommand(NewCmdExtensionList(cfg))
	cmd.AddCommand(NewCmdExtensionRemove(cfg))
	cmd.AddCommand(NewCmdExtensionConfigure(cfg))
	cmd.AddCommand(NewCmdExtensionPublish())
	cmd.AddCommand(NewCmdExtensionCreate())

	return cmd
}

func extractAlias(origin string) (string, error) {
	originUrl, err := url.Parse(origin)
	if err != nil {
		return "", fmt.Errorf("failed to parse origin: %w", err)
	}

	base := filepath.Base(originUrl.Path)

	return strings.TrimSuffix(base, filepath.Ext(base)), nil
}

func normalizeOrigin(origin string) (string, error) {
	if strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://") {
		return origin, nil
	}

	if _, err := os.Stat(origin); err != nil {
		return "", fmt.Errorf("failed to find origin: %w", err)
	}

	abs, err := filepath.Abs(origin)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return strings.Replace(abs, os.Getenv("HOME"), "~", 1), nil
}

//go:embed embed/extension.py
var pythonExtBytes []byte

//go:embed embed/extension.ts
var denoExtBytes []byte

//go:embed embed/extension.sh
var shExtBytes []byte

func NewCmdExtensionCreate() *cobra.Command {
	var flags struct {
		language string
	}

	cmd := &cobra.Command{
		Use:     "create <name>",
		Short:   "Create a new sunbeam extension",
		Aliases: []string{"new"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var language string
			if flags.language != "" {
				language = flags.language
			} else {
				cmd.Println("No language specified, guessing...")
				switch filepath.Ext(args[0]) {
				case ".py":
					cmd.Println("Detected python extension")
					language = "python"
				case ".ts":
					cmd.Println("Detected deno extension")
					language = "deno"
				case ".sh":
					cmd.Println("Detected shell extension")
					language = "sh"
				default:
					cmd.Println("No language detected, defaulting to shell")
					language = "sh"
				}
			}

			cmd.Printf("Creating extension %s\n", args[0])
			f, err := os.Create(args[0])
			if err != nil {
				return fmt.Errorf("failed to create extension: %w", err)
			}
			defer f.Close()

			switch language {
			case "python":
				if _, err := f.Write(pythonExtBytes); err != nil {
					return fmt.Errorf("failed to write extension: %w", err)
				}

			case "deno":
				if _, err := f.Write(denoExtBytes); err != nil {
					return fmt.Errorf("failed to write extension: %w", err)
				}

			case "sh":
				if _, err := f.Write(shExtBytes); err != nil {
					return fmt.Errorf("failed to write extension: %w", err)
				}
			}

			if err := os.Chmod(args[0], 0755); err != nil {
				return fmt.Errorf("failed to chmod extension: %w", err)
			}

			cmd.Printf("✅ Created %s extension %s\n", language, args[0])

			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.language, "language", "l", "", "language of extension")
	_ = cmd.RegisterFlagCompletionFunc("language", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"sh", "python", "deno"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func NewCmdExtensionInstall(cfg config.Config) *cobra.Command {
	var flags struct {
		Alias string
	}

	cmd := &cobra.Command{
		Use:     "install <origin>",
		Short:   "Install sunbeam extensions",
		Aliases: []string{"add"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			origin, err := normalizeOrigin(args[0])
			if err != nil {
				return fmt.Errorf("failed to normalize origin: %w", err)
			}

			var alias string
			if flags.Alias != "" {
				alias = flags.Alias
			} else {
				a, err := extractAlias(origin)
				if err != nil {
					return fmt.Errorf("failed to get alias: %w", err)
				}
				alias = a
			}

			if _, err := extensions.LoadExtension(origin); err != nil {
				return fmt.Errorf("failed to load extension: %w", err)
			}

			if _, ok := cfg.Extensions[alias]; ok {
				return fmt.Errorf("extension %s already exists", alias)
			}

			cfg.Extensions[alias] = extensions.Config{
				Origin: origin,
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			cmd.Printf("✅ Installed %s\n", alias)
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.Alias, "alias", "", "alias for extension")

	return cmd

}

func NewCmdExtensionRename(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "rename <alias> <new-alias>",
		Short:   "Rename sunbeam extensions",
		Aliases: []string{"mv"},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return cfg.Aliases(), cobra.ShellCompDirectiveNoFileComp
			}

			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := cfg.Extensions[args[1]]; ok {
				return fmt.Errorf("extension %s already exists", args[1])
			}

			extension, ok := cfg.Extensions[args[0]]
			if !ok {
				return fmt.Errorf("extension %s not found", args[0])
			}

			delete(cfg.Extensions, args[0])
			cfg.Extensions[args[1]] = extension

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			cmd.Printf("✅ Renamed %s to %s\n", args[0], args[1])
			return nil
		},
	}
}

func NewCmdExtensionUpgrade(cfg config.Config) *cobra.Command {
	flags := struct {
		All bool
	}{}

	cmd := &cobra.Command{
		Use:       "upgrade",
		Short:     "Upgrade sunbeam extensions",
		ValidArgs: cfg.Aliases(),
		Args:      cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !flags.All {
				return fmt.Errorf("either provide an extension or use --all")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				extension, ok := cfg.Extensions[args[0]]
				if !ok {
					return fmt.Errorf("extension %s not found", args[0])
				}

				if err := extensions.Upgrade(extension); err != nil {
					return fmt.Errorf("failed to upgrade extension: %w", err)
				}

				cmd.Printf("✅ Upgraded %s\n", args[0])
				return nil
			}

			cmd.Printf("Upgrading %d extensions...\n\n", len(cfg.Extensions))
			for alias, extension := range cfg.Extensions {
				if err := extensions.Upgrade(extension); err != nil {
					return fmt.Errorf("failed to upgrade extension %s: %w", alias, err)
				}

				cmd.Printf("✅ Upgraded %s\n", alias)
			}

			cmd.Printf("\n✅ Upgraded all extensions\n")
			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.All, "all", false, "upgrade all extensions")
	return cmd
}

func NewCmdExtensionList(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List sunbeam extensions",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var t tableprinter.TablePrinter
			if isatty.IsTerminal(os.Stdout.Fd()) {
				w, _, err := term.GetSize(int(os.Stdout.Fd()))
				if err != nil {
					return err
				}
				t = tableprinter.New(os.Stdout, true, w)
			} else {
				t = tableprinter.New(os.Stdout, false, 0)
			}

			for alias, extension := range cfg.Extensions {
				t.AddField(alias)
				t.AddField(extension.Origin)
				t.EndRow()
			}

			return t.Render()
		},
	}

	return cmd
}

func NewCmdExtensionRemove(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <alias>",
		Short:   "Remove sunbeam extensions",
		Aliases: []string{"rm", "uninstall"},
		Args:    cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			completions := cfg.Aliases()

			for _, arg := range args {
				for i, completion := range completions {
					if completion == arg {
						completions = append(completions[:i], completions[i+1:]...)
						break
					}
				}
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				delete(cfg.Extensions, arg)
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			if len(args) == 1 {
				cmd.Printf("✅ Removed %s\n", args[0])
				return nil
			}

			cmd.Printf("✅ Removed %d extensions\n", len(args))
			return nil
		},
	}
}

func NewCmdExtensionPublish() *cobra.Command {
	var flags struct {
		Public bool
		Open   bool
	}

	cmd := &cobra.Command{
		Use:   "publish <script>",
		Short: "Publish a script as a github gist",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := extensions.ExtractManifest(args[0]); err != nil {
				return fmt.Errorf("extension is invalid: %w", err)
			}

			filename := filepath.Base(args[0])
			content, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read script: %w", err)
			}

			var token string
			if env, ok := os.LookupEnv("SUNBEAM_GITHUB_TOKEN"); ok {
				token = env
			} else if env, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
				token = env
			} else {
				return fmt.Errorf("github token not found, please set SUNBEAM_GITHUB_TOKEN or GITHUB_TOKEN")
			}

			gistClient := github.NewGistClient(token)

			gist, err := gistClient.CreateGist(filename, content, flags.Public)
			if err != nil {
				return fmt.Errorf("failed to publish script: %w", err)
			}

			rawUrl := fmt.Sprintf("https://gist.githubusercontent.com/%s/%s/raw/%s", gist.Owner.Login, gist.ID, url.PathEscape(filename))
			installCmd := fmt.Sprintf("sunbeam extension install %s", rawUrl)

			if err := gistClient.PatchGistDescription(gist.ID, installCmd); err != nil {
				return fmt.Errorf("failed to patch gist description: %w", err)
			}

			if flags.Open {
				return utils.Open(gist.HtmlURL)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				fmt.Print(rawUrl)
				return nil
			}

			var t tableprinter.TablePrinter
			w, _, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				return err
			}
			t = tableprinter.New(os.Stdout, true, w)

			t.AddField("Gist URL")
			t.AddField(gist.HtmlURL)
			t.EndRow()

			t.AddField("Raw URL")
			t.AddField(rawUrl)
			t.EndRow()

			t.AddField("Install Command")
			t.AddField(installCmd)
			t.EndRow()

			return t.Render()
		},
	}

	cmd.Flags().BoolVarP(&flags.Public, "public", "p", false, "make gist public")
	cmd.Flags().BoolVarP(&flags.Open, "open", "o", false, "open gist in browser")

	return cmd
}

func NewCmdExtensionConfigure(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:       "configure <alias>",
		Short:     "Configure extension preferences",
		Aliases:   []string{"config"},
		Args:      cobra.ExactArgs(1),
		ValidArgs: cfg.Aliases(),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionConfig, ok := cfg.Extensions[args[0]]
			if !ok {
				return fmt.Errorf("extension %s not found", args[0])
			}

			extension, err := extensions.LoadExtension(extensionConfig.Origin)
			if err != nil {
				return fmt.Errorf("failed to load extension: %w", err)
			}

			if len(extension.Manifest.Preferences) == 0 {
				return fmt.Errorf("extension %s has no preferences", args[0])
			}

			var inputs []types.Input
			for _, input := range extension.Manifest.Preferences {
				input.Default = extensionConfig.Preferences[input.Name]
				input.Required = true
				inputs = append(inputs, input)
			}

			form := tui.NewForm(func(m map[string]any) tea.Msg {
				extensionConfig.Preferences = m
				cfg.Extensions[args[0]] = extensionConfig
				if err := cfg.Save(); err != nil {
					return err
				}

				return tui.ExitMsg{}
			}, inputs...)

			return tui.Draw(form)
		},
	}
}
