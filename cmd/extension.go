package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
)

func NewCmdExtension() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension",
		Aliases: []string{"ext", "extensions"},
		Short:   "Manage extensions",
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
			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(extensions)
			}

			for alias := range extensions {
				fmt.Println(alias)
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
		Use:     "install",
		Aliases: []string{"add"},
		Short:   "Install an extension",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionRoot := filepath.Join(dataHome(), "extensions")
			if err := os.MkdirAll(extensionRoot, 0755); err != nil {
				return err
			}

			if info, err := os.Stat(args[0]); err == nil && info.IsDir() {
				originDir, err := filepath.Abs(args[0])
				if err != nil {
					return err
				}
				entrypoint := filepath.Join(originDir, "sunbeam-extension")
				if _, err := os.Stat(entrypoint); err != nil {
					return fmt.Errorf("extension %s not found", args[0])
				}

				var alias string
				if flags.alias != "" {
					alias = flags.alias
				} else {
					alias = strings.TrimPrefix(filepath.Base(originDir), "sunbeam-")
				}

				extension, err := tui.LoadExtension(entrypoint)
				if err != nil {
					return err
				}

				extensionDir := filepath.Join(extensionRoot, alias)
				if err := os.MkdirAll(extensionDir, 0755); err != nil {
					return err
				}

				manifestPath := filepath.Join(extensionDir, "manifest.json")
				f, err := os.Create(manifestPath)
				if err != nil {
					return err
				}

				encoder := json.NewEncoder(f)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(extension.Manifest); err != nil {
					return err
				}

				return os.Symlink(originDir, filepath.Join(extensionDir, "src"))
			}

			// check if git is installed
			if _, err := exec.LookPath("git"); err != nil {
				return fmt.Errorf("git not found")
			}

			origin, err := url.Parse(args[0])
			if err != nil {
				return err
			}

			var alias string
			if flags.alias != "" {
				alias = flags.alias
			} else {
				alias = filepath.Base(origin.Path)
				alias = strings.TrimSuffix(alias, filepath.Ext(alias))
				alias = strings.TrimPrefix(alias, "sunbeam-")
			}

			tempdir, err := os.MkdirTemp("", "sunbeam-extension-")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tempdir)

			cloneCmd := exec.Command("git", "clone", args[0], "src")
			cloneCmd.Dir = tempdir
			cloneCmd.Stdout = os.Stdout
			cloneCmd.Stderr = os.Stderr
			if err := cloneCmd.Run(); err != nil {
				return err
			}

			extension, err := tui.LoadExtension(filepath.Join(tempdir, "src", "sunbeam-extension"))
			if err != nil {
				return err
			}

			manifestPath := filepath.Join(tempdir, "manifest.json")
			f, err := os.Create(manifestPath)
			if err != nil {
				return err
			}

			encoder := json.NewEncoder(f)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(extension.Manifest); err != nil {
				return err
			}

			if err := os.Rename(tempdir, filepath.Join(extensionRoot, alias)); err != nil {
				return err
			}

			return nil
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
				extensionDir := filepath.Join(dataHome(), "extensions", alias)
				if _, err := os.Stat(filepath.Join(extensionDir, "src", ".git")); err == nil {
					if _, err := exec.LookPath("git"); err != nil {
						return fmt.Errorf("git not found")
					}

					cmd.PrintErrln("Pulling changes...")
					pullCmd := exec.Command("git", "pull")
					pullCmd.Dir = filepath.Join(extensionDir, "src")
					pullCmd.Stdout = os.Stdout
					pullCmd.Stderr = os.Stderr
					err := pullCmd.Run()
					if err != nil {
						return err
					}
				}

				cmd.PrintErrf("Updating manifest for %s...\n", alias)
				extension, err := tui.LoadExtension(filepath.Join(extensionDir, "src", "sunbeam-extension"))
				if err != nil {
					return err
				}

				f, err := os.Create(filepath.Join(extensionDir, "manifest.json"))
				if err != nil {
					return err
				}
				defer f.Close()

				encoder := json.NewEncoder(f)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(extension.Manifest); err != nil {
					return err
				}

				cmd.PrintErrf("Done upgrading %s\n", alias)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&flags.all, "all", "a", false, "upgrade all extensions")

	return cmd
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
			extensionDir := filepath.Join(dataHome(), "extensions", args[0])
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
			extensionDir := filepath.Join(dataHome(), "extensions", args[0])
			newExtensionDir := filepath.Join(dataHome(), "extensions", args[1])

			return os.Rename(extensionDir, newExtensionDir)
		},
	}

	return cmd
}

func FindExtensions() (map[string]tui.Extension, error) {
	extensionRoot := filepath.Join(dataHome(), "extensions")
	if _, err := os.Stat(extensionRoot); err != nil {
		return nil, nil
	}

	entries, err := os.ReadDir(extensionRoot)
	if err != nil {
		return nil, err
	}
	extensionMap := make(map[string]tui.Extension)
	for _, entry := range entries {
		manifestPath := filepath.Join(extensionRoot, entry.Name(), "manifest.json")
		f, err := os.Open(manifestPath)
		if err != nil {
			continue
		}
		defer f.Close()

		var manifest types.Manifest
		if err := json.NewDecoder(f).Decode(&manifest); err != nil {
			continue
		}

		extensionMap[entry.Name()] = tui.Extension{
			Manifest:   manifest,
			Entrypoint: filepath.Join(extensionRoot, entry.Name(), "src", "sunbeam-extension"),
		}
	}

	return extensionMap, nil
}
