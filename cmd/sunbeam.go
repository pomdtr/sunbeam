package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/scripts"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/pomdtr/sunbeam/utils"
)

func Execute(version string) error {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam <script>",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

You will need to provide a compatible script as the first argument to you use sunbeam. See http://pomdtr.github.io/sunbeam for more information.`,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			if check, _ := cmd.Flags().GetBool("check"); check {
				var page []byte

				if len(args) == 0 {
					cwd, _ := os.Getwd()
					sunbeamFile := path.Join(cwd, "sunbeam.json")

					if _, err := os.Stat(sunbeamFile); os.IsNotExist(err) {
						fmt.Println("No script provided and no sunbeam.json file found in the current directory.")
						os.Exit(1)
					}

					var err error
					if page, err = os.ReadFile(sunbeamFile); err != nil {
						fmt.Println("An error occured while reading the script:", err)
						os.Exit(1)
					}
				} else {
					name, args := utils.SplitCommand(args)
					cmd := exec.Command(name, args...)
					var err error
					if page, err = cmd.Output(); err != nil {
						fmt.Println("An error occured while running the script:", err)
						os.Exit(1)
					}
				}

				var v interface{}
				if err := json.Unmarshal(page, &v); err != nil {
					fmt.Println("Script is not valid:", err)
					os.Exit(1)
				}

				if err := scripts.Schema.Validate(v); err != nil {
					fmt.Println("Script is not valid:", err)
					os.Exit(1)
				}

				fmt.Println("Script is valid!")
				os.Exit(0)
			}

			if len(args) == 0 {
				cwd, _ := os.Getwd()
				sunbeamFile := path.Join(cwd, "sunbeam.json")

				if _, err := os.Stat(sunbeamFile); os.IsNotExist(err) {
					fmt.Println("No script provided and no sunbeam.json file found in the current directory.")
					os.Exit(1)
				}

				runner := tui.NewCommandRunner(func(s string) ([]byte, error) {
					return os.ReadFile(sunbeamFile)
				})
				model := tui.NewModel(runner, tui.SunbeamOptions{
					Padding:   0,
					MaxHeight: 0,
				})
				model.Draw()
				return
			}

			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

			var generator tui.Generator
			if args[0] == "-" {
				bytes, err := ioutil.ReadAll(os.Stdin)
				if err != nil {
					fmt.Println("An error occured while reading stdin:", err)
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
					cmd := exec.Command(name, args...)
					cmd.Stdin = strings.NewReader(s)
					return cmd.Output()
				}
			}
			runner := tui.NewCommandRunner(generator)
			model := tui.NewModel(runner, tui.SunbeamOptions{
				Padding:   padding,
				MaxHeight: maxHeight,
			})
			model.Draw()
		},
	}

	rootCmd.Flags().IntP("padding", "p", lookupInt("SUNBEAM_PADDING", 0), "padding around the window")
	rootCmd.Flags().IntP("height", "H", lookupInt("SUNBEAM_HEIGHT", 0), "maximum height of the window")
	rootCmd.Flags().Bool("check", false, "check if the script is valid")

	return rootCmd.Execute()
}

func lookupInt(key string, fallback int) int {
	if env, ok := os.LookupEnv(key); ok {
		if value, err := strconv.Atoi(env); err == nil {
			return value
		}
	}

	return fallback
}
