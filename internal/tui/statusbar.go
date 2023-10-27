package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type StatusBar struct {
	Width     int
	showInput bool
	input     textinput.Model
	isLoading bool
	actions   []types.Action
	expanded  bool
	cursor    int
	spinner   spinner.Model
}

func (c *StatusBar) SetActions(actions ...types.Action) {
	c.expanded = false
	c.cursor = 0
	c.actions = actions
}

func NewStatusBar(actions ...types.Action) StatusBar {
	spinner := spinner.New()
	spinner.Style = lipgloss.NewStyle().Padding(0, 1)
	return StatusBar{
		actions: actions,
		spinner: spinner,
	}
}

func (h *StatusBar) ShowInput() {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = ""
	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)

	h.showInput = true
	h.input = ti
}

func (h *StatusBar) Focus() tea.Cmd {
	if h.showInput {
		return h.input.Focus()
	}
	return nil
}

func (h StatusBar) Init() tea.Cmd {
	if h.isLoading {
		return h.spinner.Tick
	}
	return nil
}

func (h StatusBar) Value() string {
	return h.input.Value()
}

type IsLoadingMsg struct{}

func (p StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if len(p.actions) == 0 {
				return p, nil
			}

			if p.expanded {
				if p.cursor < len(p.actions)-1 {
					p.cursor++
				} else {
					p.cursor = 0
				}

				return p, nil
			}

			if p.showInput {
				p.input.Blur()
			}
			p.expanded = true
			return p, nil
		case "right":
			if !p.expanded {
				break
			}

			if p.cursor < len(p.actions)-1 {
				p.cursor++
			} else {
				p.cursor = 0
			}
			return p, nil

		case "left":
			if !p.expanded {
				break
			}

			if p.cursor > 0 {
				p.cursor--
			} else {
				p.cursor = len(p.actions) - 1
			}
		case "enter":
			if len(p.actions) == 0 {
				return p, nil
			}
			action := p.actions[p.cursor]
			p.expanded = false
			p.cursor = 0

			return p, func() tea.Msg {
				return action
			}

		case "esc":
			if p.expanded {
				if p.showInput {
					p.input.Focus()
				}
				p.expanded = false
				p.cursor = 0
				return p, nil
			}

			if p.showInput && p.input.Value() != "" {
				p.input.SetValue("")
				return p, nil
			}

			return p, PopPageCmd
		default:
			for _, action := range p.actions {
				if fmt.Sprintf("alt+%s", action.Key) == msg.String() {
					return p, func() tea.Msg {
						return action
					}
				}
			}
		}

	case IsLoadingMsg:
		cmd := p.SetIsLoading(true)
		return p, cmd
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	if p.showInput {
		input, cmd := p.input.Update(msg)
		p.input = input
		cmds = append(cmds, cmd)
	}

	if p.isLoading {
		p.spinner, cmd = p.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return p, tea.Batch(cmds...)
}

func (p *StatusBar) SetIsLoading(isLoading bool) tea.Cmd {
	p.isLoading = isLoading
	if isLoading {
		return p.spinner.Tick
	}
	return nil
}

func (c StatusBar) View() string {
	separator := strings.Repeat("─", c.Width)
	separator = lipgloss.NewStyle().Bold(true).Render(separator)

	var prefix string
	if c.isLoading {
		prefix = c.spinner.View()
	} else {
		prefix = "   "
	}

	var input string
	if c.showInput && c.input.Focused() {
		input = c.input.View()
	}

	var accessory string
	if len(c.actions) == 1 {
		accessory = renderAction(c.actions[0].Title, "enter", c.expanded)
	} else if len(c.actions) > 1 {
		if c.expanded {
			accessories := make([]string, len(c.actions))
			for i, action := range c.actions {
				var subtitle string
				if i == 0 {
					subtitle = "enter"
				} else if action.Key != "" {
					subtitle = fmt.Sprintf("alt+%s", action.Key)
				}
				accessories[i] = renderAction(action.Title, subtitle, i == c.cursor)
			}
			availableWidth := c.Width - lipgloss.Width(prefix) - 1
			for i, accessory := range accessories {
				availableWidth -= lipgloss.Width(accessory) + 3
				if availableWidth < 0 {
					accessories = accessories[1:]
				}

				if i == c.cursor {
					break
				}
			}

			accessory = strings.Join(accessories, " · ")
		} else {
			accessory = fmt.Sprintf("%s · Actions %s", renderAction(c.actions[0].Title, "enter", false), lipgloss.NewStyle().Faint(true).Render("tab"))
		}
	}
	accessory = fmt.Sprintf("%s ", accessory)

	var statusBar string
	if !c.expanded {
		availableWidth := c.Width - lipgloss.Width(prefix) - lipgloss.Width(input) - lipgloss.Width(accessory)
		blanks := strings.Repeat(" ", max(0, availableWidth))
		statusBar = lipgloss.JoinHorizontal(lipgloss.Top, prefix, input, blanks, accessory)
	} else {
		statusBar = fmt.Sprintf("%s%s", prefix, accessory)
	}

	return lipgloss.JoinVertical(lipgloss.Left, separator, statusBar)
}

func renderAction(title string, subtitle string, selected bool) string {
	var view string
	if subtitle != "" {
		view = fmt.Sprintf("%s %s", title, lipgloss.NewStyle().Faint(true).Render(subtitle))
	} else {
		view = title
	}

	if selected {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render(view)
	}

	return view
}
