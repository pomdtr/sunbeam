package tui

import (
	"github.com/pomdtr/sunbeam/pkg/types"
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
	detail.Format = types.MarkdownFormat

	return detail
}
