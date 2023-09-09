package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

func NewCmdServe(extensionMap ExtensionMap) *cobra.Command {
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
			e.HidePort = true

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

			e.GET("/", func(c echo.Context) error {
				return c.JSON(200, map[string]any{
					"version": Version,
					"date":    Date,
				})
			})

			e.GET("/extensions", func(c echo.Context) error {
				extensions := make(map[string]ExtensionManifest)

				for name, extension := range extensionMap {
					extensions[name] = extension.Manifest
				}

				return c.JSON(200, extensions)
			})

			e.GET("/extensions/:extension", func(c echo.Context) error {
				name := c.Param("extension")
				extension, ok := extensionMap[name]
				if !ok {
					return c.String(404, "Not found")
				}
				return c.JSON(200, extension.Manifest)
			})

			e.POST("/extensions/:extension/:command", func(c echo.Context) error {
				extension := c.Param("extension")
				command := c.Param("command")

				manifest, ok := extensionMap[extension]
				if !ok {
					return echo.NewHTTPError(http.StatusNotFound, "Not found")
				}

				var input CommandInput
				if err := c.Bind(&input); err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to bind input: %s", err.Error()))
				}

				output, err := manifest.Run(command, input)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to run command: %s", err.Error()))
				}

				return c.String(200, string(output))
			})

			cmd.PrintErrf("http server started on http://%s:%d\n", flags.host, flags.port)
			return e.Start(fmt.Sprintf("%s:%d", flags.host, flags.port))
		},
	}

	cmd.Flags().IntVarP(&flags.port, "port", "p", 9999, "Port to listen on")
	cmd.Flags().StringVarP(&flags.host, "host", "H", "localhost", "Host to listen on")
	cmd.Flags().StringVarP(&flags.bearerToken, "token", "t", "", "Bearer token to use for authentication")

	return cmd
}
