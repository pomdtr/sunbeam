package cmd

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/otiai10/copy"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewCmdExtension(extensionRoot string, extensions []app.Extension) *cobra.Command {
	extensionCommand := &cobra.Command{
		Use:     "extension",
		Aliases: []string{"extensions", "ext"},
		Short:   "Manage sunbeam extensions",
		GroupID: "core",
	}

	extensionNames := make([]string, len(extensions))
	for i, extension := range extensions {
		extensionNames[i] = extension.Name()
	}

	extensionCommand.AddCommand(func() *cobra.Command {
		cmd := cobra.Command{
			Use:   "install <name> ",
			Short: "Install a sunbeam extension from a local directory or a git repository",
			Args:  cobra.ExactArgs(1),
			PreRunE: func(cmd *cobra.Command, args []string) error {
				if !cmd.Flags().Changed("dir") && !cmd.Flags().Changed("git") && !cmd.Flags().Changed("url") {
					return fmt.Errorf("must specify one of --dir, --git, or --url")
				}

				extensionName := args[0]
				re, err := regexp.Compile(`^[\w-]+$`)
				if err != nil {
					return err
				}

				if !re.MatchString(extensionName) {
					return fmt.Errorf("extension name must be alphanumeric and contain only dashes and underscores")
				}

				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				extensionName := args[0]
				targetDir := path.Join(extensionRoot, extensionName)

				if cmd.Flags().Changed("dir") {
					dir, _ := cmd.Flags().GetString("dir")
					dir, err := filepath.Abs(dir)
					if err != nil {
						return fmt.Errorf("failed to get absolute path of directory: %w", err)
					}

					if err := installFromDir(dir, targetDir); err != nil {
						return err
					}
				} else if cmd.Flags().Changed("git") {
					git, _ := cmd.Flags().GetString("git")
					if err := installFromGit(git, targetDir); err != nil {
						return err
					}
				} else if cmd.Flags().Changed("url") {
					url, _ := cmd.Flags().GetString("url")
					if err := installFromURL(url, targetDir); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("must specify one of --dir, --git, or --url")
				}

				fmt.Printf("Installed extension %s", extensionName)
				return nil
			},
		}

		cmd.Flags().String("dir", "", "Directory to install extension from")
		cmd.Flags().String("git", "", "Git repository to install extension from")
		cmd.Flags().String("url", "", "URL to install extension from")
		cmd.MarkFlagsMutuallyExclusive("dir", "git", "url")

		return &cmd
	}())

	extensionCommand.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:       "remove",
			ValidArgs: extensionNames,
			Short:     "Remove an installed extension",
			RunE: func(cmd *cobra.Command, args []string) error {
				extensionPath := path.Join(extensionRoot, args[0])
				if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
					return fmt.Errorf("extension not found: %s", extensionPath)
				}

				if err := os.RemoveAll(extensionPath); err != nil {
					return fmt.Errorf("failed to remove extension: %s", err)
				}

				fmt.Println("Removed extension", args[0])
				return nil
			},
		}
	}())

	extensionCommand.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:   "rename [old name] [new name]",
			Short: "Rename an installed extension",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				oldPath := path.Join(extensionRoot, args[0])
				if !utils.FileExists(oldPath) {
					return fmt.Errorf("extension %s is not installed", args[0])
				}

				newPath := path.Join(extensionRoot, args[1])
				if utils.FileExists(newPath) {
					return fmt.Errorf("extension %s is already installed", args[1])
				}

				if err := copy.Copy(oldPath, newPath); err != nil {
					return fmt.Errorf("failed to rename extension: %s", err)
				}

				if err := os.RemoveAll(oldPath); err != nil {
					return fmt.Errorf("failed to remove old extension: %s", err)
				}

				return nil
			},
		}
	}())

	extensionCommand.AddCommand(func() *cobra.Command {
		command := &cobra.Command{
			Use:       "upgrade",
			Short:     "Upgrade installed extension",
			Args:      cobra.ExactArgs(1),
			ValidArgs: extensionNames,
			RunE: func(cmd *cobra.Command, args []string) error {
				extensionDir := path.Join(extensionRoot, args[0])
				fi, err := os.Lstat(extensionDir)
				if os.IsNotExist(err) {
					return fmt.Errorf("extension not found: %s", args[0])
				}

				if IsLocalExtension(fi) {
					return fmt.Errorf("cannot upgrade local extensions")
				}

				gc := utils.NewGitClient(extensionDir)

				currentVersion := gc.GetCurrentVersion()
				latestVersion, err := gc.GetLatestVersion()
				if err != nil {
					return err
				}

				if currentVersion == latestVersion {
					fmt.Printf("Extension %s is already up to date", args[0])
					return nil
				}

				if err := gc.Pull(); err != nil {
					return err
				}

				manifestPath := path.Join(extensionDir, "sunbeam.yml")
				if _, err = os.Stat(manifestPath); os.IsNotExist(err) {
					return fmt.Errorf("extension %s does not have a sunbeam.yml manifest", args[0])
				}

				if _, err := app.ParseManifest(manifestPath); err != nil {
					return fmt.Errorf("failed to parse manifest: %w", err)
				}

				return nil
			},
		}

		command.Flags().Bool("all", false, "Upgrade all installed extensions")
		command.Flags().Bool("dry-run", false, "Only dispay what would be upgraded")
		return command
	}())

	extensionCommand.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:     "list",
			Short:   "List installed extensions",
			Aliases: []string{"ls"},
			Args:    cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				for name := range extensionNames {
					fmt.Println(name)
				}
			},
		}
	}())

	return extensionCommand
}

func installFromDir(extensionDir string, targetDir string) error {
	if _, err := os.Stat(path.Join(extensionDir, "sunbeam.yml")); os.IsNotExist(err) {
		return fmt.Errorf("current directory is not a sunbeam extension")
	}

	if err := os.Symlink(extensionDir, targetDir); err != nil {
		return fmt.Errorf("failed to create symlink: %s", err)
	}

	return nil
}

func installFromGit(repository string, targetDir string) error {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "sunbeam")
	if err != nil {
		return err
	}

	err = utils.GitClone(repository, tmpDir)
	if err != nil {
		return err
	}

	manifestPath := path.Join(tmpDir, "sunbeam.yml")
	if _, err = os.Stat(manifestPath); os.IsNotExist(err) {
		return err
	}

	if _, err := app.ParseManifest(manifestPath); err != nil {
		return err
	}

	os.MkdirAll(path.Dir(targetDir), 0755)
	if err := copy.Copy(tmpDir, targetDir); err != nil {
		return err
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		return err
	}

	return nil
}

func installFromURL(url string, targetDir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download extension: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download extension: %s", resp.Status)
	}

	defer resp.Body.Close()

	var extension app.Extension
	if err := yaml.NewDecoder(resp.Body).Decode(&extension); err != nil {
		return fmt.Errorf("failed to parse extension manifest: %s", err)
	}

	if err := os.Mkdir(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create extension directory: %s", err)
	}

	manifestPath := path.Join(targetDir, "sunbeam.yml")
	f, err := os.Create(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to create extension manifest: %s", err)
	}

	if err := yaml.NewEncoder(f).Encode(extension); err != nil {
		return fmt.Errorf("failed to write extension manifest: %s", err)
	}

	return nil
}

func IsLocalExtension(fi fs.FileInfo) bool {
	// Check if root is a symlink
	return fi.Mode()&os.ModeSymlink != 0
}
