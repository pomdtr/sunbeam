package cmd

import (
	"log"

	"github.com/pomdtr/sunbeam/server"
	"github.com/spf13/cobra"
)

type ServeFlags struct {
	Host string
	Port int
}

var (
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Run a sunbeam script",
		Args:  cobra.MaximumNArgs(0),
		Run:   sunbeamServe,
	}
	serveFlags ServeFlags
)

func init() {
	serveCmd.Flags().StringVarP(&serveFlags.Host, "host", "H", "localhost", "Host to serve on")
	serveCmd.Flags().IntVarP(&serveFlags.Port, "port", "p", 8080, "Port to serve on")
}

func sunbeamServe(cmd *cobra.Command, args []string) {
	err := server.Serve(serveFlags.Host, serveFlags.Port)
	if err != nil {
		log.Fatalln(err)
	}
	return
}
