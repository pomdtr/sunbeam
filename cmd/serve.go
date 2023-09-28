package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/spf13/cobra"
)

func BearerMiddleware(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if auth != fmt.Sprintf("Bearer %s", token) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func NewCmdServe() *cobra.Command {
	flags := struct {
		port        int
		host        string
		bearerToken string
	}{}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve extensions over HTTP",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := chi.NewRouter()
			r.Use(middleware.Logger)
			if flags.bearerToken != "" {
				r.Use(BearerMiddleware(flags.bearerToken))
			}

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				extension, err := tui.LoadExtension(args[0])
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to load extension: %s", err.Error()), 500)
					return
				}

				encoder := json.NewEncoder(w)
				encoder.Encode(extension.Manifest)
			})

			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				extension, err := tui.LoadExtension(args[0])
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to load extension: %s", err.Error()), 500)
					return
				}

				var input tui.CommandInput
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					http.Error(w, fmt.Sprintf("failed to decode input: %s", err.Error()), 400)
					return
				}

				output, err := extension.Run(input)
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to run command: %s", err.Error()), 500)
					return
				}

				if _, err := w.Write(output); err != nil {
					http.Error(w, fmt.Sprintf("failed to write output: %s", err.Error()), 500)
					return
				}
			})

			server := &http.Server{
				Addr:    fmt.Sprintf("%s:%d", flags.host, flags.port),
				Handler: r,
			}

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigs
				log.Print("Shutting down...")
				if err := server.Shutdown(context.Background()); err != nil {
					os.Exit(1)
				}

				os.Exit(0)
			}()

			log.Printf("Listening on http://%s:%d", flags.host, flags.port)
			server.ListenAndServe()
			return nil
		},
	}

	cmd.Flags().IntVarP(&flags.port, "port", "p", 9999, "Port to listen on")
	cmd.Flags().StringVarP(&flags.host, "host", "H", "localhost", "Host to listen on")
	cmd.Flags().StringVarP(&flags.bearerToken, "token", "t", "", "Bearer token to use for authentication")

	return cmd
}
