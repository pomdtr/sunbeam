package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/scripts"
	"github.com/pomdtr/sunbeam/utils"
)

type FormInput interface {
	Focus() tea.Cmd
	Blur()

	Height() int
	Value() string
	Title() string

	SetWidth(int)
	Update(tea.Msg) (FormInput, tea.Cmd)
	View() string
}

func NewFormInput(param *scripts.FormInput) FormInput {
	switch param.Type {
	case "textfield", "file", "directory":
		return NewTextInput(param)
	case "password":
		ti := NewTextInput(param)
		ti.SetHidden()
		return ti
	case "textarea":
		return NewTextArea(param)
	case "dropdown":
		return NewDropDown(*param)
	default:
		return nil
	}
}

type TextArea struct {
	textarea.Model
	title string
}

func (ta *TextArea) Title() string {
	return ta.title
}

func NewTextArea(formItem *scripts.FormInput) *TextArea {
	ta := textarea.New()
	ta.Prompt = ""
	ta.SetValue(formItem.Default)

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

func (ta *TextArea) Value() string {
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

func NewTextInput(formItem *scripts.FormInput) *TextInput {
	ti := textinput.New()
	ti.Prompt = ""
	ti.SetValue(formItem.Default)

	placeholder := formItem.Placeholder
	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)

	return &TextInput{
		title:       formItem.Title,
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
	placeholderPadding := utils.Max(0, width-len(ti.placeholder))
	ti.Model.Placeholder = fmt.Sprintf("%s%s", ti.placeholder, strings.Repeat(" ", placeholderPadding))
}

func (ti *TextInput) Value() string {
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

type DropDownItem struct {
	id    string
	value string
}

func (d DropDownItem) ID() string {
	return d.id
}

func (d DropDownItem) Render(width int, selected bool) string {
	if selected {
		return fmt.Sprintf("* %s", d.value)
	}
	return fmt.Sprintf("  %s", d.value)
}

func (d DropDownItem) FilterValue() string {
	return d.value
}

type DropDown struct {
	title     string
	filter    Filter
	textinput textinput.Model
	items     map[string]DropDownItem
	selection DropDownItem
}

func NewDropDown(formItem scripts.FormInput) *DropDown {
	dropdown := DropDown{}
	dropdown.items = make(map[string]DropDownItem)

	choices := make([]FilterItem, len(formItem.Choices))
	for i, formItem := range formItem.Choices {
		item := DropDownItem{
			id:    strconv.Itoa(i),
			value: formItem,
		}

		choices[i] = item
		dropdown.items[choices[i].ID()] = item
	}

	ti := textinput.New()
	ti.Prompt = ""

	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)
	ti.Placeholder = formItem.Placeholder

	dropdown.textinput = ti

	filter := NewFilter()
	filter.SetItems(choices)
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
	placeholderPadding := utils.Max(0, width-len(dd.textinput.Placeholder))
	dd.textinput.Placeholder = fmt.Sprintf("%s%s", dd.textinput.Placeholder, strings.Repeat(" ", placeholderPadding))
	dd.filter.Width = width
}

func (dd DropDown) View() string {
	modelView := dd.textinput.View()
	paddingRight := 0
	if dd.Value() == "" {
		paddingRight = utils.Max(0, dd.filter.Width-lipgloss.Width(modelView))
	}
	textInputView := fmt.Sprintf("%s%s", modelView, strings.Repeat(" ", paddingRight))

	if !dd.textinput.Focused() || dd.HasMatch() {
		return textInputView
	} else {
		separator := strings.Repeat("─", dd.filter.Width)
		return lipgloss.JoinVertical(lipgloss.Left, textInputView, separator, dd.filter.View())
	}
}

func (d DropDown) Value() string {
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
			dropDownItem := selection.(DropDownItem)

			d.selection = dropDownItem
			d.textinput.SetValue(dropDownItem.value)
			d.filter.FilterItems(dropDownItem.value)
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
