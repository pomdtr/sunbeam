package pages

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
	"github.com/pomdtr/sunbeam/bubbles/list"
	commands "github.com/pomdtr/sunbeam/commands"
)

type RootContainer struct {
	commandDirs []string
	width       int
	height      int
	textInput   textinput.Model
	*list.Model
}

func NewRootContainer(commandDirs []string) Page {
	d := NewItemDelegate()

	l := list.New([]list.Item{}, d, 0, 0)

	textInput := textinput.NewModel()
	textInput.Prompt = ""
	textInput.Placeholder = "Search for apps and commands..."
	textInput.Focus()

	return &RootContainer{Model: &l, textInput: textInput, commandDirs: commandDirs}
}

func (container *RootContainer) Init() tea.Cmd {
	return container.gatherScripts
}

func (container RootContainer) gatherScripts() tea.Msg {
	scripts := make([]commands.Script, 0)
	for _, commandDir := range container.commandDirs {
		if _, err := os.Stat(commandDir); os.IsNotExist(err) {
			log.Printf("Command directory %s does not exist", commandDir)
			continue
		}
		dirScripts, err := commands.ScanDir(commandDir)
		if err != nil {
			return err
		}
		scripts = append(scripts, dirScripts...)
	}

	return scripts
}

func (container RootContainer) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmds []tea.Cmd
	selectedItem := container.Model.SelectedItem()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlT:
			c := exec.Command("bash", "-c", "clear && fish")
			cmd := tea.ExecProcess(c, nil)
			return &container, cmd
		case tea.KeyCtrlH:
			c := exec.Command("htop")
			cmd := tea.ExecProcess(c, nil)
			return &container, cmd
		case tea.KeyCtrlE:
			if selectedItem == nil {
				break
			}
			selectedItem := selectedItem.(commands.Script)
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			c := exec.Command(editor, selectedItem.Path)
			cmd := tea.ExecProcess(c, func(err error) tea.Msg {
				if err != nil {
					return err
				}
				return container.gatherScripts()
			})
			return &container, cmd
		case tea.KeyEnter:
			if selectedItem == nil {
				break
			}
			selectedItem, ok := container.Model.SelectedItem().(commands.Script)
			if !ok {
				return &container, tea.Quit
			}
			var next = NewCommandContainer(commands.NewCommand(selectedItem))
			return &container, NewPushCmd(next)
		}
	case []commands.Script:
		items := make([]list.Item, len(msg))
		for i, script := range msg {
			items[i] = script
		}

		cmd := container.SetItems(items)
		return &container, cmd
	}

	textinput, cmd := container.textInput.Update(msg)
	cmds = append(cmds, cmd)
	container.textInput = textinput

	model, cmd := container.Model.Update(msg)
	cmds = append(cmds, cmd)
	container.Model = &model

	return &container, tea.Batch(cmds...)
}

func (c RootContainer) headerView() string {
	input := c.textInput.View()
	line := strings.Repeat("─", c.width)
	return lipgloss.JoinVertical(lipgloss.Left, input, line)
}

func (container *RootContainer) footerView() string {
	return bubbles.SunbeamFooterWithActions(container.width, "Sunbeam", "Open Command")
}

func (container *RootContainer) SetSize(width, height int) {
	container.width, container.height = width, height
	container.Model.SetSize(width, height-lipgloss.Height(container.footerView())-lipgloss.Height(container.headerView()))
}

func (container *RootContainer) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, container.headerView(), container.Model.View(), container.footerView())
}
