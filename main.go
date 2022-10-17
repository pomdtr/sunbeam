package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/alexflint/go-arg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pages"
	"github.com/pomdtr/sunbeam/scripts"
	"github.com/pomdtr/sunbeam/server"
)

type navigator struct {
	width  int
	height int
	pages  []pages.Page
}

func NewModel(pages ...pages.Page) navigator {
	return navigator{
		pages: pages,
	}
}

func (m *navigator) PushPage(page pages.Page) {
	m.pages = append(m.pages, page)
}

func (m *navigator) PopPage() {
	if len(m.pages) == 0 {
		return
	}
	m.pages = m.pages[:len(m.pages)-1]
}

func (m navigator) Init() tea.Cmd {
	if len(m.pages) == 0 {
		return nil
	}
	return m.pages[0].Init()
}

func (m *navigator) SetSize(width, height int) {
	m.width = width
	m.height = height
	for i := range m.pages {
		m.pages[i].SetSize(width, height)
	}
}

func (m navigator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case pages.PopMsg:
		if len(m.pages) == 1 {
			return m, tea.Quit
		}
		m.PopPage()
		return m, nil

	case pages.PushMsg:
		container := msg.Container
		container.SetSize(m.width, m.height)
		m.PushPage(container)
		return m, msg.Container.Init()
	case error:
		log.Printf("Error: %v", msg)
		return m, tea.Quit
	}

	if len(m.pages) == 0 {
		return m, nil
	}

	var cmd tea.Cmd
	currentPageIdx := len(m.pages) - 1
	m.pages[currentPageIdx], cmd = m.pages[currentPageIdx].Update(msg)

	return m, cmd
}

func (m navigator) View() string {
	if len(m.pages) == 0 {
		return "No pages, something went wrong"
	}
	return lipgloss.NewStyle().Render(m.pages[len(m.pages)-1].View())
}

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

func main() {
	var err error
	arg.MustParse(&args)

	if args.Serve != nil {
		err = server.Serve(args.Serve.Host, args.Serve.Port)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}
	var root pages.Page

	if args.Run != nil {
		script, err := scripts.Parse(args.Run.ScriptPath)
		if err != nil {
			log.Fatalf("Error parsing script: %v", err)
		}
		root = pages.NewCommandContainer(scripts.Command{
			Script: script,
		})
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		if args.CommandRoot == "" {
			args.CommandRoot = scripts.CommandDir
		}
		root = pages.NewRootContainer(args.CommandRoot)
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

	m := NewModel(root)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
