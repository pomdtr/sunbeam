package cmd

import (
	"fmt"
	"log"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/server"
	"github.com/spf13/cobra"
)

func NewCmdServe(extensions []*app.Extension) *cobra.Command {
	cmd := cobra.Command{
		Use:     "serve",
		GroupID: "core",
		RunE: func(cmd *cobra.Command, args []string) error {
			host, _ := cmd.Flags().GetString("host")
			port, _ := cmd.Flags().GetInt("port")

			server := server.NewServer(extensions, fmt.Sprintf("%s:%d", host, port))

			log.Printf("Listening on %s:%d", host, port)
			return server.ListenAndServe()
		},
	}

	cmd.Flags().StringP("host", "H", "localhost", "host to listen on")
	cmd.Flags().IntP("port", "p", 8080, "port to listen on")

	return &cmd
}
