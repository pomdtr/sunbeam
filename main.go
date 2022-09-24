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
	"github.com/pomdtr/sunbeam/pages"
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
		m.pages[i].SetSize(width-4, height-2)
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
		container.SetSize(m.width-4, m.height-2)
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
	return lipgloss.NewStyle().Padding(1, 2).Render(m.pages[len(m.pages)-1].View())

}

type Flags struct {
	Stdin      bool
	CommandDir string
}

func parseArgs() ([]string, Flags) {
	var f Flags
	flag.BoolVar(&f.Stdin, "stdin", false, "Read input from stdin")
	flag.StringVar(&f.CommandDir, "command-dir", "", "Directory to load plugins from")
	flag.Parse()
	return flag.Args(), f
}

func main() {
	var err error
	args, flags := parseArgs()

	var root pages.Page
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
			actionRunner = pages.NewActionRunner(command)
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
			root = pages.NewListContainer("Sunbeam", res.List, actionRunner)
		case "detail":
			root = pages.NewDetailContainer(res.Detail, actionRunner)
		}
	} else {
		if len(args) > 0 {
			script, err := commands.Parse(args[0])
			if err != nil {
				log.Fatalf("Error parsing script: %v", err)
			}
			root = pages.NewCommandContainer(commands.NewCommand(script))
		} else {
			commandDirs := commands.CommandDirs
			if flags.CommandDir != "" {
				commandDirs = append(commandDirs, flags.CommandDir)
			}
			root = pages.NewRootContainer(commandDirs)
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
