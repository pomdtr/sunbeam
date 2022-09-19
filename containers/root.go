package containers

import (
	"log"
	"os/exec"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	commands "github.com/pomdtr/sunbeam/commands"
)

type RootContainer struct {
	commandDirs []string
	list.Model
}

func NewRootContainer(commandDirs []string) RootContainer {
	d := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Commands"

	return RootContainer{Model: l, commandDirs: commandDirs}
}

func (container *RootContainer) Init() tea.Cmd {
	return container.gatherScripts
}

func (container *RootContainer) gatherScripts() tea.Msg {
	log.Println("Gathering scripts")
	scripts := make([]commands.Script, 0)
	for _, commandDir := range container.commandDirs {
		dirScripts, err := commands.ScanDir(commandDir)
		if err != nil {
			return err
		}
		scripts = append(scripts, dirScripts...)
	}

	return scripts
}

func (container *RootContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	var cmd tea.Cmd
	selectedItem := container.Model.SelectedItem()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "e":
				if selectedItem == nil {
					break
				}
				selectedItem := selectedItem.(commands.Script)
				c := exec.Command("vim", selectedItem.Path)
				cmd := tea.ExecProcess(c, func(err error) tea.Msg {
					if err != nil {
						return err
					}
					return container.gatherScripts()
				})
				return container, cmd
			}
		case tea.KeyEnter:
			selectedItem, ok := container.Model.SelectedItem().(commands.Script)
			if !ok {
				return container, tea.Quit
			}
			var next = NewScriptContainer(commands.NewCommand(selectedItem))
			next.SetSize(container.Model.Width(), container.Model.Height())
			return container, NewPushCmd(next)
		}
	case []commands.Script:
		items := make([]list.Item, len(msg))
		for i, script := range msg {
			items[i] = script
		}

		cmd = container.SetItems(items)
		return container, cmd
	}

	container.Model, cmd = container.Model.Update(msg)

	return container, cmd
}
