package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"

	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

type triggerPayload struct {
	Action types.Action   `json:"action"`
	Inputs map[string]any `json:"inputs"`
}

func NewCmdServe() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a web server to serve sunbeam",
		RunE: func(cmd *cobra.Command, args []string) error {
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				command := exec.Command("sunbeam")
				output, err := command.Output()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				w.Write(output)
			})

			http.HandleFunc("/trigger", func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				var payload triggerPayload
				if err := json.Unmarshal(body, &payload); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				b, err := json.Marshal(payload.Action)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				args := []string{"trigger"}
				for name, value := range payload.Inputs {
					args = append(args, fmt.Sprintf("--input=%s=%s", name, value))
				}

				command := exec.Command("sunbeam", args...)
				command.Stdin = bytes.NewReader(b)

				output, err := command.Output()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				w.Write(output)
			})

			port, _ := cmd.Flags().GetInt("port")
			log.Printf("Listening on http://localhost:%d", port)
			return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		},
	}

	cmd.Flags().IntP("port", "p", 8080, "port to listen on")

	return cmd
}
