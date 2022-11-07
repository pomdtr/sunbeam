package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/utils"
)

type Form struct {
	header     Header
	footer     Footer
	viewport   viewport.Model
	items      []FormItem
	focusIndex int
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
	SetWidth(int)

	Value() any

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
		Id:        formItem.Name,
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

func (ta *TextArea) Value() any {
	return ta.Model.Value()
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

func (ti *TextInput) SetWidth(width int) {
	ti.Model.Width = utils.Max(width-2, 0)
}

func (ti *TextInput) Value() any {
	return ti.Model.Value()
}

func (ti *TextInput) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	var cmd tea.Cmd
	ti.Model, cmd = ti.Model.Update(msg)
	return ti, cmd
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
	title string
	width int
	label string

	focused bool
	checked bool
}

func NewCheckbox(formItem api.FormItem) Checkbox {
	return Checkbox{
		label: formItem.Label,
		title: formItem.Title,
		width: 40,
	}
}

func (cb *Checkbox) Focus() tea.Cmd {
	cb.focused = true
	return nil
}

func (cb *Checkbox) Blur() {
	cb.focused = false
}

func (cb *Checkbox) SetWidth(width int) {
	cb.width = width
}

func (cb Checkbox) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	if !cb.focused {
		return &cb, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			cb.Toggle()
		}
	}

	return &cb, nil
}

func (cb Checkbox) View() string {
	var checkbox string
	if cb.checked {
		checkbox = fmt.Sprintf(" [x] %s", cb.label)
	} else {
		checkbox = fmt.Sprintf(" [ ] %s", cb.label)
	}

	paddingRight := utils.Max(cb.width-len(checkbox), 0)
	return fmt.Sprintf("%s%s", checkbox, strings.Repeat(" ", paddingRight))
}

func (cb Checkbox) Value() any {
	return cb.checked
}

func (cb *Checkbox) Toggle() {
	cb.checked = !cb.checked
}

type DropDownItem struct {
	title string
	value string
}

func (d DropDownItem) Render(width int, selected bool) string {
	if selected {
		return fmt.Sprintf("* %s", d.title)
	}
	return fmt.Sprintf("  %s", d.title)
}

func (d DropDownItem) FilterValue() string {
	return d.title
}

type DropDown struct {
	id     string
	data   map[string]string
	filter Filter
	value  string
}

func NewDropDown(formItem api.FormItem) DropDown {
	choices := make([]FilterItem, len(formItem.Data))
	for i, formItem := range formItem.Data {
		choices[i] = DropDownItem{
			title: formItem.Title,
			value: formItem.Value,
		}
	}

	ti := textinput.New()
	ti.Prompt = " "
	ti.Placeholder = formItem.Placeholder
	ti.PlaceholderStyle = DefaultStyles.Secondary
	ti.Width = 38

	filter := NewFilter()
	filter.SetItems(choices)

	return DropDown{
		filter: filter,
		value:  "",
	}
}

func (dd *DropDown) SetWidth(width int) {
	dd.filter.Width = width - 2
	dd.filter.viewport.Width = width
}

func (d DropDown) View() string {
	modelView := d.filter.Model.View()
	paddingRight := 0
	if d.Value() == "" {
		paddingRight = utils.Max(0, d.filter.Width-lipgloss.Width(modelView)+2)
	}
	textInputView := fmt.Sprintf("%s%s", modelView, strings.Repeat(" ", paddingRight))

	if !d.filter.Focused() {
		return textInputView
	} else if d.value != "" && d.value == d.filter.Value() {
		return textInputView
	} else {
		d.filter.viewport.Height = len(d.filter.filtered)
		return lipgloss.JoinVertical(lipgloss.Left, textInputView, d.filter.View())
	}
}

func (d DropDown) Value() any {
	return d.value
}

func (d *DropDown) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	if !d.filter.Focused() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(d.filter.filtered) == 0 {
				return d, nil
			}
			selection := d.filter.Selection()
			dropDownItem, ok := selection.(DropDownItem)
			if !ok {
				return d, NewErrorCmd(fmt.Errorf("invalid selection type: %T", selection))
			}

			d.value = dropDownItem.value
			d.filter.SetValue(dropDownItem.title)
			d.filter.CursorEnd()

			return d, nil
		}
	}

	var cmd tea.Cmd
	d.filter, cmd = d.filter.Update(msg)

	return d, cmd
}

func (d *DropDown) Focus() tea.Cmd {
	return d.filter.Focus()
}

func (d *DropDown) Blur() {
	d.filter.Blur()
}

type SubmitMsg struct {
	values map[string]any
}

func NewSubmitCmd(values map[string]any) tea.Cmd {
	return func() tea.Msg {
		return SubmitMsg{values: values}
	}
}

func NewForm(title string, items []FormItem) *Form {
	header := NewHeader()
	viewport := viewport.New(0, 0)
	footer := NewFooter(title)
	footer.SetBindings(key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("⌃S", "Submit")))

	return &Form{
		header:   header,
		footer:   footer,
		viewport: viewport,
		items:    items,
	}
}

func (c Form) Init() tea.Cmd {
	if len(c.items) == 0 {
		return nil
	}
	return c.items[0].Focus()
}

func (c Form) Update(msg tea.Msg) (Container, tea.Cmd) {
	// Handle character input and blinking
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return &c, PopCmd
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

			return &c, tea.Batch(cmds...)
		case tea.KeyCtrlS:
			values := make(map[string]any)
			for _, input := range c.items {
				values[input.Id] = input.Value()
			}
			return &c, NewSubmitCmd(values)
		}
	}

	cmd := c.updateInputs(msg)

	return &c, cmd
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
	c.footer.Width = width
	c.viewport.Width = width
	c.header.Width = width
	for _, input := range c.items {
		input.SetWidth(width / 2)
	}
	c.viewport.Height = height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
}

func (c *Form) View() string {
	maxTitleWidth := 0
	for _, item := range c.items {
		if lipgloss.Width(item.Title) > maxTitleWidth {
			maxTitleWidth = lipgloss.Width(item.Title)
		}
	}

	selectedBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("205"))
	normalBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true)
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
	paddingLeft := (c.viewport.Width - maxTitleWidth - lipgloss.Width(formView)) / 2
	formView = lipgloss.NewStyle().Padding(1, paddingLeft).Render(formView)
	c.viewport.SetContent(formView)

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
