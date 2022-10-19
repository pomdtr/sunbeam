package pages

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
	"github.com/pomdtr/sunbeam/scripts"
)

type FormContainer struct {
	response     *scripts.FormResponse
	inputs       []textinput.Model
	focusIndex   int
	submitAction func(map[string]string) tea.Cmd
	width        int
	height       int
}

func NewFormContainer(response *scripts.FormResponse, submitAction func(map[string]string) tea.Cmd) *FormContainer {
	c := &FormContainer{
		inputs:       make([]textinput.Model, len(response.Items)),
		response:     response,
		submitAction: submitAction,
	}

	var t textinput.Model
	for i, arg := range response.Items {
		t = textinput.New()

		t.Prompt = "  "
		t.Placeholder = arg.Name
		t.CharLimit = 32

		c.inputs[i] = t
	}

	return c
}

func (c *FormContainer) headerView() string {
	return bubbles.SunbeamHeader(c.width)
}

func (c FormContainer) Init() tea.Cmd {
	if len(c.inputs) == 0 {
		return nil
	}
	c.inputs[0].Prompt = "> "
	return c.inputs[0].Focus()
}

func (c *FormContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	// Handle character input and blinking
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return c, PopCmd
		case tea.KeyEnter:
			values := make(map[string]string, len(c.inputs))
			for _, input := range c.inputs {
				values[input.Placeholder] = input.Value()
			}
			return c, c.submitAction(values)
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
					c.inputs[i].Prompt = "> "
					continue
				}
				// Remove focused state
				c.inputs[i].Blur()
				c.inputs[i].Prompt = "  "
			}

			return c, tea.Batch(cmds...)
		}
	}

	cmd := c.updateInputs(msg)
	return c, cmd
}

func (c FormContainer) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(c.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range c.inputs {
		c.inputs[i], cmds[i] = c.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (c *FormContainer) SetSize(width, height int) {
	c.width = width
	c.height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footerView())
}

func (c FormContainer) footerView() string {
	return bubbles.SunbeamFooterWithActions(c.width, c.response.Title, "Submit")
}

func (c *FormContainer) View() string {
	var formItems []string
	var itemView string
	for i := range c.inputs {
		itemView = "\n" + c.inputs[i].View()
		formItems = append(formItems, itemView)
	}

	form := lipgloss.JoinVertical(lipgloss.Left, formItems...)
	form = lipgloss.Place(c.width, c.height, lipgloss.Left, lipgloss.Top, form)

	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), form, c.footerView())
}
