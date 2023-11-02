package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Form struct {
	width, height int
	viewport      viewport.Model
	isLoading     bool
	spinner       spinner.Model

	submitMsg func(map[string]any) tea.Msg

	scrollOffset int
	focusIndex   int

	inputs []Input
}

func FindMissingParams(params []types.Param, values map[string]any) []types.Param {
	var missing []types.Param
	for _, param := range params {
		if _, ok := values[param.Name]; !ok {
			if !param.Required {
				continue
			}

			missing = append(missing, param)
		}
	}

	return missing
}

func NewForm(submitMsg func(map[string]any) tea.Msg, params ...types.Param) *Form {
	viewport := viewport.New(0, 0)

	var inputs []Input
	for _, param := range params {
		switch param.Type {
		case types.ParamTypeString:
			inputs = append(inputs, NewTextInput(param))
		case types.ParamTypeBoolean:
			inputs = append(inputs, NewBooleanInput(param))
		case types.ParamTypeNumber:
			inputs = append(inputs, NewNumberInput(param))
		}
	}

	form := &Form{
		submitMsg: submitMsg,
		viewport:  viewport,
		inputs:    inputs,
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
	if len(c.inputs) == 0 {
		return nil
	}
	return c.inputs[0].Focus()
}

func (c Form) Focus() tea.Cmd {
	if len(c.inputs) == 0 {
		return nil
	}

	return c.inputs[c.focusIndex].Focus()
}

func (c *Form) Blur() tea.Cmd {
	return nil
}

func (c *Form) CurrentItem() Input {
	if c.focusIndex >= len(c.inputs) {
		return nil
	}
	return c.inputs[c.focusIndex]
}

func (f Form) itemsHeight() int {
	height := 0
	for _, item := range f.inputs {
		height += item.Height() + 2
	}
	return height
}

func (c *Form) ScrollViewport() {
	cursorOffset := 0
	for i := 0; i < c.focusIndex; i++ {
		cursorOffset += c.inputs[i].Height() + 2
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
			if c.focusIndex > len(c.inputs)-1 {
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
		case "enter", "ctrl+s":
			if msg.String() == "enter" && c.focusIndex != len(c.inputs) {
				break
			}

			return &c, func() tea.Msg {
				values := make(map[string]any)
				for _, input := range c.inputs {
					values[input.Name()] = input.Value()
				}
				return c.submitMsg(values)
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
	itemViews := make([]string, len(c.inputs))
	maxWidth := 0
	for i, input := range c.inputs {
		var inputView = lipgloss.NewStyle().Padding(0, 1).Render(input.View())
		if i == c.focusIndex {
			inputView = selectedBorder.Render(inputView)
		} else {
			inputView = normalBorder.Render(inputView)
		}

		var titleView string
		if input.Required() {
			asterisk := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("*")
			titleView = fmt.Sprintf("%s%s ", input.Name(), asterisk)
		} else {
			titleView = fmt.Sprintf("%s ", input.Name())
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
	cmds := make([]tea.Cmd, len(c.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range c.inputs {
		c.inputs[i], cmds[i] = c.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (c *Form) SetSize(width, height int) {
	c.width, c.height = width, height
	c.viewport.Height = max(0, height-2)
	for _, input := range c.inputs {
		input.SetWidth(width / 2)
	}

	c.renderInputs()

	itemsHeight := c.itemsHeight()
	c.viewport.Style = lipgloss.NewStyle().MarginTop(max((c.viewport.Height-itemsHeight)/2, 0))
	if c.viewport.Height > 0 {
		c.ScrollViewport()
	}
}

func (c *Form) View() string {
	separator := strings.Repeat("─", c.width)
	submitRow := lipgloss.NewStyle().Align(lipgloss.Right).Padding(0, 1).Width(c.width).Render(fmt.Sprintf("%s · %s", renderAction("Submit", "ctrl+s", false), renderAction("Focus Next", "tab", false)))
	return lipgloss.JoinVertical(lipgloss.Left, c.viewport.View(), separator, submitRow)
}