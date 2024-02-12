package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mitchellh/mapstructure"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

func NewCmdStd() *cobra.Command {
	cmd := &cobra.Command{
		Hidden: true,
		Args:   cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				manifest := sunbeam.Manifest{
					Title:       "Std",
					Description: `The standard library contains a set of commands that are built into sunbeam.`,
					Commands: []sunbeam.Command{
						{
							Name:  "copy",
							Title: "Copy to Clipboard",
							Mode:  sunbeam.CommandModeSilent,
							Params: []sunbeam.Param{
								{
									Name:  "text",
									Title: "Text",
									Type:  sunbeam.ParamString,
								},
							},
						},
						{
							Name:  "open",
							Title: "Open URL or Path",
							Mode:  sunbeam.CommandModeSilent,
							Params: []sunbeam.Param{
								{
									Name:     "url",
									Title:    "URL",
									Type:     sunbeam.ParamString,
									Optional: true,
								},
								{
									Name:     "path",
									Title:    "Path",
									Type:     sunbeam.ParamString,
									Optional: true,
								},
							},
						},
						{
							Name:  "edit",
							Title: "Edit File",
							Mode:  sunbeam.CommandModeTTY,
							Params: []sunbeam.Param{
								{
									Name:  "path",
									Title: "Path",
									Type:  sunbeam.ParamString,
								},
							},
						},
						{
							Name:  "run",
							Title: "Run Command",
							Mode:  sunbeam.CommandModeTTY,
							Params: []sunbeam.Param{
								{
									Name:  "command",
									Title: "Command",
									Type:  sunbeam.ParamString,
								},
								{
									Name:     "dir",
									Optional: true,
									Title:    "Directory",
									Type:     sunbeam.ParamString,
								},
							},
						},
						{
							Name:  "preview",
							Title: "Preview Command Output",
							Mode:  sunbeam.CommandModeDetail,
							Params: []sunbeam.Param{
								{
									Name:  "command",
									Title: "Command",
									Type:  sunbeam.ParamString,
								},
							},
						},
					},
				}

				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.SetEscapeHTML(false)

				return encoder.Encode(manifest)
			}

			decoder := json.NewDecoder(strings.NewReader(args[0]))
			var payload sunbeam.Payload
			if err := decoder.Decode(&payload); err != nil {
				return err
			}

			switch payload.Command {
			case "copy":
				var params struct {
					Text string `json:"text"`
				}
				if err := mapstructure.Decode(payload.Params, &params); err != nil {
					return err
				}

				return clipboard.WriteAll(params.Text)
			case "open":
				var params struct {
					Url  string `json:"url"`
					Path string `json:"path"`
				}
				if err := mapstructure.Decode(payload.Params, &params); err != nil {
					return err
				}

				if params.Url != "" && params.Path != "" {
					return fmt.Errorf("only one of url or path is allowed")
				}

				if params.Path != "" {
					return utils.Open(params.Path)
				}

				if params.Url != "" {
					return utils.Open(params.Url)
				}

				return fmt.Errorf("url or path is required")
			case "edit":
				var params struct {
					Path string `json:"path"`
				}
				if err := mapstructure.Decode(payload.Params, &params); err != nil {
					return err
				}

				shell := utils.FindShell()
				editor := utils.FindEditor()
				path := payload.Params["path"].(string)
				cmd := exec.Command(shell, "-c", fmt.Sprintf("%s %s", editor, path))
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				return cmd.Run()
			case "run":
				var params struct {
					Command string `json:"command"`
					Dir     string `json:"dir"`
				}
				if err := mapstructure.Decode(payload.Params, &params); err != nil {
					return err
				}

				shell := utils.FindShell()
				cmd := exec.Command(shell, "-c", params.Command)
				cmd.Dir = params.Dir
				if cmd.Dir == "" {
					cmd.Dir = payload.Cwd
				}
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return cmd.Run()
			default:
				return fmt.Errorf("unknown command: %s", payload.Command)
			}
		},
	}

	return cmd
}
