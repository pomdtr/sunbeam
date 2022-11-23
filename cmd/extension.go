package cmd

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/cli/go-gh"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func init() {
	extensionCommand.AddCommand(extensionBrowseCommand)
	extensionCommand.AddCommand(extensionCreateCommand)
	extensionCommand.AddCommand(extensionExecCommand)
	extensionCommand.AddCommand(extensionInstallCommand)
	extensionCommand.AddCommand(extensionListCommand)
	extensionCommand.AddCommand(extensionManageCommand)
	extensionCommand.AddCommand(extensionRemoveCommand)
	extensionCommand.AddCommand(extensionSearchCommand)
	extensionCommand.AddCommand(extensionUpgradeCommand)

	rootCmd.AddCommand(extensionCommand)
}

var extensionCommand = &cobra.Command{
	Use:     "extension",
	GroupID: "core",
	Short:   "Manage sunbeam extensions",
}

var extensionInstallCommand = &cobra.Command{
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

		// Remote repository
		url, err := url.Parse(args[0])
		if err != nil {
			log.Fatalln(err)
		}

		name := path.Base(url.Path)
		target := path.Join(app.Sunbeam.ExtensionRoot, name)
		if _, err = os.Stat(target); err == nil {
			log.Fatalf("Extension %s already installed", name)
		}

		command := exec.Command("git", "clone", url.String())
		command.Dir = app.Sunbeam.ExtensionRoot
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		err = command.Run()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Installed extension", name)
		return nil
	},
}

var extensionExecCommand = &cobra.Command{
	Use:   "exec",
	Short: "Execute an installed extension",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var extensionRemoveCommand = &cobra.Command{
	Use:   "remove",
	Short: "Remove an installed extension",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := filepath.Join(app.Sunbeam.ExtensionRoot, args[0])
		if _, err := os.Lstat(targetDir); os.IsNotExist(err) {
			return fmt.Errorf("no extension found: %q", targetDir)
		}

		err := os.RemoveAll(targetDir)
		if err != nil {
			return err
		}
		fmt.Printf("Removed extension %s", args[0])
		return nil
	},
}

var extensionSearchCommand = &cobra.Command{
	Use:   "search",
	Short: "Search for extensions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var extensionUpgradeCommand = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade installed extensions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var extensionListCommand = &cobra.Command{
	Use:   "list",
	Short: "List installed extensions",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, extension := range app.Sunbeam.Extensions {
			fmt.Println(extension.Name)
		}
		return nil
	},
}

var extensionBrowseCommand = &cobra.Command{
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

var extensionCreateCommand = &cobra.Command{
	Use:   "create",
	Short: "Create a new extension",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var extensionManageCommand = &cobra.Command{
	Use:   "manage",
	Short: "Enter a UI for managing installed extensions",
	RunE: func(cmd *cobra.Command, args []string) error {
		extensionItems := make([]tui.ListItem, 0)
		for _, extension := range app.Sunbeam.Extensions {
			extensionItems = append(extensionItems, tui.ListItem{
				Title:    extension.Title,
				Subtitle: extension.Name,
			})
		}

		list := tui.NewList("Manage Extensions")
		list.SetItems(extensionItems)

		return tui.Draw(list, globalOptions)
	},
}
