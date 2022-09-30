package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexflint/go-arg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/commands"
	"github.com/pomdtr/sunbeam/pages"
	"github.com/pomdtr/sunbeam/server"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

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
		case tea.KeyEscape:
			if len(m.pages) == 1 {
				return m, tea.Quit
			}
			m.PopPage()
			return m, nil
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

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
	Port int    `arg:"-p,--port" help:"Port to serve on"`
}

type RunCmd struct {
	ScriptPath string `arg:"positional" help:"Path to script to run"`
}

var args struct {
	Serve      *ServeCmd `arg:"subcommand:serve" help:"Serve all scripts"`
	Run        *RunCmd   `arg:"subcommand:run" help:"Run a script"`
	CommandDir string    `arg:"-c,--command-dir" help:"Directory to load commands from"`
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
		script, err := commands.Parse(args.Run.ScriptPath)
		if err != nil {
			log.Fatalf("Error parsing script: %v", err)
		}
		root = pages.NewCommandContainer(commands.Command{
			Script: script,
		})
		if err != nil {
			log.Fatalln(err)
		}

		return
	} else {
		if args.CommandDir != "" {
			commands.CommandDir = args.CommandDir
		}
		root = pages.NewRootContainer(commands.CommandDir)
	}

	// Log to a file
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}
	defer f.Close()

	m := NewModel(root)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
