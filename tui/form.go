package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/utils"
)

type Form struct {
	title      string
	footer     *Footer
	viewport   viewport.Model
	items      []FormItem
	focusIndex int
	width      int
	height     int
}

type FormItem struct {
	Required bool
	Title    string
	Id       string
	FormInput
}

type FormInput interface {
	Focus() tea.Cmd
	Blur()

	Value() string

	View() string
	Update(tea.Msg) (FormInput, tea.Cmd)
}

func NewFormItem(formItem api.FormItem) FormItem {
	var input FormInput
	switch formItem.Type {
	case "textfield":
		ti := NewTextInput(formItem)
		input = &ti
	case "textarea":
		ta := NewTextArea(formItem)
		input = &ta
	case "dropdown":
		dd := NewDropDown(formItem)
		input = &dd
	case "checkbox":
		cb := NewCheckbox(formItem)
		input = &cb
	default:
		input = nil
	}

	return FormItem{
		Required:  formItem.Required,
		Title:     formItem.Title,
		Id:        formItem.Id,
		FormInput: input,
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
	ta.Prompt = ""
	ta.SetWidth(40)
	ta.SetHeight(5)

	return TextArea{
		Model:  ta,
		secure: formItem.Secure,
	}
}

func (ta *TextArea) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	var cmd tea.Cmd
	ta.Model, cmd = ta.Model.Update(msg)
	return ta, cmd
}

type TextInput struct {
	id string
	textinput.Model
}

func NewTextInput(formItem api.FormItem) TextInput {
	ti := textinput.New()
	ti.Placeholder = formItem.Placeholder
	ti.PlaceholderStyle = DefaultStyles.Secondary
	ti.Width = 38
	ti.Prompt = " "
	if formItem.Secure {
		ti.EchoMode = textinput.EchoPassword
	}

	return TextInput{
		Model: ti,
	}
}

func (ta *TextInput) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	var cmd tea.Cmd
	ta.Model, cmd = ta.Model.Update(msg)
	return ta, cmd
}

func (ti TextInput) View() string {
	modelView := ti.Model.View()
	paddingRight := 0
	if ti.Model.Value() == "" {
		paddingRight = ti.Width - lipgloss.Width(modelView) + 2
	}
	return fmt.Sprintf("%s%s", modelView, strings.Repeat(" ", paddingRight))
}

type Checkbox struct {
	id                 string
	focused            bool
	checked            bool
	title              string
	width              int
	label              string
	true_substitution  string
	false_substitution string
}

func NewCheckbox(formItem api.FormItem) Checkbox {
	return Checkbox{
		label:              formItem.Label,
		title:              formItem.Title,
		width:              40,
		true_substitution:  formItem.TrueSubstitution,
		false_substitution: formItem.FalseSubstitution,
	}
}

func (c *Checkbox) Focus() tea.Cmd {
	c.focused = true
	return nil
}

func (c *Checkbox) Blur() {
	c.focused = false
}

func (c Checkbox) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	if !c.focused {
		return &c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			c.checked = !c.checked
		}
	}

	return &c, nil
}

func (c Checkbox) View() string {
	var checkbox string
	if c.checked {
		checkbox = fmt.Sprintf(" [ ] %s", c.label)
	} else {
		checkbox = fmt.Sprintf(" [x] %s", c.label)
	}

	paddingRight := utils.Max(c.width-len(checkbox), 0)
	return fmt.Sprintf("%s%s", checkbox, strings.Repeat(" ", paddingRight))
}

func (c Checkbox) Value() string {
	if c.checked {
		return c.true_substitution
	}

	return c.false_substitution
}

type DropDown struct {
	id        string
	data      map[string]string
	textInput textinput.Model
	selection string
	filter    Filter
}

func NewDropDown(formItem api.FormItem) DropDown {
	choices := make([]string, len(formItem.Data))
	valueMap := make(map[string]string)
	for i, item := range formItem.Data {
		choices[i] = item.Title
		valueMap[item.Title] = item.Value
	}

	ti := textinput.New()
	ti.Prompt = " "
	ti.Placeholder = formItem.Placeholder
	ti.PlaceholderStyle = DefaultStyles.Secondary
	ti.Width = 38

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
		textInput: ti,
		filter:    filter,
	}
}

func (d DropDown) View() string {
	modelView := d.textInput.View()
	paddingRight := 0
	if d.textInput.Value() == "" {
		paddingRight = d.textInput.Width - lipgloss.Width(modelView) + 2
	}
	textInputView := fmt.Sprintf("%s%s", modelView, strings.Repeat(" ", paddingRight))
	if !d.textInput.Focused() {
		return textInputView
	} else if d.selection != "" && d.selection == d.textInput.Value() {
		return textInputView
	} else if len(d.filter.matches) == 0 {
		return textInputView
	} else {
		separator := strings.Repeat("─", d.textInput.Width+2)
		d.filter.viewport.Height = len(d.filter.matches)
		return lipgloss.JoinVertical(lipgloss.Left, textInputView, separator, d.filter.View())
	}
}

func (d DropDown) Value() string {
	return d.selection
}

func (d DropDown) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	if !d.textInput.Focused() {
		return &d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(d.filter.matches) == 0 {
				return &d, nil
			}
			selection := d.filter.matches[d.filter.cursor].Str
			d.selection = selection
			d.textInput.SetValue(selection)
			d.textInput.CursorEnd()
			return &d, nil
		}
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd
	ti, cmd := d.textInput.Update(msg)
	cmds = append(cmds, cmd)

	if ti.Value() != d.textInput.Value() {
		d.selection = ""
		d.filter.Filter(ti.Value())
	} else {
		d.filter, cmd = d.filter.Update(msg)
		cmds = append(cmds, cmd)
	}

	d.textInput = ti
	return &d, tea.Batch(cmds...)
}

func (d *DropDown) Focus() tea.Cmd {
	return d.textInput.Focus()
}

func (d *DropDown) Blur() {
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

func NewForm(items []FormItem) *Form {
	footer := NewFooter()
	viewport := viewport.New(0, 0)

	return &Form{
		footer:   footer,
		viewport: viewport,
		items:    items,
	}
}

func (c *Form) headerView() string {
	return SunbeamHeader(c.width)
}

func (c Form) Init() tea.Cmd {
	if len(c.items) == 0 {
		return nil
	}
	return c.items[0].Focus()
}

func (c *Form) Update(msg tea.Msg) (*Form, tea.Cmd) {
	// Handle character input and blinking
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return c, PopCmd
		case tea.KeyCtrlS:
			values := make(map[string]string, len(c.items))
			for _, input := range c.items {
				values[input.Id] = input.Value()
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

			return c, tea.Batch(cmds...)
		}
	}

	cmd := c.updateInputs(msg)
	return c, cmd
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
	c.width = width
	c.footer.Width = width
	c.viewport.Height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View())
	c.viewport.Width = width
}

func (c *Form) View() string {

	maxTitleWidth := 0
	for _, item := range c.items {
		if lipgloss.Width(item.Title) > maxTitleWidth {
			maxTitleWidth = lipgloss.Width(item.Title)
		}
	}

	selectedBorder := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("205"))
	normalBorder := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true)
	itemViews := make([]string, len(c.items))
	for i, item := range c.items {
		paddingLeft := maxTitleWidth - lipgloss.Width(item.Title)
		var inputView = item.FormInput.View()
		if i == c.focusIndex {
			inputView = selectedBorder.Render(inputView)
		} else {
			inputView = normalBorder.Render(inputView)
		}

		itemViews[i] = lipgloss.JoinHorizontal(lipgloss.Top, strings.Repeat(" ", paddingLeft), "\n"+item.Title, "  ", inputView)
	}

	formView := lipgloss.JoinVertical(lipgloss.Left, itemViews...)
	paddingLeft := (c.width - lipgloss.Width(formView)) / 2
	c.viewport.SetContent(lipgloss.NewStyle().PaddingLeft(paddingLeft - maxTitleWidth).PaddingTop(1).Render(formView))

	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.viewport.View(), c.footer.View())
}
