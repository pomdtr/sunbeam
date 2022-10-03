package pages

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
	"github.com/pomdtr/sunbeam/commands"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()
)

type FormContainer struct {
	response     *commands.FormResponse
	inputs       []textinput.Model
	focusIndex   int
	submitAction func(map[string]string) tea.Cmd
	width        int
	height       int
}

func NewFormContainer(response *commands.FormResponse, submitAction func(map[string]string) tea.Cmd) *FormContainer {
	c := &FormContainer{
		inputs:       make([]textinput.Model, len(response.Items)),
		response:     response,
		submitAction: submitAction,
	}

	var t textinput.Model
	for i, arg := range response.Items {
		t = textinput.New()
		if i == 0 {
			t.Focus()
		}

		t.Placeholder = arg.Name
		t.CursorStyle = cursorStyle
		t.CharLimit = 32

		c.inputs[i] = t
	}

	return c
}

func (c *FormContainer) Init() tea.Cmd {
	return nil
}

func (c *FormContainer) Update(msg tea.Msg) (Page, tea.Cmd) {
	// Handle character input and blinking
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if c.focusIndex < len(c.inputs)-1 {
				cmds := make([]tea.Cmd, len(c.inputs))
				c.focusIndex++
				for i := 0; i <= len(c.inputs)-1; i++ {
					if i == c.focusIndex {
						// Set focused state
						cmds[i] = c.inputs[i].Focus()
						c.inputs[i].PromptStyle = focusedStyle
						c.inputs[i].TextStyle = focusedStyle
						continue
					}
					// Remove focused state
					c.inputs[i].Blur()
					c.inputs[i].PromptStyle = noStyle
					c.inputs[i].TextStyle = noStyle
				}

				return c, tea.Batch(cmds...)
			}

			values := make(map[string]string, len(c.inputs))
			for _, input := range c.inputs {
				values[input.Placeholder] = input.Value()
			}
			return c, c.submitAction(values)
		// Set focus to next input
		case "tab", "shift+tab", "up", "down":
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
					c.inputs[i].PromptStyle = focusedStyle
					c.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				c.inputs[i].Blur()
				c.inputs[i].PromptStyle = noStyle
				c.inputs[i].TextStyle = noStyle
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
	c.height = height - lipgloss.Height(c.footerView())
}

func (c FormContainer) footerView() string {
	return bubbles.SunbeamFooterWithActions(c.width, c.response.Title, "Submit")
}

func (c *FormContainer) View() string {
	var formItems []string
	var itemView string
	for i := range c.inputs {
		itemView = c.inputs[i].View()
		if i != len(c.inputs)-1 {
			itemView += "\n"
		}
		formItems = append(formItems, itemView)
	}

	form := lipgloss.JoinVertical(lipgloss.Left, formItems...)
	form = lipgloss.Place(c.width, c.height, lipgloss.Left, lipgloss.Center, form)

	return lipgloss.JoinVertical(lipgloss.Left, form, c.footerView())
}
