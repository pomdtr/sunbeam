package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func main() {
	var verboseFlag bool

	rootCmd := &cobra.Command{Use: "myapp"}

	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", true, "Enable verbose mode")

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if verboseFlag {
			fmt.Println("Verbose mode is enabled")
		} else {
			fmt.Println("Verbose mode is disabled")
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
