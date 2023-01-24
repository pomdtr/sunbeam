package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/olekukonko/tablewriter"
	"github.com/otiai10/copy"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

func NewCmdExtension(api app.Api, config *tui.Config) *cobra.Command {
	extensionCommand := &cobra.Command{
		Use:     "extension",
		Aliases: []string{"extensions", "ext"},
		Short:   "Manage sunbeam extensions",
		GroupID: "core",
	}

	extensionArgs := make([]string, 0, len(api.Extensions))
	for name := range api.Extensions {
		extensionArgs = append(extensionArgs, name)
	}

	extensionCommand.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:   "install <name> <directory-or-url>",
			Short: "Install a sunbeam extension from a local directory or a git repository",
			Args:  cobra.ExactArgs(2),
			PreRunE: func(cmd *cobra.Command, args []string) error {
				extensionName := args[0]
				invalidNames := []string{"extension", "check", "query", "run", "serve"}
				for _, name := range invalidNames {
					if extensionName == name {
						return fmt.Errorf("extension name %s is reserved", extensionName)
					}
				}

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
				extensionRoot := args[1]
				targetDir := path.Join(api.ExtensionRoot, extensionName)
				if _, err := os.Stat(targetDir); err == nil {
					return fmt.Errorf("extension %s is already installed at %s", extensionName, targetDir)
				}

				if _, err := os.Stat(extensionRoot); err == nil {
					extensionRoot, err = filepath.Abs(extensionRoot)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to get absolute path for extension root: %s", err)
						os.Exit(1)
					}

					if _, err = os.Stat(path.Join(extensionRoot, "sunbeam.yml")); os.IsNotExist(err) {
						return fmt.Errorf("current directory is not a sunbeam extension")
					}

					symlinkTarget := path.Join(api.ExtensionRoot, extensionName)

					if err := os.Symlink(extensionRoot, symlinkTarget); err != nil {
						fmt.Fprintln(os.Stderr, "Failed to create symlink", err)
						os.Exit(1)
					}

					fmt.Println("Installed extension", extensionName)
					return nil
				}

				tmpDir, err := os.MkdirTemp(os.TempDir(), "sunbeam")
				if err != nil {
					return err
				}

				err = utils.GitClone(extensionRoot, tmpDir)
				if err != nil {
					return err
				}

				manifestPath := path.Join(tmpDir, "sunbeam.yml")
				if _, err = os.Stat(manifestPath); os.IsNotExist(err) {
					return fmt.Errorf("extension %s does not have a sunbeam.yml manifest", extensionName)
				}

				extension, err := app.ParseManifest(manifestPath)
				if err != nil {
					return err
				}

				if err := PostInstallHook(extension); err != nil {
					return err
				}

				os.MkdirAll(path.Dir(targetDir), 0755)
				if err := copy.Copy(tmpDir, targetDir); err != nil {
					return err
				}

				if err := os.RemoveAll(tmpDir); err != nil {
					return err
				}

				fmt.Println("Installed extension", extensionName)
				return nil
			},
		}
	}())

	extensionCommand.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:       "remove",
			ValidArgs: extensionArgs,
			Short:     "Remove an installed extension",
			RunE: func(cmd *cobra.Command, args []string) error {
				extensionPath := path.Join(api.ExtensionRoot, args[0])
				if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
					fmt.Fprintln(os.Stderr, "Extension not found")
					os.Exit(1)
				}

				if err := os.RemoveAll(extensionPath); err != nil {
					fmt.Fprintln(os.Stderr, "Failed to remove extension")
					os.Exit(1)
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
				if !api.IsExtensionInstalled(args[0]) {
					return fmt.Errorf("extension %s is not installed", args[0])
				}

				if api.IsExtensionInstalled(args[1]) {
					return fmt.Errorf("extension %s is already installed", args[1])
				}

				oldPath := path.Join(api.ExtensionRoot, args[0])
				newPath := path.Join(api.ExtensionRoot, args[1])
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
			ValidArgs: extensionArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				extensionDir := path.Join(api.ExtensionRoot, args[0])
				fi, err := os.Lstat(extensionDir)
				if os.IsNotExist(err) {
					fmt.Fprintln(os.Stderr, "Extension not found")
					os.Exit(1)
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

				extension, err := app.ParseManifest(manifestPath)
				if err != nil {
					return fmt.Errorf("failed to parse manifest: %w", err)
				}

				if err := PostInstallHook(extension); err != nil {
					return err
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
				rows := make([][]string, 0)
				for name := range api.Extensions {
					rows = append(rows, []string{name})
				}

				writer := tablewriter.NewWriter(os.Stdout)
				writer.SetBorder(false)
				writer.SetColumnSeparator(" ")
				writer.AppendBulk(rows)
				writer.Render()
			},
		}
	}())

	extensionCommand.AddCommand(func() *cobra.Command {
		command := cobra.Command{
			Use:   "browse",
			Short: "Enter a UI for browsing and installing extensions",
			RunE: func(cmd *cobra.Command, args []string) (err error) {
				client := utils.NewGHClient("github.com")
				if err != nil {
					return err
				}
				res := struct {
					Items []struct {
						Name  string
						Owner struct {
							Login string
						}
						FullName    string `json:"full_name"`
						Description string
						HtmlURL     string `json:"html_url"`
					}
				}{}

				err = client.Get("search/repositories?q=topic:sunbeam-extension", &res)
				if err != nil {
					return err
				}

				extensionItems := make([]tui.ListItem, len(res.Items))
				for i, repo := range res.Items {
					item := tui.ListItem{
						Id:       strconv.Itoa(i),
						Title:    fmt.Sprintf("%s/%s", repo.Owner.Login, repo.Name),
						Subtitle: repo.Description,
					}

					if _, err := os.Stat(filepath.Join(api.ExtensionRoot, "github.com", repo.FullName)); err == nil {
						item.Accessories = []string{
							"Installed",
						}

						item.Actions = []tui.Action{
							{
								Title: "Remove Extension",
								Cmd: func() tea.Msg {
									exec.Command("sunbeam", "extension", "remove", repo.FullName).Run()
									return tea.Quit()
								},
							},
							{
								Title: "Open in Browser",
								Cmd:   tui.NewOpenUrlCmd(repo.HtmlURL),
							},
						}
					} else {
						item.Actions = []tui.Action{
							{
								Title: "Install Extension",
								Cmd: func() tea.Msg {
									exec.Command("sunbeam", "extension", "install", repo.FullName)
									return tea.Quit()
								},
							},
							{
								Title: "Open in Browser",
								Cmd:   tui.NewOpenUrlCmd(repo.HtmlURL),
							},
						}
					}

					extensionItems[i] = item
				}

				list := tui.NewList("Browse Extensions")
				list.SetItems(extensionItems)
				model := tui.NewModel(list)
				model.SetRoot(list)

				return tui.Draw(model, true)
			},
		}
		return &command
	}())

	return extensionCommand
}

func IsLocalExtension(fi fs.FileInfo) bool {
	// Check if root is a symlink
	return fi.Mode()&os.ModeSymlink != 0
}

func PostInstallHook(extension app.Extension) error {
	if extension.PostInstall == "" {
		return nil
	}
	cmd := exec.Command("sh", "-c", extension.PostInstall)
	cmd.Dir = extension.Root.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
