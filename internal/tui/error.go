package tui

import (
	"github.com/pomdtr/sunbeam/internal/types"
)

func NewErrorPage(err error, additionalActions ...types.Action) *Detail {
	var actions []types.Action
	actions = append(actions, types.Action{
		Title: "Copy error",
		Type:  types.ActionTypeCopy,
		Text:  err.Error(),
		Exit:  true,
	})
	actions = append(actions, additionalActions...)

	detail := NewDetail(err.Error(), actions...)

	return detail
}
