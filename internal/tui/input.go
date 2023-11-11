package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Input interface {
	Name() string
	Title() string
	Required() bool
	Value() any

	Focus() tea.Cmd
	Blur()

	Height() int

	SetWidth(int)
	Update(tea.Msg) (Input, tea.Cmd)
	View() string
}

type TextField struct {
	title    string
	name     string
	required bool
	textinput.Model
	placeholder string
}

func NewTextInput(param types.Input, secure bool) *TextField {
	ti := textinput.New()
	ti.Prompt = ""

	if secure {
		ti.EchoMode = textinput.EchoPassword
	}

	if defaultValue, ok := param.Default.(string); ok {
		ti.SetValue(defaultValue)
	}

	placeholder := param.Placeholder
	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)

	return &TextField{
		name:        param.Name,
		title:       param.Title,
		Model:       ti,
		placeholder: placeholder,
	}
}

func (ti *TextField) Name() string {
	return ti.name
}

func (ti *TextField) Title() string {
	return ti.title
}

func (ti *TextField) Required() bool {
	return ti.required
}

func (ti *TextField) Height() int {
	return 1
}

func (ti *TextField) SetWidth(width int) {
	ti.Model.Width = width - 1
	ti.Model.SetValue(ti.Model.Value())
	placeholderPadding := max(0, width-len(ti.placeholder))
	ti.Model.Placeholder = fmt.Sprintf("%s%s", ti.placeholder, strings.Repeat(" ", placeholderPadding))
}

func (ti *TextField) Value() any {
	return ti.Model.Value()
}

func (ti *TextField) Update(msg tea.Msg) (Input, tea.Cmd) {
	var cmd tea.Cmd
	ti.Model, cmd = ti.Model.Update(msg)
	return ti, cmd
}

func (ti TextField) View() string {
	return ti.Model.View()
}

type TextArea struct {
	textarea.Model
	title    string
	required bool
	name     string
}

func (ta *TextArea) Title() string {
	return ta.title
}

func (ta *TextArea) Required() bool {
	return ta.required
}

func (ta *TextArea) Name() string {
	return ta.name
}

func NewTextArea(input types.Input) Input {
	ta := textarea.New()
	ta.Prompt = ""

	if input.Default != nil {
		ta.SetValue(input.Default.(string))
	}

	ta.Placeholder = input.Placeholder
	ta.SetHeight(5)

	return &TextArea{
		title:    input.Title,
		required: input.Required,
		name:     input.Name,
		Model:    ta,
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

func (ta *TextArea) Update(msg tea.Msg) (Input, tea.Cmd) {
	model, cmd := ta.Model.Update(msg)
	ta.Model = model
	return ta, cmd
}

type Checkbox struct {
	name     string
	title    string
	label    string
	width    int
	required bool

	focused bool
	checked bool
}

func NewCheckbox(param types.Input) *Checkbox {
	checkbox := Checkbox{
		name:     param.Name,
		title:    param.Title,
		label:    param.Label,
		required: param.Required,
	}

	if defaultValue, ok := param.Default.(bool); ok {
		checkbox.checked = defaultValue
	}

	return &checkbox
}

func (cb *Checkbox) Name() string {
	return cb.name
}

func (cb *Checkbox) Title() string {
	return cb.title
}

func (cb *Checkbox) Required() bool {
	return cb.required
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

func (cb Checkbox) Update(msg tea.Msg) (Input, tea.Cmd) {
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
