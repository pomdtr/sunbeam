package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read [page]",
		Short: "Read page from file or stdin, and push it's content",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

			var runner *tui.CommandRunner
			if len(args) == 0 {
				if isatty.IsTerminal(os.Stdin.Fd()) {
					exitWithErrorMsg("No input provided")
				}

				bytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Println("An error occured while reading script:", err)
					os.Exit(1)
				}

				cwd, err := os.Getwd()
				if err != nil {
					exitWithErrorMsg("could not get current working directory: %s", err)
				}

				generator := func(string) ([]byte, error) {
					return bytes, nil
				}
				runner = tui.NewCommandRunner(generator, &url.URL{
					Scheme: "file",
					Path:   cwd,
				})

			} else {
				target, err := url.Parse(args[0])
				if err != nil {
					fmt.Fprintln(os.Stderr, "An error occured while parsing the url:", err)
					os.Exit(1)
				}

				switch target.Scheme {
				case "http", "https":
					generator := func(s string) ([]byte, error) {
						res, err := http.Get(args[0])
						if err != nil {
							return nil, err
						}
						defer res.Body.Close()

						if res.StatusCode != http.StatusOK {
							return nil, fmt.Errorf("http request failed with status %s", res.Status)
						}

						return io.ReadAll(res.Body)
					}

					runner = tui.NewCommandRunner(generator, &url.URL{
						Scheme: target.Scheme,
						Host:   target.Host,
						Path:   path.Dir(target.Path),
					})
				case "file", "":
					generator := func(s string) ([]byte, error) {
						return os.ReadFile(args[0])
					}
					runner = tui.NewCommandRunner(generator, &url.URL{
						Scheme: "file",
						Path:   path.Dir(args[0]),
					})

				default:
					fmt.Fprintln(os.Stderr, "Unsupported scheme:", target.Scheme)
					os.Exit(1)
				}
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := runner.Generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Println(string(output))
				return
			}

			model := tui.NewModel(runner, tui.SunbeamOptions{
				Padding:   padding,
				MaxHeight: maxHeight,
			})

			model.Draw()
		},
	}

	return cmd
}
