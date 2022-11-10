package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Header struct {
	Width int
}

func NewHeader() Header {
	return Header{}
}

func (s Header) View() string {
	header := strings.Repeat(" ", s.Width)
	header = styles.Text.Render(header)
	separator := strings.Repeat("─", s.Width)
	separator = styles.Text.Render(separator)
	return lipgloss.JoinVertical(lipgloss.Left, header, separator)
}
