package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sunbeamlauncher/sunbeam/server"
)

func NewCmdServe() *cobra.Command {
	command := cobra.Command{
		Use:     "serve",
		GroupID: "core",
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := cmd.Flags().GetString("host")
			if err != nil {
				return err
			}
			port, err := cmd.Flags().GetInt("port")
			if err != nil {
				return err
			}

			server := server.New(fmt.Sprintf("%s:%d", host, port))

			fmt.Println("Listening on", server.Addr)
			return server.ListenAndServe()
		},
	}

	command.Flags().StringP("host", "H", "localhost", "Host to listen on")
	command.Flags().IntP("port", "p", 8080, "Port to listen on")

	return &command
}
