package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/internal/fzf"
	"github.com/pomdtr/sunbeam/internal/types"
)

func separator(n int) string {
	separator := strings.Repeat("─", n)
	return lipgloss.NewStyle().Bold(true).Render(separator)
}

type StatusBar struct {
	Width int

	cursor   int
	actions  []types.Action
	filtered []types.Action
	expanded bool
}

func NewStatusBar(actions ...types.Action) StatusBar {
	return StatusBar{
		actions:  actions,
		filtered: actions,
	}
}

func (c *StatusBar) SetActions(actions ...types.Action) {
	c.expanded = false
	c.cursor = 0
	c.actions = actions
	c.filtered = actions
}

func (c *StatusBar) FilterActions(query string) {
	if query == "" {
		c.filtered = c.actions
		return
	}

	c.filtered = make([]types.Action, 0)
	scores := make([]int, 0)
	for i := 0; i < len(c.actions); i++ {
		score := fzf.Score(c.actions[i].Title, query)
		if score > 0 {
			c.filtered = append(c.filtered, c.actions[i])
			scores = append(scores, score)
		}
	}

	sort.SliceStable(c.filtered, func(i, j int) bool {
		return scores[i] > scores[j]
	})

	c.cursor = 0
}

func (p StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "right":
			if !p.expanded {
				break
			}

			if msg.String() == "tab" && len(p.filtered) == 0 {
				return p, nil
			}

			if p.cursor < len(p.filtered)-1 {
				p.cursor++
			} else {
				p.cursor = 0
			}

			return p, nil
		case "shift+tab", "left":
			if !p.expanded {
				break
			}

			if p.cursor > 0 {
				p.cursor--
			} else {
				p.cursor = len(p.filtered) - 1
			}
		case "enter":
			if len(p.filtered) == 0 {
				return p, nil
			}

			action := p.filtered[p.cursor]
			p.expanded = false
			p.cursor = 0
			return p, func() tea.Msg {
				return action
			}
		case "alt+enter":
			if p.cursor != 0 || len(p.actions) < 2 {
				break
			}

			p.expanded = false
			p.cursor = 0
			return p, func() tea.Msg {
				return p.actions[1]
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

func ActionTitle(action types.Action) string {
	if action.Title != "" {
		return action.Title
	}

	switch action.Type {
	case types.ActionTypeRun:
		return "Run"
	case types.ActionTypeCopy:
		return "Copy"
	case types.ActionTypeOpen:
		return "Open"
	case types.ActionTypeEdit:
		return "Edit"
	case types.ActionTypeReload:
		return "Reload"
	case types.ActionTypeExec:
		return "Exec"
	case types.ActionTypeExit:
		return "Exit"
	default:
		return string(action.Type)
	}
}

func (c StatusBar) View() string {
	var accessory string
	if len(c.actions) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, separator(c.Width), strings.Repeat(" ", c.Width))
	}
	if c.expanded {
		accessories := make([]string, len(c.filtered))
		for i, action := range c.filtered {
			var subtitle string
			if i == 0 {
				subtitle = "enter"
			} else if i == 1 {
				subtitle = "alt+enter"
			} else if action.Key != "" {
				subtitle = fmt.Sprintf("alt+%s", action.Key)
			}
			accessories[i] = renderAction(ActionTitle(action), subtitle, i == c.cursor)
		}

		availableWidth := c.Width
		availableWidth -= 3 * (len(c.filtered) - 2) // 3 spaces between each action
		availableWidth -= 2                         // 2 spaces for the margins
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
		accessory = fmt.Sprintf("%s · Actions %s", renderAction(ActionTitle(c.filtered[0]), "enter", false), lipgloss.NewStyle().Faint(true).Render("tab"))
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
