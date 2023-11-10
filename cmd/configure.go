package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/internal/config"
	preferences "github.com/pomdtr/sunbeam/internal/storage"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
)

func NewCmdConfigure(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:       "configure",
		Short:     "Configure extension preferences",
		GroupID:   CommandGroupCore,
		ValidArgs: cfg.Aliases(),
		Args:      cobra.MatchAll(cobra.OnlyValidArgs, cobra.MaximumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			extension, err := LoadExtension(alias, cfg.Extensions[alias])
			if err != nil {
				return err
			}

			if len(extension.Preferences) == 0 {
				return fmt.Errorf("no preferences to configure for %s", alias)
			}

			prefs, err := preferences.Load(alias, extension.Origin)
			if err != nil {
				return err
			}

			inputs := make([]types.Input, len(extension.Preferences))
			for _, input := range extension.Preferences {
				if v, ok := prefs[input.Name]; ok {
					input.Default = v
				}
				inputs = append(inputs, input)
			}

			var formError error
			form := tui.NewForm(func(m map[string]any) tea.Msg {
				formError = preferences.Save(alias, extension.Origin, m)
				return tui.ExitMsg{}
			}, inputs...)

			if err := tui.Draw(form); err != nil {
				return err
			}

			if formError != nil {
				return formError
			}

			return nil
		},
	}
}
