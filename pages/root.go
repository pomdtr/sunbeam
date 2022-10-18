package pages

import (
	"log"
	"os/exec"
	"path"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/scripts"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/skratchdot/open-golang/open"
)

type Page struct {
	scripts.Command
	input     scripts.CommandInput
	container Container
}

type model struct {
	width, height int
	rootCommand   scripts.Command

	pages []Page
}

func NewRoot(command scripts.Command) model {
	return model{rootCommand: command}
}

func (m *model) PushPage(command scripts.Command, input scripts.CommandInput) tea.Cmd {
	loading := NewLoadingContainer(command.Title())
	loading.SetSize(m.width, m.height)
	m.pages = append(m.pages, Page{Command: command, input: input, container: loading})
	return tea.Batch(loading.Init(), m.Run)
}

func (m *model) PopPage() {
	if len(m.pages) == 0 {
		return
	}
	m.pages = m.pages[:len(m.pages)-1]
}

func (m *model) CurrentPage() *Page {
	if len(m.pages) == 0 {
		return nil
	}
	return &m.pages[len(m.pages)-1]
}

func (m model) Run() tea.Msg {
	response, err := m.CurrentPage().Run(scripts.CommandInput{})
	if err != nil {
		return err
	}
	return response
}

func (m *model) Init() tea.Cmd {
	return m.PushPage(m.rootCommand, scripts.CommandInput{})
}

func (m *model) SetSize(width, height int) {
	m.width = width
	m.height = height
	for i := range m.pages {
		if m.pages[i].container != nil {
			m.pages[i].container.SetSize(width, height)
		}
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case PopMsg:
		log.Println(len(m.pages))
		if len(m.pages) == 1 {
			return m, tea.Quit
		}
		m.PopPage()
		return m, nil
	case scripts.ScriptResponse:
		switch msg.Type {
		case "list":
			list := msg.List
			if list.Title == "" {
				list.Title = m.CurrentPage().Title()
			}

			listContainer := NewListContainer(list)
			listContainer.SetSize(m.width, m.height)
			m.CurrentPage().container = listContainer
			return m, listContainer.Init()
		case "detail":
			detail := msg.Detail
			if detail.Title == "" {
				detail.Title = m.CurrentPage().Title()
			}

			detailContainer := NewDetailContainer(detail)
			m.CurrentPage().container = detailContainer
			return m, detailContainer.Init()
		case "form":
			form := msg.Form
			if form.Title == "" {
				form.Title = m.CurrentPage().Title()
			}
			// submitAction := func(values map[string]string) tea.Cmd {
			// 	switch form.Method {
			// 	case "args":
			// 		return RunCmd(c.command, c.input)
			// 	case "env":
			// 		return RunCmd(c.command, c.input)
			// 	}
			// 	return utils.NewErrorCmd("unknown form method: %s", msg.Form.Method)
			// }
		}
	case scripts.ScriptAction:
		cmd := m.RunAction(msg)
		return m, cmd

	case error:
		log.Printf("Error: %v", msg)
		return m, tea.Quit
	}

	var cmd tea.Cmd
	container, cmd := m.CurrentPage().container.Update(msg)
	m.CurrentPage().container = container

	return m, cmd
}

func (m model) Command() scripts.Command {
	return m.CurrentPage().Command
}

func (m *model) View() string {
	return m.CurrentPage().container.View()
}

func (m *model) RunAction(action scripts.ScriptAction) tea.Cmd {
	switch action.Type {
	case "push":
		commandUrl := m.Command().Url()
		commandDir := path.Dir(commandUrl.Path)
		command, err := scripts.Parse(path.Join(commandDir, action.Path))
		if err != nil {
			log.Fatal(err)
		}

		return m.PushPage(command, scripts.CommandInput{Arguments: action.Args})
	case "exec":
		var cmd *exec.Cmd
		if len(action.Command) == 1 {
			cmd = exec.Command(action.Command[0])
		} else {
			cmd = exec.Command(action.Command[0], action.Command[1:]...)
		}
		err := cmd.Run()
		if err != nil {
			return utils.SendMsg(
				scripts.ScriptResponse{
					Type: "detail",
					Detail: &scripts.DetailResponse{
						Format: "text",
						Text:   err.Error(),
					},
				},
			)
		}
		return tea.Quit
	case "open":
		err := open.Run(action.Path)
		if err != nil {
			return utils.NewErrorCmd("failed to open file: %s", err)
		}
		return tea.Quit
	case "open-url":
		err := open.Run(action.Url)
		if err != nil {
			return utils.NewErrorCmd("failed to open url: %s", err)
		}
		return tea.Quit
	case "copy":
		err := clipboard.WriteAll(action.Content)
		if err != nil {
			return utils.NewErrorCmd("failed to access clipboard: %s", err)
		}
		return tea.Quit
	default:
		log.Printf("Unknown action type: %s", action.Type)
		return tea.Quit
	}

}
