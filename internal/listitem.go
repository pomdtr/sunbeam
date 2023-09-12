package internal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg"
)

// Probably not necessary, need to be refactored
type ListItem pkg.ListItem

func (i ListItem) ID() string {
	if i.Id != "" {
		return i.Id
	}
	return i.Title
}

func (i ListItem) FilterValue() string {
	keywords := []string{i.Title, i.Subtitle}
	return strings.Join(keywords, " ")
}

func (i ListItem) Render(width int, selected bool) string {
	if width == 0 {
		return ""
	}

	title := strings.Split(i.Title, "\n")[0]
	titleStyle := lipgloss.NewStyle().Bold(true)
	subtitleStyle := lipgloss.NewStyle()
	accessoryStyle := lipgloss.NewStyle()
	if selected {
		title = fmt.Sprintf("> %s", title)
		titleStyle = titleStyle.Foreground(lipgloss.Color("13"))
		accessoryStyle = accessoryStyle.Foreground(lipgloss.Color("13"))
		subtitleStyle = subtitleStyle.Foreground(lipgloss.Color("13"))
	} else {
		subtitleStyle = subtitleStyle.Faint(true)
		accessoryStyle = accessoryStyle.Faint(true)
		title = fmt.Sprintf("  %s", title)
	}

	subtitle := strings.Split(i.Subtitle, "\n")[0]
	subtitle = " " + subtitle
	var blanks string

	accessories := " " + strings.Join(i.Accessories, " · ")

	// If the width is too small, we need to truncate the subtitle, accessories, or title (in that order)
	if width >= lipgloss.Width(title+subtitle+accessories) {
		availableWidth := width - lipgloss.Width(title+subtitle+accessories)
		blanks = strings.Repeat(" ", availableWidth)
	} else if width >= lipgloss.Width(title+accessories) {
		subtitle = subtitle[:width-lipgloss.Width(title+accessories)]
	} else if width >= lipgloss.Width(title) {
		subtitle = ""
		start_index := len(accessories) - (width - lipgloss.Width(title)) + 1
		accessories = " " + accessories[start_index:]
	} else {
		subtitle = ""
		accessories = ""
		// Why is this -1? I don't know, but it works
		title = title[:min(len(title), width)]
	}

	title = titleStyle.Render(title)
	subtitle = subtitleStyle.Render(subtitle)
	accessories = accessoryStyle.Render(accessories)

	return lipgloss.JoinHorizontal(lipgloss.Top, title, subtitle, blanks, accessories)
}
