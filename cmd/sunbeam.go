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

//go:embed sunbeam.json
var defaultConfig []byte

func Execute(version string) error {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

You will need to provide a compatible script as the first argument to you use sunbeam. See http://pomdtr.github.io/sunbeam for more information.`,
		Args:    cobra.NoArgs,
		Version: version,
		// If the config file does not exist, create it
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			homedir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("Could not find home directory: %w", err)
			}

			configDir := path.Join(homedir, ".config", "sunbeam")
			sunbeamConfig := path.Join(configDir, "sunbeam.json")

			// If the config file exists, do nothing
			if _, err := os.Stat(sunbeamConfig); err == nil {
				return nil
			}

			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("Could not create config directory: %w", err)
			}

			if err := os.WriteFile(sunbeamConfig, defaultConfig, 0644); err != nil {
				return fmt.Errorf("Could not write default config: %w", err)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

			if len(args) == 0 {
				manifest, err := findSunbeamManifest()
				if err != nil {
					fmt.Fprintf(os.Stderr, err.Error())
				}

				runner := tui.NewCommandRunner(func(s string) ([]byte, error) {
					bytes, err := os.ReadFile(manifest)
					if err != nil {
						return nil, fmt.Errorf("Could not read manifest: %w", err)
					}

					return bytes, nil
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
	rootCmd.AddCommand(NewReadCmd())

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
