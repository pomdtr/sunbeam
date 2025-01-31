package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

func NewCmdServe() *cobra.Command {
	var flags struct {
		addr string
	}
	cmd := &cobra.Command{
		Use: "serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
				exts, err := LoadExtensions(utils.ExtensionsDir(), true)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				list := sunbeam.List{}
				for _, extension := range exts {
					list.Items = append(list.Items, extension.RootItems()...)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				encoder := json.NewEncoder(w)
				encoder.SetEscapeHTML(false)

				_ = encoder.Encode(list)
			})

			http.HandleFunc("GET /{extension}", func(w http.ResponseWriter, r *http.Request) {
				entrypoint, err := extensions.FindEntrypoint(utils.ExtensionsDir(), r.PathValue("extension"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}

				extension, err := extensions.LoadExtension(entrypoint, true)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				encoder := json.NewEncoder(w)
				encoder.SetEscapeHTML(false)
				_ = encoder.Encode(extension.Manifest)
			})

			http.HandleFunc("POST /{extension}/{command}", func(w http.ResponseWriter, r *http.Request) {
				entrypoint, err := extensions.FindEntrypoint(utils.ExtensionsDir(), r.PathValue("extension"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}

				extension, err := extensions.LoadExtension(entrypoint, true)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				command, ok := extension.GetCommand(r.PathValue("command"))
				if !ok {
					http.Error(w, "command not found", http.StatusNotFound)
					return
				}

				cmd, err := extension.CmdContext(r.Context(), command, nil)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				cmd.Stdout = w
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				if err := cmd.Run(); err != nil {
					return
				}

			})

			fmt.Fprintln(cmd.OutOrStdout(), "Listening on", flags.addr)
			return http.ListenAndServe(flags.addr, nil)
		},
	}

	cmd.Flags().StringVar(&flags.addr, "addr", ":8080", "address to listen on")

	return cmd
}
