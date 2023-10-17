package tui

import "github.com/pomdtr/sunbeam/pkg/types"

func NewErrorPage(err error) *Detail {
	return NewDetail(err.Error(), types.Action{
		Title: "Copy error",
		OnAction: types.Command{
			Type: types.CommandTypeCopy,
			Text: err.Error(),
		},
	})
}
