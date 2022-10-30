package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
)

type Form struct {
	title      string
	footer     *Footer
	inputs     []FormInput
	focusIndex int
	width      int
	height     int
}

type FormInput interface {
	Focus() tea.Cmd
	Id() string
	Blur()
	Value() string
	View() string
	Update(tea.Msg) (FormInput, tea.Cmd)
}

func NewFormInput(formItem api.FormItem) FormInput {
	switch formItem.Type {
	case "textfield":
		ti := NewTextInput(formItem)
		return &ti
	case "textarea":
		ta := NewTextArea(formItem)
		return &ta
	case "dropdown":
		dd := NewDropDown(formItem)
		return &dd
	default:
		return nil
	}
}

type TextArea struct {
	id     string
	secure bool
	textarea.Model
}

func NewTextArea(formItem api.FormItem) TextArea {
	ta := textarea.New()
	ta.Placeholder = formItem.Placeholder
	ta.Prompt = "┃ "

	return TextArea{
		id:     formItem.Id,
		Model:  ta,
		secure: formItem.Secure,
	}
}

func (ta TextArea) Id() string {
	return ta.id
}

func (ta *TextArea) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	var cmd tea.Cmd
	ta.Model, cmd = ta.Model.Update(msg)
	return ta, cmd
}

type TextInput struct {
	id string
	textinput.Model
	secure bool
}

func NewTextInput(formItem api.FormItem) TextInput {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = formItem.Placeholder
	ti.PlaceholderStyle = DefaultStyles.Secondary
	ti.CharLimit = 32

	return TextInput{
		id:     formItem.Id,
		Model:  ti,
		secure: formItem.Secure,
	}
}

func (ti TextInput) Id() string {
	return ti.id
}

func (ti TextInput) View() string {
	if ti.secure {
		password := strings.Repeat("*", len(ti.Value()))
		ti.SetValue(password)
	}
	return ti.Model.View()
}

func (ta *TextInput) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	var cmd tea.Cmd
	ta.Model, cmd = ta.Model.Update(msg)
	return ta, cmd
}

type DropDown struct {
	id        string
	data      map[string]string
	textInput textinput.Model
	filter    Filter
	focused   bool
}

func NewDropDown(formItem api.FormItem) DropDown {
	choices := make([]string, len(formItem.Data))
	valueMap := make(map[string]string)
	for i, item := range formItem.Data {
		choices[i] = item.Title
		valueMap[item.Title] = item.Value
	}

	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = formItem.Placeholder
	ti.PlaceholderStyle = DefaultStyles.Secondary
	ti.Width = 20

	viewport := viewport.New(20, 3)
	filter := Filter{
		choices:    choices,
		indicator:  "* ",
		matches:    matchAll(choices),
		matchStyle: DefaultStyles.Selection,
		viewport:   &viewport,
		height:     3,
	}

	return DropDown{
		id:        formItem.Id,
		textInput: ti,
		filter:    filter,
	}
}

func (d DropDown) View() string {
	if !d.focused || d.Value() != "" {
		return d.textInput.View()
	} else {
		return lipgloss.JoinVertical(lipgloss.Left, d.textInput.View(), d.filter.View())
	}
}

func (d DropDown) Id() string {
	return d.id
}

func (d DropDown) Value() string {
	if len(d.filter.matches) > 0 && d.filter.matches[0].Str == d.textInput.Value() {
		return d.textInput.Value()
	}
	return ""
}

func (d DropDown) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	if !d.focused {
		return &d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(d.filter.matches) == 0 {
				return &d, nil
			}
			selection := d.filter.matches[d.filter.cursor]
			d.textInput.SetValue(selection.Str)
			d.textInput.CursorEnd()
			d.filter.Filter(selection.Str)
		}
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd
	ti, cmd := d.textInput.Update(msg)
	cmds = append(cmds, cmd)

	if ti.Value() != d.textInput.Value() {
		d.filter.Filter(ti.Value())
	} else {
		d.filter, cmd = d.filter.Update(msg)
		cmds = append(cmds, cmd)
	}

	d.textInput = ti
	return &d, tea.Batch(cmds...)
}

func (d *DropDown) Focus() tea.Cmd {
	d.focused = true
	return nil
}

func (d DropDown) Blur() {
	d.textInput.Blur()
}

type SubmitMsg struct {
	values map[string]string
}

func NewSubmitCmd(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		return SubmitMsg{values: values}
	}
}

func NewFormContainer(inputs []FormInput) *Form {
	footer := NewFooter()

	return &Form{
		footer: footer,
		inputs: inputs,
	}
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
		case tea.KeyCtrlS:
			values := make(map[string]string, len(c.inputs))
			for _, input := range c.inputs {
				values[input.Id()] = input.Value()
			}
			return c, NewSubmitCmd(values)
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
		c.inputs[i], cmds[i] = c.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (c *Form) SetSize(width, height int) {
	c.width = width
	c.height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View())
	c.footer.Width = width
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
