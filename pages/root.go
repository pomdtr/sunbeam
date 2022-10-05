package pages

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
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
	commandRoot url.URL
	width       int
	height      int
	textInput   textinput.Model
	*list.Model
}

func NewRootContainer(commandDir string) Page {
	d := NewItemDelegate()

	l := list.New([]list.Item{}, d, 0, 0)

	textInput := textinput.NewModel()
	textInput.Prompt = ""
	textInput.Placeholder = "Search for commands..."
	textInput.Focus()
	rootURL, err := url.Parse(commandDir)
	if err != nil {
		log.Fatal(err)
	}
	if rootURL.Scheme == "" {
		rootURL.Scheme = "file"
	}

	return &RootContainer{Model: &l, textInput: textInput, commandRoot: *rootURL}
}

func (container *RootContainer) Init() tea.Cmd {
	return container.gatherScripts
}

func (c RootContainer) gatherScripts() tea.Msg {
	scripts := make([]commands.Script, 0)
	if c.commandRoot.Scheme == "file" {
		if _, err := os.Stat(c.commandRoot.Path); os.IsNotExist(err) {
			log.Fatalf("Command directory %s does not exist", c.commandRoot.Path)
		}
		dirScripts, err := commands.ScanDir(c.commandRoot.Path)
		if err != nil {
			return err
		}
		scripts = append(scripts, dirScripts...)
	} else {
		res, err := http.Get(c.commandRoot.String())
		if err != nil {
			log.Fatalf("Could not fetch commands from %s", c.commandRoot.String())
		}
		var index map[string]commands.ScriptMetadatas
		err = json.NewDecoder(res.Body).Decode(&index)
		if err != nil {
			log.Fatalf("Could not parse commands from %s", c.commandRoot.String())
		}
		for route, metadatas := range index {
			scripts = append(scripts, commands.Script{
				Metadatas: metadatas,
				Url: url.URL{
					Scheme: c.commandRoot.Scheme,
					Host:   c.commandRoot.Host,
					Path:   route,
				},
			})
		}
	}

	return scripts
}

func (container RootContainer) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmds []tea.Cmd
	selectedItem := container.Model.SelectedItem()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return &container, PopCmd
		case tea.KeyCtrlE:
			if selectedItem == nil {
				break
			}
			selectedItem := selectedItem.(commands.Script)
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			c := exec.Command(editor, selectedItem.Url.Path)
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
			return &container, NewPushCmd(commands.Command{Script: selectedItem})
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
