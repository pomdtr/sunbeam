package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/pomdtr/sunbeam/api"
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
	RunE:  runExtensionInstall,
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
		terminal := term.FromEnv()
		_, width, err := terminal.Size()
		if err != nil {
			return err
		}
		table := tableprinter.New(os.Stdout, terminal.IsTerminalOutput(), width)

		for _, extension := range api.Sunbeam.Extensions {
			table.AddField(extension.Name)
			table.AddField(extension.Title)
			table.EndRow()
		}
		table.Render()
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
			}

			if _, ok := api.Sunbeam.Extensions[repo.Name]; ok {
				item.Accessories = []string{"Installed"}
				item.Actions = []tui.Action{
					{
						Title: "Uninstall",
					},
				}
			} else {
				item.Actions = []tui.Action{
					{
						Title: "Install",
					},
				}
			}

			extensionItems[i] = item
		}

		list := tui.NewList("Browse Extensions")
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
		for _, extension := range api.Sunbeam.Extensions {
			extensionItems = append(extensionItems, tui.ListItem{
				Title:    extension.Title,
				Subtitle: extension.Name,
			})
		}

		list := tui.NewList("Manage Installed Extensions")
		list.SetItems(extensionItems)

		return tui.Draw(list, globalOptions)
	},
}

func runExtensionInstall(cmd *cobra.Command, args []string) (err error) {
	if _, err = os.Stat(api.Sunbeam.ExtensionRoot); os.IsNotExist(err) {
		os.MkdirAll(api.Sunbeam.ExtensionRoot, 0755)
	}

	if args[0] == "." {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		name := filepath.Base(wd)
		targetLink := filepath.Join(api.Sunbeam.ExtensionRoot, name)
		if err := os.MkdirAll(filepath.Dir(targetLink), 0755); err != nil {
			return err
		}

		return os.Symlink(wd, targetLink)
	}

	// Remote repository
	url, err := url.Parse(args[0])
	if err != nil {
		log.Fatalln(err)
	}

	name := path.Base(url.Path)
	target := path.Join(api.Sunbeam.ExtensionRoot, name)
	if _, err = os.Stat(target); err == nil {
		log.Fatalf("Extension %s already installed", name)
	}

	command := exec.Command("git", "clone", url.String())
	command.Dir = api.Sunbeam.ExtensionRoot
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Installed extension", name)
	return nil
}
