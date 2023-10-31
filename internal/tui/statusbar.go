package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

func separator(n int) string {
	separator := strings.Repeat("─", n)
	return lipgloss.NewStyle().Bold(true).Render(separator)
}

type StatusBar struct {
	Width int

	cursor   int
	actions  []types.Action
	expanded bool
}

func NewStatusBar(actions ...types.Action) StatusBar {
	return StatusBar{
		actions: actions,
	}
}

func (c *StatusBar) SetActions(actions ...types.Action) {
	c.expanded = false
	c.cursor = 0
	c.actions = actions
}

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

			p.expanded = true
			return p, nil
		case "shift+tab":
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
		case "ctrl+d":
			if p.expanded {
				break
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
	}

	return p, nil
}

func (c StatusBar) View() string {
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

			availableWidth := c.Width
			availableWidth -= 3 * (len(c.actions) - 2) // 3 spaces between each action
			availableWidth -= 2                        // 2 spaces for the margins
			startIdx := 0
			endIdx := len(accessories)

			for lipgloss.Width(strings.Join(accessories[startIdx:endIdx], " · ")) > availableWidth {
				if endIdx-1 > c.cursor {
					endIdx--
				} else if startIdx < c.cursor {
					startIdx++
				} else {
					break
				}
			}

			accessory = strings.Join(accessories[startIdx:endIdx], " · ")
			if startIdx > 0 {
				accessory = fmt.Sprintf("… · %s", accessory)
			}
			if endIdx < len(accessories) {
				accessory = fmt.Sprintf("%s · …", accessory)
			}

		} else {
			accessory = fmt.Sprintf("%s · Actions %s", renderAction(c.actions[0].Title, "enter", false), lipgloss.NewStyle().Faint(true).Render("tab"))
		}
	}

	var statusbar string
	if c.expanded {
		blanks := strings.Repeat(" ", max(c.Width-lipgloss.Width(accessory)-2, 0))
		statusbar = fmt.Sprintf(" %s%s ", blanks, accessory)
	} else {
		blanks := strings.Repeat(" ", max(c.Width-lipgloss.Width(accessory)-2, 0))
		statusbar = fmt.Sprintf(" %s%s ", blanks, accessory)
	}

	return lipgloss.JoinVertical(lipgloss.Left, separator(c.Width), statusbar)
}

func renderAction(title string, subtitle string, selected bool) string {
	var view string
	if subtitle != "" {
		view = fmt.Sprintf("%s %s", title, lipgloss.NewStyle().Faint(true).Render(subtitle))
	} else {
		view = title
	}

	if selected {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true).Render(view)
	}

	return view
}
