package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/scripts"
)

type FormItem struct {
	Field scripts.Field
	input FormInput
}

type Form struct {
	inputs    []FormInput
	submitCmd func() tea.Msg

	width    int
	header   Header
	footer   Footer
	viewport viewport.Model

	scrollOffset int
	focusIndex   int
}

func hasMissingFields(fields []scripts.Field) bool {
	for _, field := range fields {
		if field.Value == "" {
			return true
		}
	}

	return false
}

func NewForm(fields []scripts.Field, submitFunc func(fields []scripts.Field) tea.Msg) *Form {
	header := NewHeader()
	viewport := viewport.New(0, 0)
	footer := NewFooter("Sunbeam")
	footer.SetBindings(
		key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("⌃S", "Submit")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Focus Next")),
	)

	items := make([]FormItem, 0)
	inputs := make([]FormInput, 0)
	for _, field := range fields {
		item := FormItem{
			Field: field,
		}

		if field.Input != nil {
			item.input = NewFormInput(field.Input)
			inputs = append(inputs, item.input)
		}
		items = append(items, item)
	}

	return &Form{
		header: header,
		submitCmd: func() tea.Msg {
			fields := make([]scripts.Field, 0)
			for _, item := range items {
				if item.input != nil {
					item.Field.Value = item.input.Value()
				}

				fields = append(fields, item.Field)
			}

			return submitFunc(fields)
		},
		footer:   footer,
		viewport: viewport,
		inputs:   inputs,
	}
}

func (c *Form) SetIsLoading(isLoading bool) tea.Cmd {
	return c.header.SetIsLoading(isLoading)
}

func (c Form) Init() tea.Cmd {
	if len(c.inputs) == 0 {
		return func() tea.Msg { return fmt.Errorf("form has no items") }
	}
	return c.inputs[0].Focus()
}

func (c *Form) CurrentItem() FormInput {
	if c.focusIndex >= len(c.inputs) {
		return nil
	}
	return c.inputs[c.focusIndex]
}

func (c *Form) ScrollViewport() {
	cursorOffset := 0
	for i := 0; i < c.focusIndex; i++ {
		cursorOffset += c.inputs[i].Height() + 2
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

func (c Form) Update(msg tea.Msg) (*Form, tea.Cmd) {
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
			if c.focusIndex == len(c.inputs) {
				c.focusIndex = 0
			} else if c.focusIndex < 0 {
				c.focusIndex = len(c.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(c.inputs))
			for i := 0; i <= len(c.inputs)-1; i++ {
				if i == c.focusIndex {
					// Set focused state
					cmds[i] = c.inputs[i].Focus()
					continue
				}
				// Remove focused state
				c.inputs[i].Blur()
			}

			c.renderInputs()
			if c.viewport.Height > 0 {
				c.ScrollViewport()
			}

			return &c, tea.Batch(cmds...)
		case tea.KeyCtrlS:
			return &c, c.submitCmd
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
	itemViews := make([]string, len(c.inputs))
	maxWidth := 0
	for i, input := range c.inputs {
		var inputView = lipgloss.NewStyle().Padding(0, 1).Render(input.View())
		if i == c.focusIndex {
			inputView = selectedBorder.Render(inputView)
		} else {
			inputView = normalBorder.Render(inputView)
		}

		itemViews[i] = lipgloss.JoinHorizontal(lipgloss.Center, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%s: ", input.Title())), inputView)
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
	cmds := make([]tea.Cmd, len(c.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range c.inputs {
		c.inputs[i], cmds[i] = c.inputs[i].Update(msg)
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
	for _, input := range c.inputs {
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
