package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/spf13/cobra"
)

var scriptTemplate = `#!/bin/sh

if [ $# -eq 1 ] && [ "$1" = "manifest" ] ; then
	exec sunbeam fetch {{ if not (eq .Token "") }} -H 'Authorization: Bearer {{ .Token }}' {{ end }} '{{ .Origin }}'
fi

exec sunbeam fetch {{ if not (eq .Token "") }} -H 'Authorization: Bearer {{ .Token }}' {{ end }} '{{ .Origin }}' -d @-
`

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <origin> [args...]",
		Short:              "Run an extension from a script, directory, or URL",
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveDefault
			}

			if len(args) > 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") || args[0] == "-" {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			entrypoint := args[0]
			if entrypoint == "." {
				entrypoint = "./sunbeam-extension"
			}

			extension, err := tui.LoadExtension(entrypoint)
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			completions := make([]string, 0)
			for _, command := range extension.Commands {
				completions = append(completions, fmt.Sprintf("%s\t%s", command.Name, command.Title))
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "--help" || args[0] == "-h" {
				return cmd.Help()
			}

			var scriptPath string
			if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") {
				origin, err := url.Parse(args[0])
				if err != nil {
					return err
				}

				var token string
				if origin.User != nil {
					if _, ok := origin.User.Password(); !ok {
						token = origin.User.Username()
						origin.User = nil
					}
				}

				template, err := template.New("script").Parse(scriptTemplate)
				if err != nil {
					return err
				}

				tempfile, err := os.CreateTemp("", "sunbeam-run-*.sh")
				if err != nil {
					return err
				}
				defer os.Remove(tempfile.Name())

				if err := template.Execute(tempfile, struct {
					Origin string
					Token  string
				}{Origin: origin.String(), Token: token}); err != nil {
					return err
				}

				if err := os.Chmod(tempfile.Name(), 0755); err != nil {
					return err
				}

				scriptPath = tempfile.Name()
			} else {
				s, err := filepath.Abs(args[0])
				if err != nil {
					return err
				}

				if info, err := os.Stat(s); err != nil {
					return err
				} else if info.IsDir() {
					scriptPath = filepath.Join(s, "sunbeam-extension")
					if _, err := os.Stat(scriptPath); err != nil {
						return fmt.Errorf("no extension found at %s", args[0])
					}
				} else {
					scriptPath = s
				}
			}

			extension, err := tui.LoadExtension(scriptPath)
			if err != nil {
				return err
			}

			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			extensions[args[0]] = extension

			rootCmd, err := NewCmdCustom(extensions, args[0])
			if err != nil {
				return fmt.Errorf("error loading extension: %w", err)
			}

			rootCmd.Use = "extension"
			rootCmd.SetArgs(args[1:])
			return rootCmd.Execute()
		},
	}

	return cmd
}
