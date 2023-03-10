package cmd

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/tui"
)

func Execute(version string) error {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

You will need to provide a compatible script as the first argument to you use sunbeam. See http://pomdtr.github.io/sunbeam for more information.`,
		Args:    cobra.NoArgs,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

			if len(args) == 0 {
				manifest, err := findSunbeamManifest()
				if err != nil {
					fmt.Fprintf(os.Stderr, err.Error())
				}

				runner := tui.NewCommandRunner(func(s string) ([]byte, error) {
					return os.ReadFile(manifest)
				})
				model := tui.NewModel(runner, tui.SunbeamOptions{
					Padding:   padding,
					MaxHeight: maxHeight,
				})
				model.Draw()
				return
			}
		},
	}

	rootCmd.PersistentFlags().IntP("padding", "p", lookupInt("SUNBEAM_PADDING", 0), "padding around the window")
	rootCmd.PersistentFlags().IntP("height", "H", lookupInt("SUNBEAM_HEIGHT", 0), "maximum height of the window")

	rootCmd.AddCommand(NewRunCmd())

	return rootCmd.Execute()
}

func findSunbeamManifest() (string, error) {
	cwd, _ := os.Getwd()
	homedir, _ := os.UserHomeDir()
	candidates := []string{
		path.Join(cwd, "sunbeam.json"),
		path.Join(homedir, ".config", "sunbeam", "sunbeam.json"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("Could not find sunbeam.json in current or config directory")
}

func lookupInt(key string, fallback int) int {
	if env, ok := os.LookupEnv(key); ok {
		if value, err := strconv.Atoi(env); err == nil {
			return value
		}
	}

	return fallback
}
