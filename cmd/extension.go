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

var extensionCmd = &cobra.Command{
	Use: "extension",
}

var extensionInstallCmd = &cobra.Command{
	Use:   "install <repository>",
	Short: "Install a sunbeam extension from a git repository",
	Args:  cobra.ExactArgs(1),
	Run:   InstallExtension,
}

func init() {
	extensionCmd.AddCommand(extensionInstallCmd)
	rootCmd.AddCommand(extensionCmd)
}

func InstallExtension(cmd *cobra.Command, args []string) {
	var err error
	if _, err = os.Stat(api.Sunbeam.ExtensionRoot); os.IsNotExist(err) {
		os.MkdirAll(api.Sunbeam.ExtensionRoot, 0755)
	}

	if _, err = os.Stat(args[0]); err == nil {
		// Local path
		extensionPath := args[0]
		if !path.IsAbs(args[0]) {
			wd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			extensionPath = path.Join(wd, args[0])
		}
		name := path.Base(extensionPath)
		err = os.Symlink(args[0], path.Join(api.Sunbeam.ExtensionRoot, name))
		if err != nil {
			log.Fatalln(err)
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
	if _, err = os.Stat(target); os.IsNotExist(err) {
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
}
