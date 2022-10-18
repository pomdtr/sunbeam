package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/alexflint/go-arg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/pages"
	"github.com/pomdtr/sunbeam/scripts"
)

type ServeCmd struct {
	Host string `arg:"-h,--host" help:"Host to serve on"`
	Port int    `arg:"-p,--port" help:"Port to serve on" default:"8080"`
}

type RunCmd struct {
	ScriptPath string `arg:"positional" help:"Path to script to run"`
}

var args struct {
	Serve       *ServeCmd `arg:"subcommand:serve" help:"Serve all scripts"`
	Run         *RunCmd   `arg:"subcommand:run" help:"Run a script"`
	CommandRoot string    `arg:"-c,--command-root" help:"Directory to load commands from"`
}

// Todo: move this to the cmd package, and switch to the cobra library
func main() {
	var err error
	arg.MustParse(&args)

	if args.Serve != nil {
		// err = server.Serve(args.Serve.Host, args.Serve.Port)
		// if err != nil {
		// 	log.Fatalln(err)
		// }
		return
	}

	var rootCmd scripts.Command
	if args.Run != nil {
		rootCmd, err = scripts.Parse(args.Run.ScriptPath)
		if err != nil {
			log.Fatalf("Error parsing script: %v", err)
		}

	} else {
		commandRoot := args.CommandRoot
		if commandRoot == "" {
			commandRoot = scripts.CommandDir
		}

		rootCmd = scripts.RootCommand{Root: commandRoot}
	}

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

	m := pages.NewRoot(rootCmd)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
