package cmd

import (
	"fmt"
	"os"
	"path"
	"strconv"

	_ "embed"

	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/tui"
)

func Execute(version string) error {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		Args:    cobra.NoArgs,
		Version: version,
		// If the config file does not exist, create it
		Run: func(cmd *cobra.Command, args []string) {
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

			cwd, _ := os.Getwd()
			manifestPath := path.Join(cwd, "sunbeam.json")

			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				cmd.Help()
				os.Exit(0)
			}

			runner := tui.NewCommandRunner(func(s string) ([]byte, error) {
				bytes, err := os.ReadFile(manifestPath)
				if err != nil {
					return nil, fmt.Errorf("Could not read manifest: %w", err)
				}

				return bytes, nil
			}, path.Dir(manifestPath))
			model := tui.NewModel(runner, tui.SunbeamOptions{
				Padding:   padding,
				MaxHeight: maxHeight,
			})
			model.Draw()
		},
	}

	rootCmd.PersistentFlags().IntP("padding", "p", lookupInt("SUNBEAM_PADDING", 0), "padding around the window")
	rootCmd.PersistentFlags().IntP("height", "H", lookupInt("SUNBEAM_HEIGHT", 0), "maximum height of the window")
	rootCmd.Flags().StringP("directory", "C", "", "cd to dir before starting sunbeam")

	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(NewPushCmd())
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewCopyCmd())
	rootCmd.AddCommand(NewOpenCmd())

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
