package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	_ "embed"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

//go:embed templates/system-message.txt
var systemMessage string

var re = regexp.MustCompile("`{3}[\\w]*\n+([\\S\\s]+?\n)`{3}")

type Conversation struct {
}

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
						Content: systemMessage,
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

				page := types.Page{
					Type: types.DetailPage,
					Preview: &types.Preview{
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
							Type:      types.RunAction,
							Title:     "Preview Page",
							OnSuccess: types.PushOnSuccess,
							Command:   "sunbeam eval",
							Input:     code,
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
