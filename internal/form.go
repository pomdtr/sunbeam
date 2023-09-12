package internal

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg"
)

type DynamicForm struct {
	*StaticForm
	page *pkg.Form

	extension Extension
	command   pkg.Command
	params    pkg.CommandParams
}

func NewDynamicForm(extension Extension, command pkg.Command, params pkg.CommandParams) *DynamicForm {
	return &DynamicForm{
		StaticForm: NewStaticForm(command.Title),

		extension: extension,
		command:   command,
		params:    params,
	}
}

func (c *DynamicForm) Init() tea.Cmd {
	return tea.Batch(c.header.SetIsLoading(true), c.Reload)
}

func (c *DynamicForm) Reload() tea.Msg {
	output, err := c.extension.Run(c.command.Name, pkg.CommandInput{
		Params: c.params,
	})
	if err != nil {
		return err
	}

	if err := pkg.ValidatePage(output); err != nil {
		return err
	}

	var form pkg.Form
	if err := json.Unmarshal(output, &form); err != nil {
		return err
	}

	return form
}

func (c *DynamicForm) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case pkg.Form:
		formitems := make([]FormItem, 0)
		for _, item := range msg.Items {
			formitem, err := NewFormItem(item)
			if err != nil {
				return c, func() tea.Msg {
					return err
				}
			}

			formitems = append(formitems, *formitem)
		}

		c.SetInputs(formitems...)
		if msg.Title != "" {
			c.StaticForm.footer.title = msg.Title
		}
		return c, nil
	case SubmitFormMsg:
		return c, func() tea.Msg {
			for name, value := range msg.Values {
				c.page.Command.Params[name] = value
			}

			page, err := CommandToPage(c.extension, pkg.CommandRef{
				Name:   c.page.Command.Name,
				Params: c.page.Command.Params,
			})
			if err != nil {
				return err
			}

			return PushPageMsg{
				Page: page,
			}
		}
	}

	return c.StaticForm.Update(msg)
}

type StaticForm struct {
	items []FormItem

	width    int
	header   Header
	footer   Footer
	viewport viewport.Model

	scrollOffset int
	focusIndex   int
}

func NewStaticForm(title string, inputs ...pkg.FormItem) *StaticForm {
	header := NewHeader()
	viewport := viewport.New(0, 0)
	footer := NewFooter(title)

	return &StaticForm{
		header:   header,
		footer:   footer,
		viewport: viewport,
	}
}

func (c *StaticForm) SetInputs(inputs ...FormItem) {
	c.items = inputs

	c.footer.SetBindings(
		key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("⌃S", "Submit")),
	)

	if len(c.items) > 0 {
		c.footer.bindings = append(c.footer.bindings, key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Next Input")))
	}
}

func (c *StaticForm) SetIsLoading(isLoading bool) tea.Cmd {
	return c.header.SetIsLoading(isLoading)
}

func (c StaticForm) Init() tea.Cmd {
	if len(c.items) == 0 {
		return nil
	}
	return c.items[0].Focus()
}

func (c StaticForm) Focus() tea.Cmd {
	if len(c.items) == 0 {
		return nil
	}

	return c.items[c.focusIndex].Focus()
}

func (c *StaticForm) CurrentItem() FormInput {
	if c.focusIndex >= len(c.items) {
		return nil
	}
	return c.items[c.focusIndex]
}

func (c *StaticForm) ScrollViewport() {
	cursorOffset := 0
	for i := 0; i < c.focusIndex; i++ {
		cursorOffset += c.items[i].Height() + 2
	}

	if c.CurrentItem() == nil {
		return
	}
	maxRequiredVisibleHeight := cursorOffset + c.CurrentItem().Height() + 2
	for maxRequiredVisibleHeight > c.viewport.Height+c.scrollOffset {
		c.viewport.LineDown(1)
		c.scrollOffset += 1
	}

	for cursorOffset < c.scrollOffset {
		c.viewport.LineUp(1)
		c.scrollOffset -= 1
	}
}

func (c StaticForm) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return &c, func() tea.Msg {
				return PopPageMsg{}
			}
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

			c.renderInputs()
			if c.viewport.Height > 0 {
				c.ScrollViewport()
			}

			return &c, tea.Batch(cmds...)
		case tea.KeyCtrlS:
			values := make(map[string]any)
			for _, input := range c.items {
				if input.Value() == "" && !input.Optional {
					return &c, func() tea.Msg {
						return fmt.Errorf("required field %s is empty", input.Name)
					}
				}
				values[input.Name] = input.Value()
			}

			return &c, SubmitCmd(values)
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	if cmd = c.updateInputs(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	c.renderInputs()

	if c.header, cmd = c.header.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return &c, tea.Batch(cmds...)
}

func SubmitCmd(values map[string]any) tea.Cmd {
	return func() tea.Msg {
		return SubmitMsg(values)
	}
}

type SubmitMsg map[string]any

func (c *StaticForm) renderInputs() {
	selectedBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("13"))
	normalBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true)
	itemViews := make([]string, len(c.items))
	maxWidth := 0
	for i, input := range c.items {
		var inputView = lipgloss.NewStyle().Padding(0, 1).Render(input.View())
		if i == c.focusIndex {
			inputView = selectedBorder.Render(inputView)
		} else {
			inputView = normalBorder.Render(inputView)
		}

		var titleView string
		if input.Optional {
			titleView = fmt.Sprintf("%s ", input.Title)
		} else {
			asterisk := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("*")
			titleView = fmt.Sprintf("%s%s ", input.Title, asterisk)
		}

		itemViews[i] = lipgloss.JoinHorizontal(lipgloss.Center, lipgloss.NewStyle().Bold(true).Render(titleView), inputView)
		if lipgloss.Width(itemViews[i]) > maxWidth {
			maxWidth = lipgloss.Width(itemViews[i])
		}
	}

	for i := range itemViews {
		itemViews[i] = lipgloss.NewStyle().Width(maxWidth).Align(lipgloss.Right).Render(itemViews[i])
	}

	formView := lipgloss.JoinVertical(lipgloss.Left, itemViews...)
	formView = lipgloss.NewStyle().Width(c.footer.Width).Align(lipgloss.Center).Render(formView)

	c.viewport.SetContent(formView)
}

func (c StaticForm) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(c.items))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range c.items {
		c.items[i].FormInput, cmds[i] = c.items[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

type SubmitFormMsg struct {
	Id     string
	Values map[string]any
}

func (c *StaticForm) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width

	c.width = width
	for _, input := range c.items {
		input.SetWidth(width / 2)
	}
	c.viewport.Height = height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())

	c.renderInputs()
	if c.viewport.Height > 0 {
		c.ScrollViewport()
	}
}

func (c *StaticForm) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
