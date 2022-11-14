package cmd

import (
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var installExtensionCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a sunbeam extension",
	Args:  cobra.ExactArgs(1),
	Run:   InstallExtension,
}

func init() {
	rootCmd.AddCommand(installExtensionCmd)
}

func InstallExtension(cmd *cobra.Command, args []string) {
	_, err := git.PlainClone("/tmp", false, &git.CloneOptions{
		URL:      args[0],
		Progress: os.Stdout,
	})
	if err != nil {
		log.Fatalln(err)
	}

}

// func ExtensionInstall() *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "install",
// 		Short: "Install an extension",
// 		Run:   ExtensionInstall,
// 	}
// }
