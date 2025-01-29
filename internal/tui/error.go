package tui

import "github.com/pomdtr/sunbeam/pkg/sunbeam"

func NewErrorPage(err error, additionalActions ...sunbeam.Action) *Detail {
	var actions []sunbeam.Action
	actions = append(actions, sunbeam.Action{
		Title: "Copy error",
		Type:  sunbeam.ActionTypeCopy,
		Copy: &sunbeam.CopyAction{
			Text: err.Error(),
		},
	})
	actions = append(actions, additionalActions...)

	detail := NewDetail(err.Error(), actions...)

	return detail
}
