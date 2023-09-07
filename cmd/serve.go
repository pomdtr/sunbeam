package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

type Extension struct {
	Origin   string
	Manifest ExtensionManifest
}

type ExtensionManifest struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Origin      string     `json:"origin,omitempty"`
	Entrypoint  Entrypoint `json:"entrypoint,omitempty"`
	Commands    []Command  `json:"commands"`
}

func LoadManifest(manifestPath string) (ExtensionManifest, error) {
	var manifest ExtensionManifest
	f, err := os.Open(manifestPath)
	if err != nil {
		return manifest, err
	}

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&manifest); err != nil {
		return manifest, err
	}

	return manifest, nil
}

type Entrypoint []string

func (e *Entrypoint) UnmarshalJSON(b []byte) error {
	var entrypoint string
	if err := json.Unmarshal(b, &entrypoint); err == nil {
		*e = Entrypoint{entrypoint}
		return nil
	}

	var entrypoints []string
	if err := json.Unmarshal(b, &entrypoints); err == nil {
		*e = Entrypoint(entrypoints)
		return nil
	}

	return fmt.Errorf("invalid entrypoint: %s", string(b))
}

type Command struct {
	Name        string            `json:"name"`
	Title       string            `json:"title"`
	Mode        CommandMode       `json:"mode"`
	Hidden      bool              `json:"hidden,omitempty"`
	Description string            `json:"description,omitempty"`
	Arguments   []CommandArgument `json:"arguments,omitempty"`
}

type CommandMode string

const (
	CommandModeList      CommandMode = "filter"
	CommandModeGenerator CommandMode = "generator"
	CommandModeDetail    CommandMode = "detail"
	CommandModeText      CommandMode = "text"
	CommandModeSilent    CommandMode = "silent"
)

type CommandArgument struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     any    `json:"default,omitempty"`
	Optional    bool   `json:"optional"`
	Description string `json:"description"`
}

type ArgumentType string

const (
	ArgumentTypeString ArgumentType = "string"
	ArgumentTypeBool   ArgumentType = "bool"
)

type CommandInput struct {
	Command string         `json:"command"`
	Query   string         `json:"query"`
	Params  map[string]any `json:"params"`
}

func (ext Extension) Run(input CommandInput) ([]byte, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(ext.Origin, "application/json", bytes.NewReader(inputBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("command failed: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func NewCmdServe() *cobra.Command {
	flags := struct {
		port        int
		host        string
		bearerToken string
	}{}

	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "Serve extensions over HTTP",
		GroupID: coreGroupID,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}

			extensionRoot := args[0]
			if info, err := os.Stat(extensionRoot); os.IsNotExist(err) {
				return fmt.Errorf("extension root %s does not exist", extensionRoot)
			} else if err != nil {
				return err
			} else if !info.IsDir() {
				return fmt.Errorf("extension root %s is not a directory", extensionRoot)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			e := echo.New()
			e.HideBanner = true

			if flags.bearerToken != "" {
				e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
					return func(c echo.Context) error {
						token := c.Request().Header.Get("Authorization")
						if token != fmt.Sprintf("Bearer %s", flags.bearerToken) {
							return c.String(401, "Unauthorized")
						}

						return next(c)
					}
				})
			}

			extensions, err := LoadExtensions()
			if err != nil {
				return err
			}

			e.GET("/", func(c echo.Context) error {
				return c.JSON(200, extensions)
			})

			e.GET("/:name", func(c echo.Context) error {
				name := c.Param("name")
				manifest, ok := extensions[name]
				if !ok {
					return c.String(404, "Not found")
				}
				return c.JSON(200, manifest)
			})

			e.POST("/:name", func(c echo.Context) error {
				name := c.Param("name")
				manifest, ok := extensions[name]
				if !ok {
					return echo.NewHTTPError(http.StatusNotFound, "Not found")
				}

				var input CommandInput
				if err := c.Bind(&input); err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to bind input: %s", err.Error()))
				}

				output, err := manifest.Run(input)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to run command: %s", err.Error()))
				}

				return c.String(200, string(output))
			})

			return e.Start(fmt.Sprintf("%s:%d", flags.host, flags.port))
		},
	}

	cmd.Flags().IntVarP(&flags.port, "port", "p", 9999, "Port to listen on")
	cmd.Flags().StringVarP(&flags.host, "host", "H", "localhost", "Host to listen on")
	cmd.Flags().StringVarP(&flags.bearerToken, "token", "t", "", "Bearer token to use for authentication")

	return cmd
}

func LoadExtensions() (map[string]Extension, error) {
	var extensions map[string]Extension
	f, err := os.Open(filepath.Join(dataDir, "extensions.json"))
	if err != nil {
		return extensions, err
	}

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&extensions); err != nil {
		return extensions, err
	}

	return extensions, nil
}
