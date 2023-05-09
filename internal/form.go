package internal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/types"
)

type Form struct {
	items     []*FormItem
	submitCmd func(values map[string]string) tea.Msg

	width    int
	header   Header
	footer   Footer
	viewport viewport.Model

	scrollOffset int
	focusIndex   int
}

func NewForm(title string, submitCmd func(values map[string]string) tea.Msg, inputs ...types.Input) *Form {
	header := NewHeader()
	viewport := viewport.New(0, 0)
	footer := NewFooter(title)
	footer.SetBindings(
		key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("⌃S", "Submit")),
	)

	var items []*FormItem
	for _, input := range inputs {
		item, err := NewFormItem(input)
		if err != nil {
			continue
		}

		items = append(items, item)
	}

	if len(items) > 0 {
		footer.bindings = append(footer.bindings, key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Next Input")))
	}

	return &Form{
		header:    header,
		submitCmd: submitCmd,
		footer:    footer,
		viewport:  viewport,
		items:     items,
	}
}

func (c *Form) SetIsLoading(isLoading bool) tea.Cmd {
	return c.header.SetIsLoading(isLoading)
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
		switch msg.Type {
		case tea.KeyEscape:
			return &c, func() tea.Msg {
				return PopPageMsg{}
			}
		// Set focus to next input
		case tea.KeyTab, tea.KeyShiftTab:
			s := msg.String()

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				c.focusIndex--
			} else {
				c.focusIndex++
			}

			// Cycle focus
			if c.focusIndex == len(c.items) {
				c.focusIndex = 0
			} else if c.focusIndex < 0 {
				c.focusIndex = len(c.items) - 1
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
		case tea.KeyCtrlS:
			values := make(map[string]string)
			for _, input := range c.items {
				if input.Value() == "" && !input.Optional {
					return &c, func() tea.Msg {
						return fmt.Errorf("required field %s is empty", input.Name)
					}
				}
				values[input.Name] = input.Value()
			}

			return &c, func() tea.Msg {
				return c.submitCmd(values)
			}
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	if cmd = c.updateInputs(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	c.renderInputs()

	if c.header, cmd = c.header.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

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
	formView = lipgloss.NewStyle().Width(c.footer.Width).Align(lipgloss.Center).Render(formView)

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

type SubmitFormMsg struct {
	Id     string
	Values map[string]any
}

func (c *Form) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width

	c.width = width
	for _, input := range c.items {
		input.SetWidth(width / 2)
	}
	c.viewport.Height = height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())

	c.renderInputs()
	if c.viewport.Height > 0 {
		c.ScrollViewport()
	}
}

func (c *Form) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
