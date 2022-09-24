package pages

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Page interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Page, tea.Cmd)
	View() string
	SetSize(width, height int)
}
