package cmd

import (
	"os"
	"strconv"

	_ "embed"

	"github.com/spf13/cobra"
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
	}

	rootCmd.PersistentFlags().IntP("padding", "p", lookupInt("SUNBEAM_PADDING", 0), "padding around the window")
	rootCmd.PersistentFlags().IntP("height", "H", lookupInt("SUNBEAM_HEIGHT", 0), "maximum height of the window")

	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(NewReadCmd())
	rootCmd.AddCommand(NewQueryCmd())

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
