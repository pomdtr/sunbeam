package cmd

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
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
	// as base64
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authorization := r.Header.Get("Authorization"); authorization == fmt.Sprint("Bearer ", token) {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

func NewCmdServe() *cobra.Command {
	flags := struct {
		port         int
		host         string
		bearerToken  string
		withoutToken bool
	}{}

	cmd := &cobra.Command{
		Use:   "serve <script>",
		Short: "Serve extensions over HTTP",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := chi.NewRouter()
			r.Use(middleware.Logger)
			var token string
			if flags.bearerToken != "" {
				token = flags.bearerToken
			} else if !flags.withoutToken {
				t, err := generateRandomToken()
				if err != nil {
					return err
				}
				token = t
			}

			if token != "" {
				r.Use(BearerMiddleware(token))
			}

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				extension, err := tui.LoadExtension(args[0])
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to load extension: %s", err.Error()), 500)
					return
				}

				encoder := json.NewEncoder(w)
				if err := encoder.Encode(extension.Manifest); err != nil {
					http.Error(w, fmt.Sprintf("failed to encode manifest: %s", err.Error()), 500)
					return
				}
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

			if token != "" {
				log.Printf("sunbeam command: sunbeam run http://%s@%s:%d\n", token, flags.host, flags.port)
				log.Printf("curl command: curl -H 'Authorization: Bearer %s' http://%s:%d\n", token, flags.host, flags.port)
			} else {
				log.Printf("sunbeam command: sunbeam run http://%s:%d\n", flags.host, flags.port)
				log.Printf("curl command: curl http://%s:%d\n", flags.host, flags.port)
			}

			if err := server.ListenAndServe(); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&flags.port, "port", "p", 9999, "Port to listen on")
	cmd.Flags().StringVarP(&flags.host, "host", "H", "localhost", "Host to listen on")
	cmd.Flags().Bool("without-token", false, "Disable bearer token authentication")
	cmd.Flags().String("token", "", "Bearer token to use for authentication")

	return cmd
}

const (
	alphanumericChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tokenLength       = 16
)

func generateRandomToken() (string, error) {
	var tokenRunes []rune
	runesCount := big.NewInt(int64(len(alphanumericChars)))

	for i := 0; i < tokenLength; i++ {
		randomIndex, err := rand.Int(rand.Reader, runesCount)
		if err != nil {
			return "", err
		}
		tokenRunes = append(tokenRunes, []rune(alphanumericChars)[randomIndex.Int64()])
	}

	return string(tokenRunes), nil
}
