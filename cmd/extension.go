package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
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
	cmd.AddCommand(NewCmdExtensionInstall())
	cmd.AddCommand(NewCmdExtensionUpgrade())
	cmd.AddCommand(NewCmdExtensionRemove())
	cmd.AddCommand(NewCmdExtensionRename())
	cmd.AddCommand(NewCmdExtensionBrowse())

	return cmd
}

func NewCmdExtensionBrowse() *cobra.Command {
	return &cobra.Command{
		Use:   "browse",
		Short: "Browse extensions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return utils.Open("https://github.com/topics/sunbeam-extension")
		},
	}
}

func NewCmdExtensionList() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List installed extensions",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionMap, err := FindExtensions()
			if err != nil {
				return err
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(extensionMap)
			}

			for alias, extension := range extensionMap {
				fmt.Printf("%s\t%s\n", alias, extension.Title)
			}

			return nil
		},
	}
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
				alias := flags.alias
				if flags.alias == "" {
					return fmt.Errorf("must specify an alias for http extensions")
				}

				if err := validateAlias(flags.alias); err != nil {
					return err
				}

				return httpInstall(args[0], filepath.Join(extensionRoot, alias))
			} else if strings.HasPrefix(args[0], "github:") {
				origin := strings.TrimPrefix(args[0], "github:")
				parts := strings.Split(origin, "/")
				if len(parts) < 2 {
					return fmt.Errorf("invalid github origin")
				}
				owner, repo := parts[0], parts[1]
				entrypoint := filepath.Join(parts[2:]...)

				var alias string
				if flags.alias != "" {
					alias = flags.alias
				} else {
					alias = strings.TrimPrefix(repo, "sunbeam-")
				}

				if err := validateAlias(alias); err != nil {
					return err
				}

				return gitInstall(fmt.Sprintf("https://github.com/%s/%s", owner, repo), filepath.Join(extensionRoot, alias), entrypoint)
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
					alias = strings.TrimSuffix(alias, filepath.Ext(alias))
					alias = strings.TrimPrefix(alias, "sunbeam-")
				}

				if err := validateAlias(alias); err != nil {
					return err
				}

				return localInstall(origin, filepath.Join(extensionRoot, alias))
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
		Use:     "upgrade",
		Aliases: []string{"update"},
		Short:   "Upgrade an extension",
		Args:    cobra.MaximumNArgs(1),
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
			// check if git is installed
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

	if metadata.Type == extensions.ExtensionTypeLocal || metadata.Type == extensions.ExtensionTypeHttp {
		return cacheManifest(metadata.Entrypoint, filepath.Join(extensionDir, "manifest.json"))
	} else if metadata.Type == extensions.ExtensionTypeGit {
		pullCmd := exec.Command("git", "pull")
		pullCmd.Dir = filepath.Join(extensionDir, "src")
		pullCmd.Stderr = os.Stderr
		if err := pullCmd.Run(); err != nil {
			return err
		}

		return cacheManifest(filepath.Join(extensionDir, "src", "sunbeam-extension"), filepath.Join(extensionDir, "manifest.json"))
	} else {
		return fmt.Errorf("unknown extension type")
	}
}

func NewCmdExtensionRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove an extension",
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
			extensionDir := filepath.Join(extensionRoot, args[0])
			return os.RemoveAll(extensionDir)
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
		manifestBytes, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var manifest types.Manifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			continue
		}

		extensionMap[entry.Name()] = extensions.Extension{
			Manifest:   manifest,
			Entrypoint: entrypoint,
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
	cmd := exec.Command(entrypoint)
	cmd.Dir = filepath.Dir(entrypoint)

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
		Manifest:   manifest,
		Entrypoint: entrypoint,
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

	info, err := os.Stat(origin)
	if err != nil {
		return err
	}

	var entrypoint string
	if info.IsDir() {
		entrypoint = filepath.Join(origin, "sunbeam-extension")
	} else {
		entrypoint = origin
	}

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

	return cacheManifest(entrypoint, filepath.Join(extensionDir, "manifest.json"))
}

func gitInstall(origin string, extensionDir string, entrypoint string) (err error) {
	if entrypoint == "" {
		entrypoint = "sunbeam-extension"
	}

	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(extensionDir)
		}
	}()

	cloneCmd := exec.Command("git", "clone", "--depth=1", origin, filepath.Join(extensionDir, "src"))
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return err
	}

	metadataFile, err := os.Create(filepath.Join(extensionDir, "metadata.json"))
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(metadataFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(extensions.Metadata{
		Type:       extensions.ExtensionTypeGit,
		Origin:     origin,
		Entrypoint: filepath.Join("src", entrypoint),
	}); err != nil {
		return err
	}

	entrypoint = filepath.Join(extensionDir, "src", entrypoint)
	if info, err := os.Stat(entrypoint); err != nil {
		return fmt.Errorf("entrypoint %s not found", entrypoint)
	} else if info.IsDir() {
		entrypoint = filepath.Join(entrypoint, "sunbeam-extension")
		if _, err := os.Stat(entrypoint); err != nil {
			return fmt.Errorf("entrypoint %s not found", entrypoint)
		}
	}

	return cacheManifest(entrypoint, filepath.Join(extensionDir, "manifest.json"))
}

func httpInstall(origin string, extensionDir string) (err error) {
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(extensionDir)
		}
	}()

	resp, err := http.Get(origin)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading extension: %s", resp.Status)
	}

	entrypointPath := filepath.Join(extensionDir, "sunbeam-extension")
	entrypointFile, err := os.Create(entrypointPath)
	if err != nil {
		return err
	}

	if err := resp.Write(entrypointFile); err != nil {
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
		Origin:     origin,
		Entrypoint: filepath.Join(extensionDir, "sunbeam-extension"),
	}); err != nil {
		return err
	}

	return cacheManifest(entrypointPath, filepath.Join(extensionDir, "manifest.json"))
}
