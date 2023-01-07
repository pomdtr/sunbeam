package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/sunbeamlauncher/sunbeam/app"
	"github.com/sunbeamlauncher/sunbeam/tui"
	"github.com/sunbeamlauncher/sunbeam/utils"
)

type FormSpec struct {
	Command string
	Title   string
	Inputs  []app.ScriptInput
}

type FormExec struct {
	Command string
	*tui.Form
}

func (f FormExec) Update(msg tea.Msg) (tui.Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.SubmitMsg:
		funcMap := template.FuncMap{}
		for key, value := range msg.Values {
			key := key
			value := value
			funcMap[key] = func() any {
				return value
			}
		}

		command, err := utils.RenderString(f.Command, funcMap)
		if err != nil {
			return f, tui.NewErrorCmd(fmt.Errorf("failed to render command: %w", err))
		}

		return f, func() tea.Msg {
			return tui.ExecCommandMsg{
				Command: command,
			}
		}
	}

	form, cmd := f.Form.Update(msg)

	f.Form = form.(*tui.Form)

	return f, cmd
}

func NewCmdForm() *cobra.Command {
	return &cobra.Command{
		Use:     "form",
		Short:   "Show form view",
		GroupID: "core",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			var formSpec FormSpec
			err = json.Unmarshal(input, &formSpec)
			if err != nil {
				return err
			}
			if formSpec.Title == "" {
				formSpec.Title = "Form"
			}

			formitems := make([]tui.FormItem, len(formSpec.Inputs))
			for i, input := range formSpec.Inputs {
				formitems[i] = tui.NewFormItem(input)
			}

			form := tui.NewForm("form", formSpec.Title, formitems)
			formExec := FormExec{
				Form:    form,
				Command: formSpec.Command,
			}

			model := tui.NewModel(formExec)
			return tui.Draw(model)
		},
	}
}
