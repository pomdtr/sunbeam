package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Input interface {
	Name() string
	Required() bool
	Value() any

	Focus() tea.Cmd
	Blur()

	Height() int

	SetWidth(int)
	Update(tea.Msg) (Input, tea.Cmd)
	View() string
}

type TextInput struct {
	name     string
	required bool
	textinput.Model
	placeholder string
}

func NewTextInput(param types.Param) *TextInput {
	ti := textinput.New()
	ti.Prompt = ""

	placeholder := param.Description
	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)

	return &TextInput{
		name:        param.Name,
		Model:       ti,
		placeholder: placeholder,
	}
}

func (ti *TextInput) Name() string {
	return ti.name
}

func (ti *TextInput) Required() bool {
	return ti.required
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

func (ti *TextInput) Update(msg tea.Msg) (Input, tea.Cmd) {
	var cmd tea.Cmd
	ti.Model, cmd = ti.Model.Update(msg)
	return ti, cmd
}

func (ti TextInput) View() string {
	return ti.Model.View()
}

type BooleanInput struct {
	name     string
	label    string
	width    int
	required bool

	focused bool
	checked bool
}

func NewBooleanInput(param types.Param) *BooleanInput {
	checkbox := BooleanInput{
		label: param.Description,
	}

	return &checkbox
}

func (cb *BooleanInput) Name() string {
	return cb.name
}

func (cb *BooleanInput) Required() bool {
	return cb.required
}

func (cb *BooleanInput) Height() int {
	return 1
}

func (cb *BooleanInput) Focus() tea.Cmd {
	cb.focused = true
	return nil
}

func (cb *BooleanInput) Blur() {
	cb.focused = false
}

func (cb *BooleanInput) SetWidth(width int) {
	cb.width = width
}

func (cb BooleanInput) Update(msg tea.Msg) (Input, tea.Cmd) {
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

func (cb BooleanInput) View() string {
	var checkbox string
	if cb.checked {
		checkbox = fmt.Sprintf("[x] %s", cb.label)
	} else {
		checkbox = fmt.Sprintf("[ ] %s", cb.label)
	}

	padding := max(0, cb.width-len(checkbox))

	return fmt.Sprintf("%s%s", checkbox, strings.Repeat(" ", padding))
}

func (cb BooleanInput) Value() any {
	return cb.checked
}

func (cb *BooleanInput) Toggle() {
	cb.checked = !cb.checked
}

type NumberInput struct {
	*TextInput
}

func NewNumberInput(param types.Param) Input {
	ni := NumberInput{
		TextInput: NewTextInput(param),
	}
	return &ni
}

func (ni *NumberInput) Value() any {
	text, ok := ni.TextInput.Value().(string)
	if !ok {
		return 0
	}

	value, err := strconv.Atoi(text)
	if err != nil {
		return 0
	}

	return value
}
