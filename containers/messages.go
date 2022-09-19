package containers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/utils"
)

type PushMsg struct {
	Container Container
}

func NewPushCmd(container Container) func() tea.Msg {
	return utils.SendMsg(PushMsg{Container: container})
}

type PopMsg struct{}

var PopCmd = utils.SendMsg(PopMsg{})
