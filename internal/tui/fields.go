package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type FormInput interface {
	Focus() tea.Cmd
	Blur()

	Height() int
	Value() any

	SetWidth(int)
	Update(tea.Msg) (FormInput, tea.Cmd)
	View() string
}

type FormItem struct {
	FormInput
	Required bool
	Title    string
	Name     string
}

func NewFormItem(item types.Field) (FormItem, error) {
	var input FormInput
	switch item.Type {
	case types.TextInput:
		input = NewTextInput(item)
	case types.TextAreaInput:
		input = NewTextArea(item)
	case types.CheckboxInput:
		input = NewCheckbox(item)
	case types.SelectInput:
		input = NewDropDown(item)
	default:
		return FormItem{}, fmt.Errorf("unknown input type %s", item.Type)
	}

	return FormItem{
		Name:      item.Name,
		Title:     item.Title,
		Required:  item.Required,
		FormInput: input,
	}, nil
}

type TextArea struct {
	textarea.Model
	title string
}

func (ta *TextArea) Title() string {
	return ta.title
}

func NewTextArea(formItem types.Field) *TextArea {
	ta := textarea.New()
	ta.Cursor.SetMode(cursor.CursorStatic)
	ta.Prompt = ""
	if formItem.Default != nil {
		ta.SetValue(formItem.Default.(string))
	}

	ta.Placeholder = formItem.Placeholder
	ta.SetHeight(5)

	return &TextArea{
		Model: ta,
		title: formItem.Title,
	}
}

func (ta *TextArea) Height() int {
	return ta.Model.Height()
}

func (ta *TextArea) SetWidth(w int) {
	ta.Model.SetWidth(w)
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
	title       string
	placeholder string
}

func NewTextInput(input types.Field) *TextInput {
	ti := textinput.New()
	ti.Cursor.SetMode(cursor.CursorStatic)
	ti.Prompt = ""
	if input.Default != nil {
		ti.SetValue(input.Default.(string))
	}

	placeholder := input.Placeholder
	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)

	return &TextInput{
		title:       input.Title,
		Model:       ti,
		placeholder: placeholder,
	}
}

func (ti *TextInput) Title() string {
	return ti.title
}

func (ti *TextInput) SetHidden() {
	ti.EchoMode = textinput.EchoPassword
}

func (ti *TextInput) Height() int {
	return 1
}

func (ti *TextInput) SetWidth(width int) {
	ti.Model.Width = width - 1
	ti.Model.SetValue(ti.Model.Value())
	placeholderPadding := max(0, width-len(ti.placeholder))
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
	width int

	focused bool
	checked bool
}

func NewCheckbox(input types.Field) *Checkbox {
	var defaultValue bool
	if input.Default != nil {
		defaultValue = input.Default.(bool)
	}

	return &Checkbox{
		label:   input.Label,
		title:   input.Title,
		checked: defaultValue,
	}
}

func (cb *Checkbox) Height() int {
	return 1
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
		checkbox = fmt.Sprintf("[x] %s", cb.label)
	} else {
		checkbox = fmt.Sprintf("[ ] %s", cb.label)
	}

	padding := max(0, cb.width-len(checkbox))

	return fmt.Sprintf("%s%s", checkbox, strings.Repeat(" ", padding))
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
	value any
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
	title     string
	filter    Filter
	textinput textinput.Model
	items     map[string]DropDownItem
	selection DropDownItem
}

func NewDropDown(formItem types.Field) *DropDown {
	dropdown := DropDown{}
	dropdown.items = make(map[string]DropDownItem)

	var defaultValue string
	if formItem.Default != nil {
		defaultValue = formItem.Default.(string)
	}

	choices := make([]FilterItem, len(formItem.Choices))
	var defaultId string
	for i, formItem := range formItem.Choices {

		item := DropDownItem{
			id:    strconv.Itoa(i),
			title: formItem.Title,
			value: formItem.Value,
		}
		if formItem.Value == defaultValue {
			defaultId = item.ID()
			dropdown.selection = item
		}

		choices[i] = item
		dropdown.items[choices[i].ID()] = item
	}

	ti := textinput.New()
	ti.Cursor.SetMode(cursor.CursorStatic)
	ti.SetValue(defaultValue)
	ti.Prompt = ""

	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)
	ti.Placeholder = formItem.Placeholder

	dropdown.textinput = ti

	filter := NewFilter()
	filter.SetItems(choices...)
	if defaultId != "" {
		filter.Select(defaultId)
	}
	filter.FilterItems("")
	filter.DrawLines = false
	filter.Height = 3

	dropdown.filter = filter
	dropdown.title = formItem.Title

	return &dropdown
}

func (dd DropDown) HasMatch() bool {
	return dd.selection.id != "" && dd.selection.value == dd.textinput.Value()
}

func (dd *DropDown) Height() int {
	return 5
}

func (dd *DropDown) SetWidth(width int) {
	dd.textinput.Width = width - 1
	placeholderPadding := max(0, width-len(dd.textinput.Placeholder))
	dd.textinput.Placeholder = fmt.Sprintf("%s%s", dd.textinput.Placeholder, strings.Repeat(" ", placeholderPadding))
	dd.filter.Width = width
}

func (dd DropDown) View() string {
	modelView := dd.textinput.View()
	paddingRight := 0
	if dd.Value() == "" {
		paddingRight = max(0, dd.filter.Width-lipgloss.Width(modelView))
	}
	textInputView := fmt.Sprintf("%s%s", modelView, strings.Repeat(" ", paddingRight))

	if !dd.textinput.Focused() || dd.HasMatch() {
		return textInputView
	} else {
		separator := strings.Repeat("─", dd.filter.Width)
		return lipgloss.JoinVertical(lipgloss.Left, textInputView, separator, dd.filter.View())
	}
}

func (d DropDown) Value() any {
	return d.selection.value
}

func (d DropDown) Title() string {
	return d.title
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
			if selection != nil {
				return d, nil
			}

			dropDownItem := selection.(DropDownItem)
			d.selection = dropDownItem
			switch value := dropDownItem.value.(type) {
			case string:
				d.textinput.SetValue(value)
				d.filter.FilterItems(value)
			case int:
				str := strconv.Itoa(value)
				d.textinput.SetValue(str)
				d.filter.FilterItems(str)
			case bool:
				d.textinput.SetValue(strconv.FormatBool(value))
				d.filter.FilterItems(strconv.FormatBool(value))
			}

			d.textinput.CursorEnd()

			return d, nil
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	ti, cmd := d.textinput.Update(msg)
	cmds = append(cmds, cmd)
	if ti.Value() != d.textinput.Value() {
		d.filter.FilterItems(ti.Value())
	}
	d.textinput = ti

	d.filter, cmd = d.filter.Update(msg)
	cmds = append(cmds, cmd)

	return d, tea.Batch(cmds...)
}

func (d *DropDown) Focus() tea.Cmd {
	return d.textinput.Focus()
}

func (d *DropDown) Blur() {
	d.textinput.Blur()
}
