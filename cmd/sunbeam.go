package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/pages"
	"github.com/pomdtr/sunbeam/scripts"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sunbeam",
	Short: "Command Line Launcher",
	Run:   Sunbeam,
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(serveCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func Sunbeam(cmd *cobra.Command, args []string) {
	var logFile string
	// Log to a file
	if env := os.Getenv("SUNBEAM_LOG_FILE"); env != "" {
		logFile = env
	} else {
		if _, err := os.Stat(xdg.StateHome); os.IsNotExist(err) {
			err = os.MkdirAll(xdg.StateHome, 0755)
			if err != nil {
				log.Fatalln(err)
			}
		}
		logFile = path.Join(xdg.StateHome, "sunbeam.log")
	}
	f, err := tea.LogToFile(logFile, "debug")
	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}
	defer f.Close()

	rootCmd := scripts.RootCommand{Root: scripts.CommandDir}
	m := pages.NewRoot(rootCmd)
	p := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
