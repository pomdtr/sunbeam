package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Form struct {
	id            string
	width, height int
	viewport      viewport.Model
	isLoading     bool
	spinner       spinner.Model

	scrollOffset int
	focusIndex   int

	items []FormItem
}

type SubmitMsg map[string]any

func NewForm(id string, items ...FormItem) *Form {
	viewport := viewport.New(0, 0)

	form := &Form{
		id:       id,
		viewport: viewport,
		items:    items,
	}

	return form
}

func (c *Form) SetIsLoading(isLoading bool) tea.Cmd {
	c.isLoading = isLoading
	if isLoading {
		return c.spinner.Tick
	}

	return nil
}

func (c Form) Init() tea.Cmd {
	if len(c.items) == 0 {
		return nil
	}
	return c.items[0].Focus()
}

func (c Form) Focus() tea.Cmd {
	if len(c.items) == 0 {
		return nil
	}

	return c.items[c.focusIndex].Focus()
}

func (c *Form) Blur() tea.Cmd {
	return nil
}

func (c *Form) CurrentItem() FormInput {
	if c.focusIndex >= len(c.items) {
		return nil
	}
	return c.items[c.focusIndex]
}

func (c *Form) ScrollViewport() {
	cursorOffset := 0
	for i := 0; i < c.focusIndex; i++ {
		cursorOffset += c.items[i].Height() + 2
	}

	if c.CurrentItem() == nil {
		return
	}
	maxRequiredVisibleHeight := cursorOffset + c.CurrentItem().Height() + 2
	for maxRequiredVisibleHeight > c.viewport.Height+c.scrollOffset {
		c.viewport.LineDown(1)
		c.scrollOffset += 1
	}

	for cursorOffset < c.scrollOffset {
		c.viewport.LineUp(1)
		c.scrollOffset -= 1
	}
}

func (c Form) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return &c, func() tea.Msg {
				return PopPageMsg{}
			}
		// Set focus to next input
		case "tab", "shift+tab":
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				c.focusIndex--
			} else {
				c.focusIndex++
			}

			// Cycle focus
			if c.focusIndex > len(c.items) {
				c.focusIndex = 0
			} else if c.focusIndex < 0 {
				c.focusIndex = len(c.items)
			}

			cmds := make([]tea.Cmd, len(c.items))
			for i := 0; i <= len(c.items)-1; i++ {
				if i == c.focusIndex {
					// Set focused state
					cmds[i] = c.items[i].Focus()
					continue
				}
				// Remove focused state
				c.items[i].Blur()
			}

			c.renderInputs()
			if c.viewport.Height > 0 {
				c.ScrollViewport()
			}

			return &c, tea.Batch(cmds...)
		case "enter", "ctrl+s":
			if msg.String() == "enter" && c.focusIndex != len(c.items) {
				break
			}

			return &c, func() tea.Msg {
				values := make(map[string]any)
				for _, input := range c.items {
					if input.Value() == "" && !input.Optional {
						return nil
					}
					values[input.Name] = input.Value()
				}

				return SubmitMsg(values)
			}
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	if cmd = c.updateInputs(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	c.renderInputs()

	return &c, tea.Batch(cmds...)
}

func (c *Form) renderInputs() {
	selectedBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("13"))
	normalBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true)
	itemViews := make([]string, len(c.items))
	maxWidth := 0
	for i, input := range c.items {
		var inputView = lipgloss.NewStyle().Padding(0, 1).Render(input.View())
		if i == c.focusIndex {
			inputView = selectedBorder.Render(inputView)
		} else {
			inputView = normalBorder.Render(inputView)
		}

		var titleView string
		if input.Optional {
			titleView = fmt.Sprintf("%s ", input.Title)
		} else {
			asterisk := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("*")
			titleView = fmt.Sprintf("%s%s ", input.Title, asterisk)
		}

		itemViews[i] = lipgloss.JoinHorizontal(lipgloss.Center, lipgloss.NewStyle().Bold(true).Render(titleView), inputView)
		if lipgloss.Width(itemViews[i]) > maxWidth {
			maxWidth = lipgloss.Width(itemViews[i])
		}
	}

	for i := range itemViews {
		itemViews[i] = lipgloss.NewStyle().Width(maxWidth).Align(lipgloss.Right).Render(itemViews[i])
	}

	formView := lipgloss.JoinVertical(lipgloss.Left, itemViews...)
	formView = lipgloss.NewStyle().Width(c.width).Align(lipgloss.Center).Render(formView)

	c.viewport.SetContent(formView)
}

func (c Form) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(c.items))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range c.items {
		c.items[i].FormInput, cmds[i] = c.items[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (c *Form) SetSize(width, height int) {
	c.width, c.height = width, height
	c.viewport.Height = max(0, height-2)
	for _, input := range c.items {
		input.SetWidth(width / 2)
	}

	c.renderInputs()
	if c.viewport.Height > 0 {
		c.ScrollViewport()
	}
}

func (c *Form) View() string {
	separator := strings.Repeat("─", c.width)

	var submitAction string
	if c.focusIndex == len(c.items) {
		submitAction = renderAction("Submit", "enter", true)
	} else {
		submitAction = renderAction("Submit", "ctrl+s", false)
	}

	submitRow := lipgloss.NewStyle().Align(lipgloss.Right).Padding(0, 1).Width(c.width).Render(fmt.Sprintf("%s · %s", submitAction, renderAction("Focus Next", "tab", false)))
	return lipgloss.JoinVertical(lipgloss.Left, c.viewport.View(), separator, submitRow)
}
