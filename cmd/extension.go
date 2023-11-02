package cmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/acarl005/stripansi"
	"github.com/cli/cli/pkg/findsh"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	extensionRoot = filepath.Join(utils.DataHome(), "extensions")
)

func NewCmdExtension() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension",
		Aliases: []string{"ext", "extensions"},
		Short:   "Manage extensions",
		GroupID: CommandGroupCore,
	}

	cmd.AddCommand(NewCmdExtensionList())
	cmd.AddCommand(NewCmdExtensionEdit())
	cmd.AddCommand(NewCmdExtensionInstall())
	cmd.AddCommand(NewCmdExtensionUpgrade())
	cmd.AddCommand(NewCmdExtensionRemove())
	cmd.AddCommand(NewCmdExtensionRename())
	cmd.AddCommand(NewCmdExtensionCreate())

	return cmd
}

func NewCmdExtensionList() *cobra.Command {
	flags := struct {
		json bool
	}{}
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List installed extensions",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionMap, err := FindExtensions()
			if err != nil {
				return err
			}

			if flags.json {
				extensions := make([]extensions.Extension, 0)
				for _, extension := range extensionMap {
					extensions = append(extensions, extension)
				}

				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(extensions)
			}

			isTTY := isatty.IsTerminal(os.Stdout.Fd())
			var table tableprinter.TablePrinter
			if isTTY {
				width, _, err := term.GetSize(int(os.Stdout.Fd()))
				if err != nil {
					return err
				}

				table = tableprinter.New(os.Stdout, true, width)
			} else {
				table = tableprinter.New(os.Stdout, false, 0)
			}

			for alias, extension := range extensionMap {
				table.AddField(alias)
				table.AddField(extension.Title)
				table.AddField(extension.Origin)
				table.EndRow()
			}

			return table.Render()
		},
	}

	cmd.Flags().BoolVar(&flags.json, "json", false, "output as json")
	return cmd
}

//go:embed templates/*
var embedFs embed.FS

func NewCmdExtensionCreate() *cobra.Command {
	flags := struct {
		language string
		output   string
	}{}

	cmd := &cobra.Command{
		Use:   "create <output>",
		Short: "Create an extension",
		Args:  cobra.ExactArgs(1),
		Example: heredoc.Doc(`
			sunbeam extension create my-extension.sh
			sunbeam extension create --language sh my-extension
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			var templatePath string
			switch flags.language {
			case "sh":
				templatePath = "templates/extension.sh"
			case "python":
				templatePath = "templates/extension.py"
			case "deno":
				templatePath = "templates/extension.ts"
			default:
				extension := filepath.Ext(args[0])
				switch extension {
				case ".sh":
					templatePath = "templates/extension.sh"
				case ".py":
					templatePath = "templates/extension.py"
				case ".ts":
					templatePath = "templates/extension.ts"
				default:
					return fmt.Errorf("language flag was not specified and could not determine language from extension")
				}
			}

			embedBytes, err := embedFs.ReadFile(templatePath)
			if err != nil {
				return err
			}

			if err := os.WriteFile(args[0], embedBytes, 0755); err != nil {
				return err
			}

			if err := os.Chmod(args[0], 0755); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.language, "language", "l", "", "language for extension")
	_ = cmd.RegisterFlagCompletionFunc("language", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"sh", "python", "deno"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func NewCmdExtensionInstall() *cobra.Command {
	flags := struct {
		alias string
	}{}

	cmd := &cobra.Command{
		Use:     "install <src>",
		Aliases: []string{"add"},
		Short:   "Install an extension",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			validateAlias := func(alias string) error {
				subCmds := cmd.Parent().Parent().Commands()
				for _, subCmd := range subCmds {
					if subCmd.Name() == alias {
						return fmt.Errorf("alias %s already exists", alias)
					}
				}

				return nil
			}

			if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") {
				origin, err := url.Parse(args[0])
				if err != nil {
					return err
				}

				alias := flags.alias
				if flags.alias == "" {
					alias = filepath.Base(origin.Path)
					if alias == "." || alias == "/" {
						return fmt.Errorf("could not determine alias, please specify with --alias")
					}

					alias = strings.TrimPrefix(alias, "sunbeam-")
					alias = strings.TrimSuffix(alias, filepath.Ext(alias))
				}

				if err := validateAlias(flags.alias); err != nil {
					return err
				}

				if err := httpInstall(origin, filepath.Join(extensionRoot, alias)); err != nil {
					return err
				}

				cmd.Printf("Installed %s\n", alias)
				return nil
			} else {
				origin, err := filepath.Abs(args[0])
				if err != nil {
					return err
				}

				var alias string
				if flags.alias != "" {
					alias = flags.alias
				} else {
					alias = filepath.Base(origin)
					alias = strings.TrimPrefix(alias, "sunbeam-")
					alias = strings.TrimSuffix(alias, filepath.Ext(alias))
				}

				if err := validateAlias(alias); err != nil {
					return err
				}

				if err := localInstall(origin, filepath.Join(extensionRoot, alias)); err != nil {
					return err
				}

				cmd.Printf("Installed %s\n", alias)
				return nil
			}
		},
	}

	cmd.Flags().StringVarP(&flags.alias, "alias", "a", "", "alias for extension")
	return cmd
}

func NewCmdExtensionUpgrade() *cobra.Command {
	flags := struct {
		all bool
	}{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an extension",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if flags.all {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			extensions, err := FindExtensions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for alias := range extensions {
				completions = append(completions, alias)
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.all && len(args) > 0 {
				return fmt.Errorf("cannot use --all with an extension")
			}

			if !flags.all && len(args) == 0 {
				return fmt.Errorf("must specify an extension or use --all")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			toUpgrade := make([]string, 0)
			if flags.all {
				extensions, err := FindExtensions()
				if err != nil {
					return err
				}

				for alias := range extensions {
					toUpgrade = append(toUpgrade, alias)
				}
			} else {
				toUpgrade = append(toUpgrade, args[0])
			}

			for _, alias := range toUpgrade {
				cmd.PrintErrln()
				cmd.PrintErrf("Upgrading %s...\n", alias)
				extensionDir := filepath.Join(extensionRoot, alias)

				if err := upgradeExtension(extensionDir); err != nil {
					cmd.PrintErrln(err)
					continue
				}

				cmd.PrintErrln("Done")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&flags.all, "all", "a", false, "upgrade all extensions")

	return cmd
}

func upgradeExtension(extensionDir string) error {
	metadataFile, err := os.Open(filepath.Join(extensionDir, "metadata.json"))
	if err != nil {
		return err
	}

	var metadata extensions.Metadata
	if err := json.NewDecoder(metadataFile).Decode(&metadata); err != nil {
		return err
	}

	switch metadata.Type {
	case extensions.ExtensionTypeLocal:
		return cacheManifest(metadata.Entrypoint, filepath.Join(extensionDir, "manifest.json"))
	case extensions.ExtensionTypeHttp:
		resp, err := http.Get(metadata.Origin)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error downloading extension: %s", resp.Status)
		}

		entrypointPath := filepath.Join(extensionDir, metadata.Entrypoint)
		f, err := os.OpenFile(entrypointPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}

		if _, err := io.Copy(f, resp.Body); err != nil {
			return err
		}

		return cacheManifest(entrypointPath, filepath.Join(extensionDir, "manifest.json"))
	default:
		return fmt.Errorf("unknown extension type")
	}
}

func NewCmdExtensionRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove an extension",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			extensions, err := FindExtensions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for alias := range extensions {
				completions = append(completions, alias)
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, alias := range args {
				extensionDir := filepath.Join(extensionRoot, alias)
				if err := os.RemoveAll(extensionDir); err != nil {
					return err
				}
			}

			return nil
		},
	}

	return cmd
}

func NewCmdExtensionRename() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename an extension",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			extensions, err := FindExtensions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for alias := range extensions {
				completions = append(completions, alias)
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionDir := filepath.Join(extensionRoot, args[0])
			newExtensionDir := filepath.Join(extensionRoot, args[1])
			return os.Rename(extensionDir, newExtensionDir)
		},
	}

	return cmd
}

func NewCmdExtensionEdit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit an extension",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			extensions, err := FindExtensions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for alias := range extensions {
				completions = append(completions, alias)
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionMap, err := FindExtensions()
			if err != nil {
				return err
			}

			extension, ok := extensionMap[args[0]]
			if !ok {
				return fmt.Errorf("extension %s not found", args[0])
			}

			if extension.Type != extensions.ExtensionTypeLocal {
				return fmt.Errorf("cannot edit extension %s", args[0])
			}

			editor := utils.FindEditor()
			editCmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", editor, extension.Entrypoint))
			editCmd.Stdin = os.Stdin
			editCmd.Stdout = os.Stdout
			editCmd.Stderr = os.Stderr

			return editCmd.Run()
		},
	}

	return cmd
}

func FindExtensions() (extensions.ExtensionMap, error) {
	extensionMap := make(map[string]extensions.Extension)
	if _, err := os.Stat(extensionRoot); err != nil {
		return extensionMap, nil
	}

	entries, err := os.ReadDir(extensionRoot)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		extensionDir := filepath.Join(extensionRoot, entry.Name())

		metadataBytes, err := os.ReadFile(filepath.Join(extensionDir, "metadata.json"))
		if err != nil {
			continue
		}

		var metadata extensions.Metadata
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			continue
		}

		var entrypoint string
		if filepath.IsAbs(metadata.Entrypoint) {
			entrypoint = metadata.Entrypoint
		} else {
			entrypoint = filepath.Join(extensionDir, metadata.Entrypoint)
		}

		manifestPath := filepath.Join(extensionDir, "manifest.json")
		shouldCache, err := IsNewer(entrypoint, manifestPath)
		if err != nil {
			continue
		}
		if shouldCache {
			if err := cacheManifest(entrypoint, manifestPath); err != nil {
				continue
			}
		}

		manifestBytes, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var manifest types.Manifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			continue
		}

		extensionMap[entry.Name()] = extensions.Extension{
			Manifest: manifest,
			Metadata: metadata,
		}
	}

	return extensionMap, nil
}

func cacheManifest(entrypoint string, manifestPath string) error {
	extension, err := LoadExtension(entrypoint)
	if err != nil {
		return err
	}

	f, err := os.Create(manifestPath)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(extension.Manifest)
}

func LoadExtension(entrypoint string) (extensions.Extension, error) {
	var args []string
	if runtime.GOOS == "windows" {
		sh, err := findsh.Find()
		if err != nil {
			return extensions.Extension{}, err
		}
		args = []string{sh, "-c", `command "$@"`, "--", entrypoint}
	} else {
		args = []string{entrypoint}
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = filepath.Dir(entrypoint)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SUNBEAM=1")

	manifestBytes, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return extensions.Extension{}, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
		}

		return extensions.Extension{}, err
	}

	if err := schemas.ValidateManifest(manifestBytes); err != nil {
		return extensions.Extension{}, err
	}

	var manifest types.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return extensions.Extension{}, err
	}

	return extensions.Extension{
		Manifest: manifest,
		Metadata: extensions.Metadata{
			Type:       extensions.ExtensionTypeLocal,
			Origin:     entrypoint,
			Entrypoint: entrypoint,
		},
	}, nil
}

func localInstall(origin string, extensionDir string) (err error) {
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(extensionDir)
		}
	}()

	entrypoint := origin
	metadataFile, err := os.Create(filepath.Join(extensionDir, "metadata.json"))
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(metadataFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(extensions.Metadata{
		Type:       extensions.ExtensionTypeLocal,
		Origin:     origin,
		Entrypoint: entrypoint,
	}); err != nil {
		return err
	}

	if err := os.Chmod(entrypoint, 0755); err != nil {
		return err
	}

	return cacheManifest(entrypoint, filepath.Join(extensionDir, "manifest.json"))
}

func httpInstall(origin *url.URL, extensionDir string) (err error) {
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(extensionDir)
		}
	}()

	resp, err := http.Get(origin.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading extension: %s", resp.Status)
	}

	var filename string
	extension := filepath.Ext(filepath.Base(origin.Path))
	if extension != "" {
		filename = fmt.Sprintf("extension%s", extension)
	} else {
		filename = "extension"
	}

	entrypointPath := filepath.Join(extensionDir, filename)
	entrypointFile, err := os.Create(entrypointPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(entrypointFile, resp.Body); err != nil {
		return err
	}

	if err := os.Chmod(entrypointPath, 0755); err != nil {
		return err
	}

	metadataFile, err := os.Create(filepath.Join(extensionDir, "metadata.json"))
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(metadataFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(extensions.Metadata{
		Type:       extensions.ExtensionTypeHttp,
		Origin:     origin.String(),
		Entrypoint: filename,
	}); err != nil {
		return err
	}

	return cacheManifest(entrypointPath, filepath.Join(extensionDir, "manifest.json"))
}

func IsNewer(pathA, pathB string) (bool, error) {
	if pathA == pathB {
		return true, nil
	}

	infoA, err := os.Stat(pathA)
	if err != nil {
		return false, err
	}

	infoB, err := os.Stat(pathB)
	if err != nil {
		return false, err
	}

	return infoA.ModTime().After(infoB.ModTime()), nil
}
