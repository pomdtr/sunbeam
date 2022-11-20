package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"

	"github.com/pomdtr/sunbeam/api"
	"github.com/spf13/cobra"
)

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

var extensionListCommand = &cobra.Command{
	Use:   "list",
	Short: "List installed extensions",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, extension := range api.Sunbeam.Extensions {
			fmt.Println(extension.Name)
		}
		return nil
	},
}

func init() {
	extensionCommand.AddCommand(extensionInstallCommand)
	extensionCommand.AddCommand(extensionListCommand)
	rootCmd.AddCommand(extensionCommand)
}

func runExtensionInstall(cmd *cobra.Command, args []string) (err error) {
	if _, err = os.Stat(api.Sunbeam.ExtensionRoot); os.IsNotExist(err) {
		os.MkdirAll(api.Sunbeam.ExtensionRoot, 0755)
	}

	if _, err = os.Stat(args[0]); err == nil {
		// Local path
		extensionPath := args[0]
		if !path.IsAbs(args[0]) {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			extensionPath = path.Join(wd, args[0])
		}
		name := path.Base(extensionPath)
		err = os.Symlink(args[0], path.Join(api.Sunbeam.ExtensionRoot, name))
		if err != nil {
			return err
		}
		fmt.Println("Installed extension", name)
		os.Exit(0)
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
