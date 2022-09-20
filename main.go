package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jwalton/go-supportscolor"
	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/commands"
	"github.com/pomdtr/sunbeam/containers"
)

func setup() {
	term := supportscolor.Stderr()
	if term.Has16m {
		lipgloss.SetColorProfile(termenv.TrueColor)
	} else if term.Has256 {
		lipgloss.SetColorProfile(termenv.ANSI256)
	} else {
		lipgloss.SetColorProfile(termenv.ANSI)
	}

}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	width  int
	height int
	pages  []containers.Container
}

func NewModel(pages ...containers.Container) model {
	return model{
		pages: pages,
	}
}

func (m *model) PushPage(page containers.Container) {
	m.pages = append(m.pages, page)
}

func (m *model) PopPage() {
	if len(m.pages) == 0 {
		return
	}
	m.pages = m.pages[:len(m.pages)-1]
}

func (m model) Init() tea.Cmd {
	if len(m.pages) == 0 {
		return nil
	}
	return m.pages[0].Init()
}

func (m *model) SetSize(width, height int) {
	m.width = width
	m.height = height
	for i := range m.pages {
		m.pages[i].SetSize(width, height)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case containers.PushMsg:
		page := msg.Container
		page.SetSize(m.width, m.height)
		m.PushPage(msg.Container)
		cmds = append(cmds, msg.Container.Init())
	case containers.PopMsg:
		if len(m.pages) == 1 {
			return m, tea.Quit
		}
		m.PopPage()
		return m, nil

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
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)

}

func (m model) View() string {
	if len(m.pages) == 0 {
		return "No pages"
	}
	return m.pages[len(m.pages)-1].View()
}

type Flags struct {
	Stdin bool
}

func parseArgs() ([]string, Flags) {
	var f Flags
	flag.BoolVar(&f.Stdin, "stdin", false, "Read input from stdin")
	flag.Parse()
	return flag.Args(), f
}

func main() {
	var err error
	args, flags := parseArgs()

	var root containers.Container
	if flags.Stdin {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		res := commands.ScriptResponse{}
		err = json.Unmarshal(bytes, &res)
		if err != nil {
			log.Fatal(err)
		}

		err = commands.Validator.Struct(res)
		if err != nil {
			log.Fatal(err)
		}

		var actionRunner func(action commands.ScriptAction) tea.Cmd

		if len(args) > 0 {
			script, err := commands.Parse(args[1])
			if err != nil {
				log.Fatalf("Error parsing script: %v", err)
			}
			command := commands.NewCommand(script)
			actionRunner = containers.NewActionRunner(command)
		} else {
			actionRunner = func(action commands.ScriptAction) tea.Cmd {
				callback := func(params any) {
					bytes, err := json.Marshal(params)
					if err != nil {
						log.Fatalf("Error marshalling params: %v", err)
					}
					fmt.Println(string(bytes))
				}
				commands.RunAction(action, callback)

				return tea.Quit
			}
		}
		switch res.Type {
		case "list":
			root = containers.NewListContainer(res.List, actionRunner)
		case "detail":
			root = containers.NewDetailContainer(res.Detail, actionRunner)
		}
	} else {
		if len(args) > 0 {
			script, err := commands.Parse(args[0])
			if err != nil {
				log.Fatalf("Error parsing script: %v", err)
			}
			root = containers.NewCommandContainer(commands.NewCommand(script))

		} else {
			root = containers.NewRootContainer(commands.CommandDirs)
		}
	}

	// Log to a file
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}
	defer f.Close()

	m := NewModel(root)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
