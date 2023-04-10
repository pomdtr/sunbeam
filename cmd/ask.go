package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	_ "embed"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

//go:embed templates/create-script-prompt.md
var createScriptMessageTemplate string
var createScriptMessage string

//go:embed templates/edit-script-prompt.md
var editScriptMessageTemplate string
var editScriptMessage string

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
	editScriptMessage, err = renderTemplate(editScriptMessageTemplate, data)
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
	log.Fatalf(createScriptMessage)
	cmd := &cobra.Command{
		Use:   "ask",
		Short: "Ask a question",
		Long:  `Ask a question`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, ok := os.LookupEnv("OPENAI_API_KEY")
			if !ok {
				return fmt.Errorf("OPENAI_API_KEY not set")
			}

			var messages []openai.ChatCompletionMessage
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				if err := json.Unmarshal(b, &messages); err != nil {
					return err
				}
			} else {
				messages = []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: createScriptMessage,
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: strings.Join(args, " "),
					},
				}
			}

			generator := func() ([]byte, error) {
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
				// code := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}"

				page := types.Page{
					Type: types.DetailPage,
					Preview: &types.Preview{
						Type:     types.StaticPreviewType,
						Language: "go",
						Text:     code,
					},
					Actions: []types.Action{
						{
							Type:  types.CopyAction,
							Title: "Copy Code",
							Text:  code,
						},
						{
							Type:  types.PushPageAction,
							Title: "Run Code",
							Page: &types.PageRef{
								Type:    "dynamic",
								Command: "sunbeam eval",
							},
							Input: code,
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

				return json.Marshal(page)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator()
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return nil
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
			return nil
		},
	}

	return cmd
}
