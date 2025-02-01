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

	type ExtensionItem struct {
		Name string `json:"name"`
		Path string `json:"path"`
		sunbeam.Manifest
	}

	cmd := &cobra.Command{
		Use: "serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			http.HandleFunc("GET /extensions", func(w http.ResponseWriter, r *http.Request) {
				exts, err := LoadExtensions(utils.ExtensionsDir(), true)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				var refs []ExtensionItem
				for _, ext := range exts {
					refs = append(refs, ExtensionItem{
						Name:     ext.Name,
						Path:     fmt.Sprintf("/extensions/%s", ext.Name),
						Manifest: ext.Manifest,
					})
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				encoder := json.NewEncoder(w)
				encoder.SetEscapeHTML(false)
				_ = encoder.Encode(refs)
			})

			http.HandleFunc("GET /extensions/{extension}", func(w http.ResponseWriter, r *http.Request) {
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
				_ = encoder.Encode(ExtensionItem{
					Name:     extension.Name,
					Path:     fmt.Sprintf("/extensions/%s", extension.Name),
					Manifest: extension.Manifest,
				})
			})

			http.HandleFunc("POST /extensions/{extension}/{command}", func(w http.ResponseWriter, r *http.Request) {
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
