package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	_ "embed"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

//go:embed templates/create-script-prompt.md
var createScriptMessageTemplate string
var createScriptMessage string

func init() {
	renderTemplate := func(templateStr string, data any) (string, error) {
		t, err := template.New("").Parse(templateStr)
		if err != nil {
			return "", err
		}

		var b strings.Builder
		if err := t.Execute(&b, data); err != nil {
			return "", err
		}
		return b.String(), nil
	}

	data := struct {
		Typescript string
	}{
		Typescript: types.TypeScript,
	}

	var err error
	createScriptMessage, err = renderTemplate(createScriptMessageTemplate, data)
	if err != nil {
		panic(err)
	}
}

var re = regexp.MustCompile("`{3}[\\w]*\n+([\\S\\s]+?\n)`{3}")

func extractMarkdownCodeblock(markdown string) string {
	match := re.FindStringSubmatch(markdown)
	if len(match) < 2 {
		return markdown
	}

	return match[1]
}

func NewCmdAsk() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ask",
		Short: "Ask a question",
		Long:  `Ask a question`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := args[0]
			token, ok := os.LookupEnv("OPENAI_API_KEY")
			if !ok {
				return fmt.Errorf("OPENAI_API_KEY not set")
			}

			messages := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: createScriptMessage,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			}

			generator := func() (*types.Page, error) {
				client := openai.NewClient(token)
				res, err := client.CreateChatCompletion(
					context.Background(),
					openai.ChatCompletionRequest{
						Model:    openai.GPT4,
						Messages: messages,
					},
				)
				if err != nil {
					return nil, err
				}

				code := extractMarkdownCodeblock(res.Choices[0].Message.Content)

				page := types.Page{
					Type:  types.DetailPage,
					Title: args[0],
					Preview: &types.Preview{
						Type:     types.StaticPreviewType,
						Language: "go",
						Text:     code,
					},
					Actions: []types.Action{
						{
							Type:  types.PushPageAction,
							Title: "Eval Code",
							Page: &types.Target{
								Input:   code,
								Type:    types.DynamicTarget,
								Command: "sunbeam eval",
							},
						},
						{
							Type:  types.CopyAction,
							Title: "Copy Code",
							Text:  code,
						},
						{
							Type:  types.PushPageAction,
							Title: "Edit Prompt",
							Page: &types.Target{
								Type:    types.DynamicTarget,
								Input:   code,
								Command: "sunbeam ask ${input:prompt}",
							},
							Inputs: []types.Input{
								{Type: types.TextFieldInput, Name: "prompt", Title: "Prompt", Default: prompt},
							},
						},
						{
							Type:      types.RunAction,
							Title:     "Save Code",
							Command:   "cp /dev/stdin ${input:filepath}",
							OnSuccess: types.ExitOnSuccess,
							Input:     code,
							Inputs: []types.Input{
								{Type: types.TextFieldInput, Name: "filepath", Title: "Filepath"},
							},
						},
					},
				}

				return &page, nil
			}

			if !isOutputInteractive() {
				output, err := generator()
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}

				json.NewEncoder(os.Stdout).Encode(output)
				return nil
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
			return nil
		},
	}

	return cmd
}
