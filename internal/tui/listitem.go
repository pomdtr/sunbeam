package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

// Probably not necessary, need to be refactored
type ListItem types.ListItem

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

func RenderItem(title string, subtitle string, accessories []string, width int, selected bool) string {
	if width == 0 {
		return ""
	}
	title = strings.Split(title, "\n")[0]
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

	subtitle = strings.Split(subtitle, "\n")[0]
	subtitle = " " + subtitle
	var blanks string

	accessory := " " + strings.Join(accessories, " · ")

	// If the width is too small, we need to truncate the subtitle, accessory, or title (in that order)
	if width >= lipgloss.Width(title+subtitle+accessory) {
		availableWidth := width - lipgloss.Width(title+subtitle+accessory)
		blanks = strings.Repeat(" ", availableWidth)
	} else if width >= lipgloss.Width(title+accessory) {
		subtitle = subtitle[:width-lipgloss.Width(title+accessory)]
	} else if width >= lipgloss.Width(accessory) {
		subtitle = ""
		title = title[:width-lipgloss.Width(accessory)]
	} else {
		subtitle = ""
		accessory = ""
		title = title[:min(len(title), width)]
	}

	title = titleStyle.Render(title)
	subtitle = subtitleStyle.Render(subtitle)
	accessory = accessoryStyle.Render(accessory)

	return lipgloss.JoinHorizontal(lipgloss.Top, title, subtitle, blanks, accessory)

}

func (i ListItem) Render(width int, selected bool) string {
	return RenderItem(i.Title, i.Subtitle, i.Accessories, width, selected)
}
