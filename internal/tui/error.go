package tui

import (
	"github.com/pomdtr/sunbeam/pkg/types"
)

func NewErrorPage(err error, additionalActions ...types.Action) *Detail {
	var actions []types.Action
	actions = append(actions, types.Action{
		Title: "Copy error",
		OnAction: types.Command{
			Type: types.CommandTypeCopy,
			Text: err.Error(),
			Exit: true,
		},
	})
	actions = append(actions, additionalActions...)

	return NewDetail(err.Error(), actions...)
}
