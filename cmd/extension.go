package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

func NewCmdExtension() *cobra.Command {
	extensionCommand := &cobra.Command{
		Use:     "extension",
		Aliases: []string{"extensions", "ext"},
		Short:   "Manage sunbeam extensions",
	}

	extensionArgs := make([]string, 0, len(app.Sunbeam.Extensions))
	for _, extension := range app.Sunbeam.Extensions {
		extensionArgs = append(extensionArgs, extension.Name)
	}

	extensionCommand.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:   "install <repository>",
			Short: "Install a sunbeam extension from a git repository",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if args[0] == "." {
					wd, err := os.Getwd()
					if err != nil {
						return err
					}

					if _, err = os.Stat(path.Join(wd, "sunbeam.yml")); os.IsNotExist(err) {
						return fmt.Errorf("current directory is not a sunbeam extension")
					}

					if err := app.Sunbeam.AddExtension(path.Base(wd), app.ExtensionConfig{
						Root: wd,
					}); err != nil {
						return err
					}

					fmt.Printf("Installed extension")
					return nil
				}

				repo, err := utils.ParseWithHost(args[0], "github.com")
				if err != nil {
					return err
				}

				tmpDir, err := os.MkdirTemp(os.TempDir(), "sunbeam")
				if err != nil {
					return err
				}

				err = utils.GitClone(repo.Url(), tmpDir)
				if err != nil {
					return err
				}

				manifestPath := path.Join(tmpDir, "sunbeam.yml")
				if _, err = os.Stat(manifestPath); os.IsNotExist(err) {
					return fmt.Errorf("extension %s does not have a sunbeam.yml manifest", repo.Name)
				}

				manifestBytes, err := os.ReadFile(manifestPath)
				if err != nil {
					return fmt.Errorf("failed to read manifest for extension %s", repo.Name)
				}
				manifest, err := app.ParseManifest(manifestBytes)
				if err != nil {
					return err
				}

				if manifest.PostInstall != "" {
					cmd := exec.Command("sh", "-c", manifest.PostInstall)
					cmd.Dir = tmpDir
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Stdin = os.Stdin
					if err != nil {
						return err
					}
				}

				target := path.Join(app.Sunbeam.ExtensionRoot, repo.Host, repo.Owner, repo.Name)
				os.MkdirAll(path.Dir(target), 0755)
				if err := os.Rename(tmpDir, target); err != nil {
					return err
				}

				if err := app.Sunbeam.AddExtension(manifest.Name, app.ExtensionConfig{
					Remote: repo.Url(),
					Root:   target,
				}); err != nil {
					return err
				}

				fmt.Printf("Installed extension %s", repo.Name)
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
				return app.Sunbeam.RemoveExtension(args[0])
			},
		}
	}())

	extensionCommand.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:       "upgrade",
			Short:     "Upgrade installed extension",
			Args:      cobra.ExactArgs(1),
			ValidArgs: extensionArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				extension, ok := app.Sunbeam.ExtensionConfigs[args[0]]
				if extension.Remote == "" {
					return fmt.Errorf("extension %s is not installed from a remote repository", args[0])
				}

				if !ok {
					return fmt.Errorf("extension %s not found", args[0])
				}
				gc := utils.NewGitClient(extension.Root)

				currentVersion := gc.GetCurrentVersion()
				latestVersion, err := gc.GetLatestVersion()
				if err != nil {
					return err
				}

				if currentVersion == latestVersion {
					fmt.Printf("Extension %s is already up to date", args[0])
					return nil
				}

				return gc.Pull()
			},
		}
	}())

	extensionCommand.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:     "list",
			Short:   "List installed extensions",
			Aliases: []string{"ls"},
			Args:    cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				homedir, _ := os.UserHomeDir()

				rows := make([][]string, 0, len(app.Sunbeam.Extensions))
				for name, extension := range app.Sunbeam.ExtensionConfigs {
					if extension.Remote != "" {
						gc := utils.NewGitClient(extension.Root)
						origin := gc.GetOrigin()
						repo, _ := utils.ParseWithHost(origin, "github.com")
						version := gc.GetCurrentVersion()
						rows = append(rows, []string{name, repo.FullName(), version[:7]})
					} else {
						dir := strings.Replace(extension.Root, homedir, "~", 1)
						rows = append(rows, []string{name, dir, ""})
					}
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

					if _, err := os.Stat(filepath.Join(app.Sunbeam.ExtensionRoot, "github.com", repo.FullName)); err == nil {
						item.Accessories = []string{
							"Installed",
						}

						item.Actions = []tui.Action{
							{
								Title: "Remove Extension",
								Cmd:   tui.NewExecCmd(fmt.Sprintf("sunbeam extension remove %s", repo.Name)),
							},
						}
					} else {
						item.Actions = []tui.Action{
							{
								Title: "Install Extension",
								Cmd:   tui.NewExecCmd(fmt.Sprintf("sunbeam extension install %s", repo.HtmlURL)),
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
				root := tui.NewModel(list)

				return tui.Draw(root)
			},
		}
		return &command
	}())
	return extensionCommand
}
