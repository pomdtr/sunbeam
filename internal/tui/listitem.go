package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/internal/types"
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
	return strings.Trim(strings.Join(keywords, " "), " ")
}

func RenderItem(title string, subtitle string, accessories []string, width int, selected bool) string {
	if width == 0 {
		return ""
	}
	title = strings.Split(title, "\n")[0]
	titleStyle := lipgloss.NewStyle()
	subtitleStyle := lipgloss.NewStyle()
	accessoryStyle := lipgloss.NewStyle()
	if selected {
		title = fmt.Sprintf("> %s", title)
		titleStyle = titleStyle.Foreground(lipgloss.Color("13")).Bold(true)
		accessoryStyle = accessoryStyle.Foreground(lipgloss.Color("13"))
		subtitleStyle = subtitleStyle.Foreground(lipgloss.Color("13"))
	} else {
		subtitleStyle = subtitleStyle.Faint(true)
		accessoryStyle = accessoryStyle.Faint(true)
		title = fmt.Sprintf("  %s", title)
	}

	subtitle = strings.Split(subtitle, "\n")[0]
	subtitle = " " + subtitle
	accessory := "  " + strings.Join(accessories, " · ")

	var blanks string

	// If the width is too small, we need to truncate the subtitle, title and accessory
	for lipgloss.Width(title+subtitle+accessory) > width {
		if words := strings.Split(subtitle, " "); len(words) > 1 {
			subtitle = strings.Join(words[:len(words)-1], " ")
		} else if len(accessory) > 0 {
			accessory = accessory[:len(accessory)-1]
		} else {
			title = title[:len(title)-1]
		}
	}

	extraWidth := width - lipgloss.Width(title+subtitle+accessory)
	blanks = strings.Repeat(" ", extraWidth)

	title = titleStyle.Render(title)
	subtitle = subtitleStyle.Render(subtitle)
	accessory = accessoryStyle.Render(accessory)

	return lipgloss.JoinHorizontal(lipgloss.Top, title, subtitle, blanks, accessory)

}

func (i ListItem) Render(width int, selected bool) string {
	return RenderItem(i.Title, i.Subtitle, i.Accessories, width, selected)
}
