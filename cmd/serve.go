package cmd

import (
	"fmt"
	"log"

	"github.com/pomdtr/sunbeam/web"
	"github.com/spf13/cobra"
)

func NewCmdServe() *cobra.Command {
	serveCmd := &cobra.Command{
		Use: "serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme, err := cmd.Flags().GetString("theme")
			if err != nil {
				return err
			}

			host, err := cmd.Flags().GetString("host")
			if err != nil {
				return err
			}
			port, err := cmd.Flags().GetInt("port")
			if err != nil {
				return err
			}

			addr := fmt.Sprintf("%s:%d", host, port)
			server, err := web.NewServer(addr, theme)
			if err != nil {
				return err
			}

			log.Printf("Listening on http://%s", addr)
			return server.ListenAndServe()
		},
	}

	serveCmd.Flags().StringP("host", "H", "localhost", "Host to listen on")
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	serveCmd.Flags().String("theme", "Tomorrow Night", "Theme to use for the frontend")

	return serveCmd
}
