package tui

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/utils"
)

var debounceDuration = 300 * time.Millisecond

type ListItem struct {
	Id          string
	Title       string
	Subtitle    string
	Preview     string
	PreviewCmd  func() string
	Accessories []string
	Actions     []Action
}

func ParseScriptItem(scriptItem app.ScriptItem) ListItem {
	actions := make([]Action, len(scriptItem.Actions))
	for i, scriptAction := range scriptItem.Actions {
		if i == 0 {
			scriptAction.Shortcut = "enter"
		}
		actions[i] = NewAction(scriptAction)
	}

	return ListItem{
		Id:       scriptItem.Id,
		Title:    scriptItem.Title,
		Subtitle: scriptItem.Subtitle,
		Preview:  scriptItem.Preview,
		PreviewCmd: func() string {
			cmd := scriptItem.PreviewCommand()
			if cmd == nil {
				return ""
			}
			out, err := cmd.Output()
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					return string(exitErr.Stderr)
				}
				return err.Error()
			}

			return string(out)
		},
		Accessories: scriptItem.Accessories,
		Actions:     actions,
	}

}

func (i ListItem) ID() string {
	return i.Id
}

func (i ListItem) FilterValue() string {
	if i.Subtitle == "" {
		return i.Title
	}
	return fmt.Sprintf("%s %s", i.Title, i.Subtitle)
}

func (i ListItem) Render(width int, selected bool) string {
	if width == 0 {
		return ""
	}

	var title string
	var titleColor lipgloss.TerminalColor
	if selected {
		title = fmt.Sprintf("> %s", i.Title)
		titleColor = accentColor
	} else {
		title = fmt.Sprintf("  %s", i.Title)
		titleColor = theme.Fg()
	}

	subtitle := fmt.Sprintf(" %s", i.Subtitle)
	var blanks string
	accessories := fmt.Sprintf(" %s", strings.Join(i.Accessories, " • "))

	if width >= lipgloss.Width(title+subtitle+accessories) {
		availableWidth := width - lipgloss.Width(title+subtitle+accessories)
		blanks = strings.Repeat(" ", availableWidth)
	} else if width >= lipgloss.Width(title+accessories) {
		subtitle = subtitle[:width-lipgloss.Width(title+accessories)]
	} else if width >= lipgloss.Width(accessories) {
		subtitle = ""
		title = title[:width-lipgloss.Width(accessories)]
	} else {
		accessories = ""
		title = title[:width]
	}

	title = styles.Bold.Copy().Foreground(titleColor).Render(title)
	subtitle = styles.Faint.Render(subtitle)
	blanks = styles.Regular.Render(blanks)
	accessories = styles.Faint.Render(accessories)

	return lipgloss.JoinHorizontal(lipgloss.Top, title, subtitle, blanks, accessories)
}

type List struct {
	header  Header
	footer  Footer
	actions ActionList

	Dynamic     bool
	ShowPreview bool

	previewContent string
	filter         Filter
	viewport       viewport.Model
}

func NewList(title string) *List {
	actions := NewActionList()

	header := NewHeader()

	viewport := viewport.New(0, 0)
	viewport.Style = styles.Regular

	filter := NewFilter()
	filter.DrawLines = true

	footer := NewFooter(title)

	return &List{
		actions:  actions,
		header:   header,
		filter:   filter,
		viewport: viewport,
		footer:   footer,
	}
}

func (c *List) Init() tea.Cmd {
	if len(c.filter.choices) > 0 {
		return tea.Batch(c.FilterItems(c.Query()), c.header.Focus())
	}
	return c.header.Focus()
}

func (c *List) SetSize(width, height int) {
	availableHeight := utils.Max(0, height-lipgloss.Height(c.header.View())-lipgloss.Height(c.footer.View()))
	c.footer.Width = width
	c.header.Width = width
	c.actions.SetSize(width, height)
	if c.ShowPreview {
		listWidth := width / 3
		c.filter.SetSize(listWidth, availableHeight)
		c.viewport.Width = c.viewport.Width - listWidth
		c.viewport.Height = availableHeight
		c.setPreviewContent(c.previewContent)
	} else {
		c.filter.SetSize(width, availableHeight)
	}
}

func (l *List) setPreviewContent(content string) {
	l.previewContent = content
	content = wordwrap.String(content, l.viewport.Width-4)
	content = lipgloss.NewStyle().Padding(0, 1).Width(l.viewport.Width).Render(content)
	l.viewport.SetContent(content)
}

func (c *List) SetItems(items []ListItem) tea.Cmd {
	filterItems := make([]FilterItem, len(items))
	for i, item := range items {
		filterItems[i] = item
	}

	c.filter.SetItems(filterItems)
	return c.FilterItems(c.Query())
}

func (c *List) SetIsLoading(isLoading bool) tea.Cmd {
	return c.header.SetIsLoading(isLoading)
}

type PreviewUpdateMsg struct {
	selectedItem ListItem
}

type PreviewContentMsg string

func (l *List) updateActions(item ListItem) tea.Cmd {
	l.actions.SetTitle(item.Title)
	l.actions.SetActions(item.Actions...)

	if len(item.Actions) == 0 {
		l.footer.SetBindings()
	} else if len(item.Actions) == 1 {
		l.footer.SetBindings(
			key.NewBinding(key.WithKeys(item.Actions[0].Shortcut), key.WithHelp("↩", item.Actions[0].Title)),
		)
	} else {
		l.footer.SetBindings(
			key.NewBinding(key.WithKeys(item.Actions[0].Shortcut), key.WithHelp("↩", item.Actions[0].Title)),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Actions")),
		)
	}

	var cmd tea.Cmd
	if l.ShowPreview {
		if item.Preview != "" {
			l.setPreviewContent(item.Preview)
		} else if item.PreviewCmd != nil {
			if l.previewContent == "" {
				cmd = func() tea.Msg {
					return PreviewUpdateMsg{selectedItem: item}
				}
			} else {
				cmd = tea.Tick(debounceDuration, func(_ time.Time) tea.Msg {
					return PreviewUpdateMsg{selectedItem: item}
				})
			}
		}
	}
	return cmd
}

func (c *List) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			if c.actions.Focused() {
				break
			} else if c.header.input.Value() != "" {
				c.header.input.SetValue("")
				cmd := c.FilterItems("")
				return c, cmd
			} else {
				return c, PopCmd
			}
		case tea.KeyShiftDown:
			c.viewport.LineDown(1)
			return c, nil
		case tea.KeyShiftUp:
			c.viewport.LineUp(1)
			return c, nil
		}
	case updateQueryMsg:
		if msg.query != c.Query() {
			return c, nil
		}

		return c, NewReloadPageCmd(app.ScriptInputs{
			"query": app.ScriptInput{
				Value: msg.query,
			},
		})
	case PreviewUpdateMsg:
		if c.filter.Selection() == nil {
			return c, nil
		}
		if msg.selectedItem.ID() != c.filter.Selection().ID() {
			return c, nil
		}
		cmd := c.header.SetIsLoading(true)
		return c, tea.Batch(cmd, func() tea.Msg {
			return PreviewContentMsg(msg.selectedItem.PreviewCmd())
		})
	case PreviewContentMsg:
		c.header.SetIsLoading(false)
		c.setPreviewContent(string(msg))
		return c, nil
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	c.actions, cmd = c.actions.Update(msg)
	cmds = append(cmds, cmd)

	if c.actions.Focused() {
		return c, tea.Batch(cmds...)
	}

	header, cmd := c.header.Update(msg)
	cmds = append(cmds, cmd)
	if header.Value() != c.header.Value() {
		if c.Dynamic {
			cmd = tea.Tick(debounceDuration, func(_ time.Time) tea.Msg {
				return updateQueryMsg{query: header.Value()}
			})
			cmds = append(cmds, cmd)
		} else {
			cmd = c.FilterItems(header.Value())
			cmds = append(cmds, cmd)
		}
	}
	c.header = header

	filter, cmd := c.filter.Update(msg)
	cmds = append(cmds, cmd)
	if filter.Selection() == nil {
		c.actions.SetTitle("")
		c.actions.SetActions()
		c.footer.SetBindings()
		c.setPreviewContent("")
	} else if c.filter.Selection() == nil || c.filter.Selection().ID() != filter.Selection().ID() {
		selection := filter.Selection().(ListItem)
		cmd = c.updateActions(selection)
		cmds = append(cmds, cmd)
	}
	c.filter = filter

	return c, tea.Batch(cmds...)
}

type updateQueryMsg struct {
	query string
}

func (c *List) FilterItems(query string) tea.Cmd {
	c.filter.FilterItems(query)
	if c.filter.Selection() != nil {
		return c.updateActions(c.filter.Selection().(ListItem))
	}
	return nil
}

func (c List) View() string {
	if c.actions.Focused() {
		return c.actions.View()
	}

	if c.ShowPreview {
		var separatorChars = make([]string, c.viewport.Height)
		for i := 0; i < c.viewport.Height; i++ {
			separatorChars[i] = "│"
		}
		separator := styles.Regular.Render(strings.Join(separatorChars, "\n"))
		view := lipgloss.JoinHorizontal(lipgloss.Top, c.filter.View(), separator, c.viewport.View())

		return lipgloss.JoinVertical(lipgloss.Top, c.header.View(), view, c.footer.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.filter.View(), c.footer.View())
}

func (c List) Query() string {
	return c.header.input.Value()
}

func NewErrorCmd(err error) func() tea.Msg {
	return func() tea.Msg {
		return err
	}
}
