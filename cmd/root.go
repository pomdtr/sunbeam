package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	cobracompletefig "github.com/withfig/autocomplete-tools/integrations/cobra"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
)

const (
	coreGroupID   = "core"
	customGroupID = "custom"
)

var (
	Version = "dev"
	Date    = "unknown"
)

var (
	options internal.SunbeamOptions
)

func init() {
	options = internal.SunbeamOptions{
		MaxHeight:  utils.LookupIntEnv("SUNBEAM_HEIGHT", 0),
		MaxWidth:   utils.LookupIntEnv("SUNBEAM_WIDTH", 0),
		FullScreen: utils.LookupBoolEnv("SUNBEAM_FULLSCREEN", true),
		Border:     utils.LookupBoolEnv("SUNBEAM_BORDER", false),
		Margin:     utils.LookupIntEnv("SUNBEAM_MARGIN", 0),
	}
}

func Execute() error {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		Version:      fmt.Sprintf("%s (%s)", Version, Date),
		SilenceUsage: true,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
	}

	rootCmd.AddGroup(
		&cobra.Group{ID: coreGroupID, Title: "Core Commands"},
		&cobra.Group{ID: customGroupID, Title: "Custom Commands"},
	)

	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewReadCmd())
	rootCmd.AddCommand(NewExtensionCmd())
	rootCmd.AddCommand(NewFetchCmd())
	rootCmd.AddCommand(NewParseCmd())
	rootCmd.AddCommand(NewListCmd())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewDetailCmd())
	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(cobracompletefig.CreateCompletionSpecCommand())

	docCmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation for sunbeam",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := buildDoc(rootCmd)
			if err != nil {
				return err
			}

			fmt.Println(doc)
			return nil
		},
	}
	rootCmd.AddCommand(docCmd)

	for name, command := range extensions {
		customCmd, err := NewCustomCmd(name, command)
		if err != nil {
			return err
		}

		rootCmd.AddCommand(customCmd)
	}

	manCmd := &cobra.Command{
		Use:    "generate-man-pages [path]",
		Short:  "Generate Man Pages for sunbeam",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			header := &doc.GenManHeader{
				Title:   "MINE",
				Section: "3",
			}
			err := doc.GenManTree(rootCmd, header, args[0])
			if err != nil {
				return err
			}

			return nil
		},
	}
	rootCmd.AddCommand(manCmd)

	return rootCmd.Execute()

}

func Run(generator internal.PageGenerator) error {
	if !isatty.IsTerminal(os.Stderr.Fd()) {
		output, err := generator()
		if err != nil {
			return fmt.Errorf("could not generate page: %s", err)
		}

		if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
			return fmt.Errorf("could not encode page: %s", err)
		}

		return nil
	}

	runner := internal.NewRunner(generator)
	err := internal.Draw(runner, options)
	if errors.Is(err, internal.ErrInterrupted) && isatty.IsTerminal(os.Stdout.Fd()) {
		return nil
	}

	return err
}

func buildDoc(command *cobra.Command) (string, error) {
	if command.GroupID == customGroupID {
		return "", nil
	}

	var page strings.Builder
	err := doc.GenMarkdown(command, &page)
	if err != nil {
		return "", err
	}

	out := strings.Builder{}
	for _, line := range strings.Split(page.String(), "\n") {
		if strings.Contains(line, "SEE ALSO") {
			break
		}

		out.WriteString(line + "\n")
	}

	for _, child := range command.Commands() {
		childPage, err := buildDoc(child)
		if err != nil {
			return "", err
		}
		out.WriteString(childPage)
	}

	return out.String(), nil
}

func NewCustomCmd(name string, extension Extension) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:                name,
		Short:              extension.Title,
		Long:               extension.Description,
		DisableFlagParsing: true,
		Args:               cobra.NoArgs,
		GroupID:            customGroupID,
	}

	for _, command := range extension.Commands {
		subcmd := &cobra.Command{
			Use:                name,
			Short:              extension.Title,
			Long:               extension.Description,
			DisableFlagParsing: true,
			Args:               cobra.NoArgs,
			GroupID:            customGroupID,
			RunE: func(cmd *cobra.Command, _ []string) error {
				var args []string
				entrypoint := command.Entrypoint
				if command.Entrypoint == "" {
					entrypoint = extension.Entrypoint
					args = append(args, command.Name)
				}

				argumentMap := make(map[string]any)
				for _, argument := range command.Arguments {
					switch argument.Type {
					case "string":
						value, err := cmd.Flags().GetString(argument.Name)
						if err != nil {
							return err
						}

						argumentMap[argument.Name] = value
					case "bool":
						value, err := cmd.Flags().GetBool(argument.Name)
						if err != nil {
							return err
						}

						argumentMap[argument.Name] = value
					default:
						return fmt.Errorf("unsupported argument type: %s", argument.Type)
					}
				}

				input, err := json.Marshal(argumentMap)
				if err != nil {
					return err
				}

				return Run(internal.NewCommandGenerator(&types.Command{
					Name:  filepath.Join(extension.RootDir, entrypoint),
					Args:  args,
					Input: string(input),
				}))
			},
		}

		for _, argument := range command.Arguments {
			switch argument.Type {
			case "string":
				cmd.Flags().String(argument.Name, "", argument.Description)
			case "bool":
				cmd.Flags().Bool(argument.Name, false, argument.Description)
			default:
				return nil, fmt.Errorf("unsupported argument type: %s", argument.Type)
			}
		}

		cmd.AddCommand(subcmd)
	}

	return cmd, nil
}
