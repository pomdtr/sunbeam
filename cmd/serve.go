package cmd

import (
	"log"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/server"
	"github.com/spf13/cobra"
)

func NewCmdServe(api app.Api) *cobra.Command {
	return &cobra.Command{
		Use:     "serve",
		GroupID: "core",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := server.NewServer(api.Extensions, ":8080")

			log.Println("Listening on :8080")
			err := server.ListenAndServe()

			return err
		},
	}
}
