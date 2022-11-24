package cmd

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/cli/go-gh"
	"github.com/olekukonko/tablewriter"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/git"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewCmdExtension() *cobra.Command {
	extensionCommand := &cobra.Command{
		Use:     "extension",
		Aliases: []string{"extensions", "ext"},
		GroupID: "core",
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
			RunE: func(cmd *cobra.Command, args []string) (err error) {
				if _, err = os.Stat(app.Sunbeam.ExtensionRoot); os.IsNotExist(err) {
					os.MkdirAll(app.Sunbeam.ExtensionRoot, 0755)
				}

				if args[0] == "." {
					wd, err := os.Getwd()
					if err != nil {
						return err
					}

					name := filepath.Base(wd)
					targetLink := filepath.Join(app.Sunbeam.ExtensionRoot, name)
					if err := os.MkdirAll(filepath.Dir(targetLink), 0755); err != nil {
						return err
					}

					err = os.Symlink(wd, targetLink)
					if err != nil {
						return err
					}

					fmt.Printf("Installed extension %s", name)
					return nil
				}

				repo, err := git.ParseWithHost(args[0], "github.com")
				if err != nil {
					return err
				}

				target := path.Join(app.Sunbeam.ExtensionRoot, repo.Name)
				if _, err = os.Stat(target); err == nil {
					log.Fatalf("Extension %s already installed", repo.Name)
				}

				return git.Clone(repo.Host, repo.FullName(), target)
			},
		}
	}())

	extensionCommand.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:       "remove",
			ValidArgs: extensionArgs,
			Short:     "Remove an installed extension",
			RunE: func(cmd *cobra.Command, args []string) error {
				extension, ok := app.Sunbeam.Extensions[args[0]]
				if !ok {
					return fmt.Errorf("extension %s not found", args[0])
				}

				err := os.RemoveAll(extension.Dir())
				if err != nil {
					return err
				}
				fmt.Printf("Removed extension %s", args[0])
				return nil
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
				extension, ok := app.Sunbeam.Extensions[args[0]]
				if !ok {
					return fmt.Errorf("extension %s not found", args[0])
				}
				dir := extension.Dir()
				gc := git.NewGitClient(dir)

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
				rows := make([][]string, 0, len(app.Sunbeam.Extensions))
				for _, extension := range app.Sunbeam.Extensions {
					gc := git.NewGitClient(extension.Dir())
					origin := gc.GetOrigin()
					repo, _ := git.ParseWithHost(origin, "github.com")
					version := gc.GetCurrentVersion()
					rows = append(rows, []string{extension.Name, repo.FullName(), version[:7]})
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
		return &cobra.Command{
			Use:   "browse",
			Short: "Enter a UI for browsing and installing extensions",
			RunE: func(cmd *cobra.Command, args []string) (err error) {
				client, err := gh.RESTClient(nil)
				if err != nil {
					return err
				}
				res := struct {
					Items []struct {
						Name        string
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
						Title:    repo.Name,
						Subtitle: repo.Description,
						PreviewCmd: func() string {
							res := struct {
								Content string
							}{}
							err := client.Get(fmt.Sprintf("repos/%s/readme", repo.Name), &res)
							if err != nil {
								return err.Error()
							}
							content, err := base64.StdEncoding.DecodeString(res.Content)
							if err != nil {
								return err.Error()
							}
							return string(content)
						},
					}

					var primaryAction tui.Action
					if _, ok := app.Sunbeam.Extensions[repo.Name]; ok {
						primaryAction = tui.Action{
							Title: "Uninstall",
							Cmd:   tui.NewExecCmd(exec.Command("sunbeam", "extension", "remove", repo.Name)),
						}
					} else {
						primaryAction = tui.Action{
							Title: "Install",
							Cmd:   tui.NewExecCmd(exec.Command("sunbeam", "extension", "install", repo.HtmlURL)),
						}
					}

					item.Actions = []tui.Action{primaryAction, {
						Title: "Open in Browser",
						Cmd:   tui.NewOpenUrlCmd(repo.HtmlURL),
					}}

					extensionItems[i] = item
				}

				list := tui.NewList("Browse Extensions")
				list.ShowPreview = true
				list.SetItems(extensionItems)
				return tui.Draw(list, globalOptions)
			},
		}
	}())
	extensionCommand.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:   "create",
			Short: "Create a new extension",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
	}())
	return extensionCommand
}
