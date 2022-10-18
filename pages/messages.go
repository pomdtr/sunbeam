package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/utils"
)

type PushMsg struct {
	Page Container
}

func NewPushCmd(page Container) func() tea.Msg {
	return utils.SendMsg(PushMsg{Page: page})
}

type PopMsg struct{}

var PopCmd = utils.SendMsg(PopMsg{})
