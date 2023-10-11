package cmd

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
)

func NewCmdServe() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve",
		GroupID: CommandGroupCore,
	}

	cmd.AddCommand(
		NewCmdServeHTTP(),
		NewCmdServeSSH(),
	)

	return cmd
}

func NewCmdServeSSH() *cobra.Command {
	var flags struct {
		host string
		port int
	}

	cmd := &cobra.Command{
		Use:  "ssh [src]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var middleware func(s ssh.Session) (tea.Model, []tea.ProgramOption)
			if len(args) == 0 {
				generator := func() (map[string]extensions.Extension, []types.ListItem, error) {
					extensions, err := FindExtensions()
					if err != nil {
						return nil, nil, err
					}

					items := make([]types.ListItem, 0)
					for alias, extension := range extensions {
						for _, command := range extension.Commands {
							if !IsRootCommand(command) {
								continue
							}
							items = append(items, types.ListItem{
								Title:    command.Title,
								Subtitle: extension.Title,
								Actions: []types.Action{
									{
										Title: "Run",
										OnAction: types.Command{
											Type:      types.CommandTypeRun,
											Extension: alias,
											Command:   command.Name,
										},
									},
								},
							})
						}
					}

					return extensions, items, nil
				}
				middleware = func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
					rootList := tui.NewRootList("Sunbeam", generator)
					return tui.NewPaginator(rootList), []tea.ProgramOption{tea.WithAltScreen()}
				}
			} else {
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

				extension, err := extensions.Load(entrypoint)
				if err != nil {
					return err
				}

				extensionMap, err := FindExtensions()
				if err != nil {
					return err
				}
				extensionMap[args[0]] = extension

				rootCommands := make([]types.CommandSpec, 0)
				for _, command := range extension.Commands {
					if !IsRootCommand(command) {
						continue
					}

					rootCommands = append(rootCommands, command)
				}

				if len(rootCommands) == 0 {
					return fmt.Errorf("no root commands found in extension %s", args[0])
				}

				if len(rootCommands) == 1 {
					command := rootCommands[0]

					middleware = func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
						runner, _ := tui.NewRunner(extensionMap, tui.CommandRef{
							Extension: args[0],
							Command:   command.Name,
						})
						return tui.NewPaginator(runner), []tea.ProgramOption{tea.WithAltScreen()}
					}
				} else {
					generator := func() (map[string]extensions.Extension, []types.ListItem, error) {
						items := make([]types.ListItem, 0)
						for _, command := range rootCommands {
							items = append(items, types.ListItem{
								Title:    command.Title,
								Subtitle: extension.Title,
								Actions: []types.Action{
									{
										Title: "Run Command",
										OnAction: types.Command{
											Type:      types.CommandTypeRun,
											Extension: args[0],
											Command:   command.Name,
										},
									},
								},
							})
						}
						return extensionMap, items, nil
					}

					middleware = func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
						page := tui.NewRootList(extension.Title, generator)
						return tui.NewPaginator(page), []tea.ProgramOption{tea.WithAltScreen()}
					}
				}
			}

			s, err := wish.NewServer(
				wish.WithAddress(fmt.Sprintf("%s:%d", flags.host, flags.port)),
				wish.WithHostKeyPath(".ssh/term_info_ed25519"),
				wish.WithMiddleware(
					bm.Middleware(middleware),
					lm.Middleware(),
				),
			)
			if err != nil {
				return err
			}

			done := make(chan os.Signal, 1)
			signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			log.Info("Starting SSH server", "host", flags.host, "port", flags.port)
			go func() {
				if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
					log.Error("could not start server", "error", err)
					done <- nil
				}
			}()

			<-done
			log.Info("Stopping SSH server")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer func() { cancel() }()
			if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
				log.Error("could not stop server", "error", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.host, "host", "H", "localhost", "Host to listen on")
	cmd.Flags().IntVarP(&flags.port, "port", "p", 9999, "Port to listen on")

	return cmd
}

func NewCmdServeHTTP() *cobra.Command {
	var flags struct {
		port         int
		host         string
		token        string
		withoutToken bool
	}

	cmd := &cobra.Command{
		Use:   "http [src]",
		Short: "Serve extensions over HTTP",
		Args:  cobra.MaximumNArgs(1),
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

			if len(args) > 0 {
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

				extension, err := extensions.Load(entrypoint)
				if err != nil {
					return err
				}

				registerExtension(r, extension)
			} else {
				extensionMap, err := FindExtensions()
				if err != nil {
					return err
				}
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					endpoints := make(map[string]string)
					for alias := range extensionMap {
						endpoint, err := url.JoinPath(r.URL.String(), alias)
						if err != nil {
							http.Error(w, fmt.Sprintf("failed to join path: %s", err.Error()), 500)
							return
						}
						endpoints[alias] = endpoint
					}

					encoder := json.NewEncoder(w)
					if err := encoder.Encode(endpoints); err != nil {
						http.Error(w, fmt.Sprintf("failed to encode endpoints: %s", err.Error()), 500)
						return
					}
				})

				for alias, extension := range extensionMap {
					group := chi.NewRouter()
					registerExtension(group, extension)
					r.Mount(fmt.Sprintf("/%s", alias), group)
				}
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

func registerExtension(r chi.Router, extension extensions.Extension) {
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
