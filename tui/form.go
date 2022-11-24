package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/utils"
)

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

func NewFormItem(formItem app.ScriptParams) (FormItem, error) {
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
		return FormItem{}, fmt.Errorf("unknown form item type: %s", formItem.Type)
	}

	return FormItem{
		Required:  formItem.Required,
		Title:     formItem.Title,
		Id:        formItem.Name,
		FormInput: input,
	}, nil
}

type TextArea struct {
	textarea.Model
}

func NewTextArea(formItem app.ScriptParams) TextArea {
	ta := textarea.New()
	ta.FocusedStyle.Text = styles.Regular
	ta.BlurredStyle.Text = styles.Regular

	ta.Placeholder = formItem.Placeholder
	ta.SetHeight(5)

	return TextArea{
		Model: ta,
	}
}

func (ta *TextArea) SetWidth(w int) {
	ta.Model.SetWidth(w - 2)
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
	textinput.Model
	placeholder string
}

func NewTextInput(formItem app.ScriptParams) TextInput {
	ti := textinput.New()
	ti.Prompt = ""
	ti.TextStyle = styles.Regular.Copy()
	ti.PlaceholderStyle = styles.Regular.Copy()

	return TextInput{
		Model: ti,
	}
}

func (ti *TextInput) SetWidth(width int) {
	ti.Model.Width = width - 1
	placeholderPadding := width - len(ti.placeholder)
	ti.Model.Placeholder = fmt.Sprintf("%s%s", ti.placeholder, strings.Repeat(" ", placeholderPadding))
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
	return ti.Model.View()
}

type Checkbox struct {
	title string
	label string
	Style lipgloss.Style

	focused bool
	checked bool
}

func NewCheckbox(formItem app.ScriptParams) Checkbox {
	return Checkbox{
		Style: styles.Regular.Copy(),
		label: formItem.Label,
		title: formItem.Title,
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
	cb.Style.Width(width)
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
		checkbox = fmt.Sprintf("[x] %s", cb.label)
	} else {
		checkbox = fmt.Sprintf("[ ] %s", cb.label)
	}

	return cb.Style.Render(checkbox)
}

func (cb Checkbox) Value() any {
	return cb.checked
}

func (cb *Checkbox) Toggle() {
	cb.checked = !cb.checked
}

type DropDownItem struct {
	id    string
	title string
	value string
}

func (d DropDownItem) ID() string {
	return d.id
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
	filter    Filter
	textinput textinput.Model
	value     string
}

func NewDropDown(formItem app.ScriptParams) DropDown {
	choices := make([]FilterItem, len(formItem.Data))
	for i, formItem := range formItem.Data {
		choices[i] = DropDownItem{
			id:    strconv.Itoa(i),
			title: formItem.Title,
			value: formItem.Value,
		}
	}

	ti := textinput.New()
	ti.Prompt = " "
	ti.Placeholder = formItem.Placeholder
	ti.PlaceholderStyle = styles.Regular

	filter := NewFilter()
	filter.SetItems(choices)

	return DropDown{
		filter:    filter,
		textinput: ti,
		value:     "",
	}
}

func (dd *DropDown) SetWidth(width int) {
	dd.filter.Width = width - 2
}

func (d DropDown) View() string {
	modelView := d.textinput.View()
	paddingRight := 0
	if d.Value() == "" {
		paddingRight = utils.Max(0, d.filter.Width-lipgloss.Width(modelView)+2)
	}
	textInputView := fmt.Sprintf("%s%s", modelView, strings.Repeat(" ", paddingRight))

	if !d.textinput.Focused() {
		return textInputView
	} else if d.value != "" && d.value == d.textinput.Value() {
		return textInputView
	} else {
		d.filter.Height = len(d.filter.filtered)
		return lipgloss.JoinVertical(lipgloss.Left, textInputView, d.filter.View())
	}
}

func (d DropDown) Value() any {
	return d.value
}

func (d *DropDown) Update(msg tea.Msg) (FormInput, tea.Cmd) {
	if !d.textinput.Focused() {
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
			d.filter.FilterItems(dropDownItem.value)
			d.textinput.SetValue(dropDownItem.value)
			d.textinput.CursorEnd()

			return d, nil
		}
	}

	var cmd tea.Cmd
	d.filter, cmd = d.filter.Update(msg)

	return d, cmd
}

func (d *DropDown) Focus() tea.Cmd {
	return d.textinput.Focus()
}

func (d *DropDown) Blur() {
	d.textinput.Blur()
}

type Form struct {
	header     Header
	submitCmd  func(values map[string]any) tea.Cmd
	footer     Footer
	viewport   viewport.Model
	items      []FormItem
	focusIndex int
}

func NewForm(title string, items []FormItem, submitCmd func(values map[string]any) tea.Cmd) *Form {
	header := NewHeader()
	viewport := viewport.New(0, 0)
	footer := NewFooter(title)
	footer.SetBindings(
		key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("⌃S", "Submit")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Focus Next")),
	)

	return &Form{
		header:    header,
		footer:    footer,
		submitCmd: submitCmd,
		viewport:  viewport,
		items:     items,
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
			return &c, c.submitCmd(values)
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
	c.header.Width = width

	for _, input := range c.items {
		input.SetWidth(width / 2)
	}
	c.viewport.Height = height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
}

func (c *Form) View() string {
	selectedBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(accentColor)
	normalBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(theme.Fg())
	itemViews := make([]string, len(c.items))
	maxWidth := 0
	for i, item := range c.items {
		var inputView = item.FormInput.View()
		if i == c.focusIndex {
			inputView = selectedBorder.Render(inputView)
		} else {
			inputView = normalBorder.Render(inputView)
		}

		itemViews[i] = lipgloss.JoinHorizontal(lipgloss.Top, fmt.Sprintf("\n%s: ", item.Title), inputView)
		if lipgloss.Width(itemViews[i]) > maxWidth {
			maxWidth = lipgloss.Width(itemViews[i])
		}
	}

	for i := range itemViews {
		itemViews[i] = styles.Regular.Copy().Width(maxWidth).Align(lipgloss.Right).Render(itemViews[i])
	}

	formView := lipgloss.JoinVertical(lipgloss.Left, itemViews...)
	formView = styles.Regular.Copy().Width(c.footer.Width).Align(lipgloss.Center).PaddingTop(1).Render(formView)

	c.viewport.SetContent(formView)

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
