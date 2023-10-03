package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmdExtension() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Manage extensions",
	}

	cmd.AddCommand(NewCmdExtensionList())
	cmd.AddCommand(NewCmdExtensionInstall())
	cmd.AddCommand(NewCmdExtensionUpgrade())
	cmd.AddCommand(NewCmdExtensionRemove())
	cmd.AddCommand(NewCmdExtensionRename())

	return cmd
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
				extensionDir, err := filepath.Abs(args[0])
				if err != nil {
					return err
				}
				entrypoint := filepath.Join(extensionDir, "sunbeam-extension")
				if _, err := os.Stat(entrypoint); err != nil {
					return fmt.Errorf("extension %s not found", args[0])
				}

				// create symlink to current directory
				var target string
				if flags.alias != "" {
					target = filepath.Join(extensionRoot, flags.alias)
				} else {
					target = filepath.Join(extensionRoot, strings.TrimPrefix(filepath.Base(extensionDir), "sunbeam-"))
				}

				return os.Symlink(extensionDir, target)
			}

			// check if git is installed
			if _, err := exec.LookPath("git"); err != nil {
				return fmt.Errorf("git not found")
			}

			tempdir, err := os.MkdirTemp("", "sunbeam-extension-")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tempdir)

			cloneCmd := exec.Command("git", "clone", args[0])
			cloneCmd.Dir = tempdir
			cloneCmd.Stdout = os.Stdout
			cloneCmd.Stderr = os.Stderr

			if err := cloneCmd.Run(); err != nil {
				return err
			}

			entries, err := os.ReadDir(tempdir)
			if err != nil {
				return err
			}

			extensionDir := entries[0]
			alias := strings.TrimPrefix(extensionDir.Name(), "sunbeam-")
			if flags.alias != "" {
				alias = flags.alias
			}

			if err := os.Rename(filepath.Join(tempdir, extensionDir.Name()), filepath.Join(extensionRoot, alias)); err != nil {
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
		Use:   "upgrade",
		Short: "Upgrade an extension",
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

			completions := make([]string, 0)
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
			if _, err := exec.LookPath("git"); err != nil {
				return fmt.Errorf("git not found")
			}

			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			toUpgrade := make([]string, 0)
			for alias, extensionPath := range extensions {
				extensionDir := filepath.Dir(extensionPath)
				if _, err := os.Stat(filepath.Join(extensionDir, ".git")); err != nil {
					continue
				}

				if flags.all {
					toUpgrade = append(toUpgrade, extensionDir)
					continue
				}

				if len(args) > 0 && args[0] == alias {
					toUpgrade = append(toUpgrade, extensionDir)
				}
			}

			if len(toUpgrade) == 0 {
				return nil
			}

			for _, extensionDir := range toUpgrade {
				pullCmd := exec.Command("git", "pull")
				pullCmd.Dir = extensionDir
				pullCmd.Stdout = os.Stdout
				pullCmd.Stderr = os.Stderr

				err := pullCmd.Run()
				if err != nil {
					return err
				}
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
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			extensions, err := FindExtensions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			completions := make([]string, 0)
			for alias := range extensions {
				completions = append(completions, alias)
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			extensionPath, ok := extensions[args[0]]
			if !ok {
				return fmt.Errorf("extension %s not found", args[0])
			}

			if info, err := os.Stat(extensionPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
				return os.Remove(extensionPath)
			}

			return os.RemoveAll(filepath.Dir(extensionPath))
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

			completions := make([]string, 0)
			for alias := range extensions {
				completions = append(completions, alias)
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			src, dst := args[0], args[1]
			extensionPath, ok := extensions[src]
			if !ok {
				return fmt.Errorf("extension %s not found", src)
			}

			extensionDir := filepath.Dir(extensionPath)
			extensionRoot := filepath.Dir(extensionDir)
			if err := os.Rename(extensionDir, filepath.Join(extensionRoot, dst)); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func FindExtensions() (map[string]string, error) {
	var dirs []string
	extensionDir := filepath.Join(dataHome(), "extensions")
	entries, err := os.ReadDir(extensionDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			dest, err := os.Readlink(filepath.Join(extensionDir, entry.Name()))
			if err != nil {
				continue
			}

			info, err := os.Stat(dest)
			if err != nil {
				continue
			}

			if !info.IsDir() {
				continue
			}

			dirs = append(dirs, filepath.Join(extensionDir, entry.Name()))
			continue
		}

		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(extensionDir, entry.Name()))
			continue
		}

	}

	extensions := make(map[string]string)
	for _, dir := range dirs {
		dir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			if entry.Name() != "sunbeam-extension" {
				continue
			}

			// check if file is executable
			if info, err := entry.Info(); err != nil || info.Mode()&0111 == 0 {
				continue
			}

			alias := filepath.Base(dir)
			if _, ok := extensions[alias]; ok {
				continue
			}

			extensions[alias] = filepath.Join(dir, entry.Name())
		}
	}

	return extensions, nil
}
