package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <script>",
		Short: "Run a script",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")
			var generator tui.Generator
			if path.Ext(args[0]) == ".json" {
				generator = func(s string) ([]byte, error) {
					return os.ReadFile(args[0])

				}
			} else if args[0] == "-" {
				bytes, err := ioutil.ReadAll(os.Stdin)
				if err != nil {
					fmt.Println("An error occured while reading script:", err)
					os.Exit(1)
				}

				generator = func(s string) ([]byte, error) {
					if s != "" {
						return nil, fmt.Errorf("stdin script do not support beeing refreshed")
					}
					return bytes, nil
				}
			} else {
				generator = func(s string) ([]byte, error) {
					name, args := utils.SplitCommand(args)
					command := exec.Command(name, args...)

					return command.Output()
				}
			}
			if check, _ := cmd.Flags().GetBool("check"); check {
				page, err := generator("")
				if err != nil {
					fmt.Println("An error occured while reading script:", err)
					os.Exit(1)
				}

				var v interface{}
				if err := json.Unmarshal(page, &v); err != nil {
					fmt.Println("Script is not valid:", err)
					os.Exit(1)
				}

				if err := schemas.Schema.Validate(v); err != nil {
					fmt.Println("Script is not valid:", err)
					os.Exit(1)
				}

				fmt.Println("Script is valid!")
				os.Exit(0)
				return
			}

			runner := tui.NewCommandRunner(generator)
			model := tui.NewModel(runner, tui.SunbeamOptions{
				Padding:   padding,
				MaxHeight: maxHeight,
			})
			model.Draw()
		},
	}

	cmd.Flags().Bool("check", false, "Check the script output format")

	return cmd
}
