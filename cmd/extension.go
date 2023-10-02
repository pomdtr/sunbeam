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

	cmd.AddCommand(NewCmdList())
	cmd.AddCommand(NewCmdInstall())
	cmd.AddCommand(NewCmdUpgrade())
	cmd.AddCommand(NewCmdRemove())
	cmd.AddCommand(NewCmdRename())

	return cmd
}

func NewCmdList() *cobra.Command {
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

			for alias, extension := range extensions {
				fmt.Printf("%s\t%s\n", alias, extension)
			}

			return nil
		},
	}
}

func NewCmdInstall() *cobra.Command {
	flags := struct {
		alias string
	}{}

	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"add"},
		Short:   "Install an extension",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionsDir := filepath.Join(dataHome(), "extensions")
			if err := os.MkdirAll(extensionsDir, 0755); err != nil {
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
					target = filepath.Join(extensionsDir, flags.alias)
				} else {
					target = filepath.Join(extensionsDir, strings.TrimPrefix(filepath.Base(extensionDir), "sunbeam-"))
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

			alias := strings.TrimPrefix(entries[0].Name(), "sunbeam-")
			if flags.alias != "" {
				alias = flags.alias
			}

			if err := os.Rename(filepath.Join(tempdir, entries[0].Name()), filepath.Join(extensionsDir, alias)); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.alias, "alias", "a", "", "alias for extension")
	return cmd
}

func NewCmdUpgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade an extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			extensionPath, ok := extensions[args[0]]
			if !ok {
				return fmt.Errorf("extension %s not found", args[0])
			}

			// check if .git directory exists
			extensionDir := filepath.Dir(extensionPath)
			if _, err := os.Stat(filepath.Join(extensionDir, ".git")); err != nil {
				return fmt.Errorf("extension %s is not installed with git", args[0])
			}

			// check if git is installed
			if _, err := exec.LookPath("git"); err != nil {
				return fmt.Errorf("git not found")
			}

			pullCmd := exec.Command("git", "pull")
			pullCmd.Dir = extensionDir
			pullCmd.Stdout = os.Stdout
			pullCmd.Stderr = os.Stderr

			return pullCmd.Run()
		},
	}

	return cmd
}

func NewCmdRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove an extension",
		Args:  cobra.ExactArgs(1),
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

func NewCmdRename() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename an extension",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			extensionPath, ok := extensions[args[0]]
			if !ok {
				return fmt.Errorf("extension %s not found", args[0])
			}

			extensionDir := filepath.Dir(extensionPath)
			extensionsDir := filepath.Dir(extensionDir)
			if err := os.Rename(extensionDir, filepath.Join(extensionsDir, args[1])); err != nil {
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
