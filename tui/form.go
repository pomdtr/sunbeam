package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
)

type FormField struct {
	Id string
	textinput.Model
}

type Form struct {
	title      string
	footer     *Footer
	inputs     []FormField
	focusIndex int
	width      int
	height     int
}

type SubmitMsg struct {
	values map[string]string
}

func NewSubmitCmd(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		return SubmitMsg{values: values}
	}
}

func NewFormContainer(title string, params []api.SunbeamParam) *Form {
	footer := NewFooter()
	c := &Form{
		footer: footer,
		title:  title,
		inputs: make([]FormField, len(params)),
	}

	var t textinput.Model
	for i, param := range params {
		t = textinput.New()

		t.Prompt = fmt.Sprintf("  %s: ", param.Title)
		t.Placeholder = param.Placeholder
		t.CharLimit = 32

		c.inputs[i] = FormField{
			Id:    param.Id,
			Model: t,
		}
	}

	return c
}

func (c *Form) headerView() string {
	return SunbeamHeader(c.width)
}

func (c Form) Init() tea.Cmd {
	if len(c.inputs) == 0 {
		return nil
	}
	return c.inputs[0].Focus()
}

func (c *Form) Update(msg tea.Msg) (*Form, tea.Cmd) {
	// Handle character input and blinking
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return c, PopCmd
		case tea.KeyEnter:
			values := make(map[string]string, len(c.inputs))
			for _, input := range c.inputs {
				values[input.Id] = input.Value()
			}
			return c, NewSubmitCmd(values)
		// Set focus to next input
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
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

			return c, tea.Batch(cmds...)
		}
	}

	cmd := c.updateInputs(msg)
	return c, cmd
}

func (c Form) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(c.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range c.inputs {
		c.inputs[i].Model, cmds[i] = c.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (c *Form) SetSize(width, height int) {
	c.width = width
	c.height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View())
}

func (c *Form) View() string {
	var formItems []string
	var itemView string
	for i := range c.inputs {
		itemView = "\n" + c.inputs[i].View()
		formItems = append(formItems, itemView)
	}

	form := lipgloss.JoinVertical(lipgloss.Left, formItems...)
	form = lipgloss.Place(c.width, c.height, lipgloss.Left, lipgloss.Top, form)

	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), form, c.footer.View())
}
