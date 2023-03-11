package cmd

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <page>",
		Short: "read page from stdin or file",
		Run: func(cmd *cobra.Command, args []string) {
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")
			var generator tui.Generator
			var dir string

			if args[0] == "-" {
				bytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Println("An error occured while reading script:", err)
					os.Exit(1)
				}

				generator = func(string) ([]byte, error) {
					return bytes, nil
				}
			} else {
				generator = func(s string) ([]byte, error) {
					return os.ReadFile(args[0])
				}
				dir = path.Dir(args[0])
			}

			if check, _ := cmd.Flags().GetBool("check"); check {
				page, err := generator("")
				if err != nil {
					fmt.Println("An error occured while reading the file:", err)
					os.Exit(1)
				}

				if err := schemas.Validate(page); err != nil {
					fmt.Println("File is not valid:", err)
					os.Exit(1)
				}

				fmt.Println("File is valid!")
				return
			}

			runner := tui.NewCommandRunner(generator, dir)
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
