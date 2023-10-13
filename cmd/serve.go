package cmd

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
)

func NewCmdServe() *cobra.Command {
	var flags struct {
		port         int
		host         string
		token        string
		withoutToken bool
	}

	cmd := &cobra.Command{
		Use:     "serve <path>",
		Short:   "Serve extensions over HTTP",
		Args:    cobra.ExactArgs(1),
		GroupID: CommandGroupCore,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := chi.NewRouter()
			r.Use(middleware.Logger)
			var token string
			if flags.token != "" {
				token = flags.token
			} else if !flags.withoutToken {
				t, err := generateRandomToken()
				if err != nil {
					return err
				}
				token = t
			}

			if token != "" {
				r.Use(AuthMiddleware(token))
			}

			entrypoint, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			if info, err := os.Stat(entrypoint); err != nil {
				return err
			} else if info.IsDir() {
				entrypoint = filepath.Join(entrypoint, "sunbeam-extension")
				if _, err := os.Stat(entrypoint); err != nil {
					return err
				}
			}

			extension, err := LoadExtension(entrypoint)
			if err != nil {
				return err
			}

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				encoder := json.NewEncoder(w)
				if err := encoder.Encode(extension.Manifest); err != nil {
					http.Error(w, fmt.Sprintf("failed to encode manifest: %s", err.Error()), 500)
					return
				}
			})

			for _, command := range extension.Commands {
				command := command
				r.Post(fmt.Sprintf("/%s", command.Name), func(w http.ResponseWriter, r *http.Request) {
					var input types.CommandInput
					if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
						http.Error(w, fmt.Sprintf("failed to decode input: %s", err.Error()), 400)
						return
					}

					output, err := extension.Run(command.Name, input)
					if err != nil {
						http.Error(w, fmt.Sprintf("failed to run command: %s", err.Error()), 500)
						return
					}

					if _, err := w.Write(output); err != nil {
						http.Error(w, fmt.Sprintf("failed to write output: %s", err.Error()), 500)
						return
					}
				})
			}

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
				log.Info("Listening on", "url", fmt.Sprintf("http://%s:%d?token=%s", flags.host, flags.port, token))
			} else {
				log.Info("Listening on", "url", fmt.Sprintf("http://%s:%d", flags.host, flags.port))
			}

			if err := server.ListenAndServe(); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&flags.port, "port", "p", 9999, "Port to listen on")
	cmd.Flags().StringVarP(&flags.host, "host", "H", "localhost", "Host to listen on")
	cmd.Flags().BoolVar(&flags.withoutToken, "without-token", false, "Disable token authentication")
	cmd.Flags().StringVar(&flags.token, "token", "", "Bearer token to use for authentication")

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

func AuthMiddleware(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get token from query string
			queryToken := r.URL.Query().Get("token")
			if queryToken != "" {
				if queryToken != token {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			// get token from authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				if authHeader != fmt.Sprintf("Bearer %s", token) {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "missing token", http.StatusUnauthorized)
		})
	}
}
