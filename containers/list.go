package containers

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jinzhu/copier"
	commands "github.com/pomdtr/sunbeam/commands"
)

type ListContainer struct {
	list   *list.Model
	runner NewSelectActionCmd
}

var listContainer = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)

func NewListContainer(res commands.ListResponse, runner NewSelectActionCmd) ListContainer {
	var l list.Model
	copier.Copy(&l, &listContainer)
	l.SetShowTitle(false)

	listItems := make([]list.Item, len(res.Items))
	for i, item := range res.Items {
		listItems[i] = item
	}
	l.SetItems(listItems)

	return ListContainer{
		list:   &l,
		runner: runner,
	}
}

func (ListContainer) Init() tea.Cmd {
	return nil
}

func (c ListContainer) SetSize(width, height int) {
	c.list.SetSize(width, height)
}

func (c ListContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	var cmd tea.Cmd

	selectedItem := c.list.SelectedItem()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		// Handle key
		case tea.KeyRunes:
			if selectedItem == nil || c.list.FilterState() == list.Filtering {
				break
			}
			selectedItem := selectedItem.(commands.ScriptItem)
			for _, action := range selectedItem.Actions {
				if action.Keybind == string(msg.Runes) {
					return c, c.runner(action)
				}
			}
		case tea.KeyEscape:
			if c.list.FilterState() != list.Filtering {
				return c, PopCmd
			}
		case tea.KeyEnter:
			if selectedItem == nil {
				break
			}
			selectedItem := selectedItem.(commands.ScriptItem)
			if c.list.FilterState() != list.Filtering && len(selectedItem.Actions) > 0 {
				primaryAction := selectedItem.Actions[0]
				return c, c.runner(primaryAction)
			}
		}
	case commands.ScriptResponse:
		log.Printf("Pushing %d items", len(msg.List.Items))
		items := make([]list.Item, len(msg.List.Items))
		for i, item := range msg.List.Items {
			items[i] = item
		}
		cmd = c.list.SetItems(items)
	}

	l, cmd := c.list.Update(msg)
	c.list = &l

	return c, cmd
}

func (c ListContainer) View() string {
	return c.list.View()
}
