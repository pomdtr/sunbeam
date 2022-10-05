package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/commands"
	"github.com/pomdtr/sunbeam/utils"
)

type PushMsg struct {
	Container Page
}

func NewPushCmd(cmd commands.Command) func() tea.Msg {
	return utils.SendMsg(PushMsg{Container: NewCommandContainer(cmd)})
}

type PopMsg struct{}

var PopCmd = utils.SendMsg(PopMsg{})
