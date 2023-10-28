package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/spf13/cobra"
)

func NewCmdLoop() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "loop",
		GroupID: CommandGroupDev,
		Hidden:  true,
		Short:   "Loop",
		RunE: func(cmd *cobra.Command, args []string) error {
			fifo, err := os.OpenFile(args[0], os.O_WRONLY, os.ModeNamedPipe)
			if err != nil {
				fmt.Printf("Error opening FIFO for writing: %v\n", err)
			}
			defer fifo.Close()

			for {
				extensionMap, err := FindExtensions()
				if err != nil {
					return err
				}

				rootList := tui.NewRootList("Sunbeam", extensionMap)
				if err := tui.Draw(rootList); err != nil {
					return err
				}

				// Write to the FIFO
				if _, err := fifo.WriteString("hide\n"); err != nil {
					fmt.Printf("Error writing to FIFO: %v\n", err)
				}

				// clear the screen
				termenv.DefaultOutput().ClearScreen()

				time.Sleep(500 * time.Millisecond)
			}
		},
	}

	return cmd
}
