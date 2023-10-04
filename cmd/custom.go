package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/google/shlex"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"muzzammil.xyz/jsonc"
)

func NewCmdCustom(use string, extension tui.Extension) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use: use,
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	for _, subcommand := range extension.Commands {
		subcommand := subcommand
		cmd := &cobra.Command{
			Use:    subcommand.Name,
			Short:  subcommand.Title,
			Hidden: subcommand.Hidden,
			RunE: func(cmd *cobra.Command, args []string) error {
				params := make(map[string]any)
				for _, param := range subcommand.Params {
					if !cmd.Flags().Changed(param.Name) {
						continue
					}
					switch param.Type {
					case types.ParamTypeString:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}
						params[param.Name] = value
					case types.ParamTypeBoolean:
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}
						params[param.Name] = value
					}
				}

				if subcommand.Mode == types.CommandModeView {
					if !isatty.IsTerminal(os.Stdout.Fd()) {
						output, err := extension.Run(tui.CommandInput{
							Command: subcommand.Name,
							Params:  params,
						})

						if err != nil {
							return err
						}

						if _, err := os.Stdout.Write(output); err != nil {
							return err
						}
					}
					return tui.Draw(tui.NewRunner(extension, subcommand, params), MaxHeigth)
				}

				out, err := extension.Run(tui.CommandInput{
					Command: subcommand.Name,
					Params:  params,
				})
				if err != nil {
					return err
				}

				if len(out) == 0 {
					return nil
				}

				var command types.Command
				if err := jsonc.Unmarshal(out, &command); err != nil {
					return err
				}

				switch command.Type {
				case types.CommandTypeCopy:
					return clipboard.WriteAll(command.Text)
				case types.CommandTypeOpen:
					return utils.OpenWith(command.Target, command.App)
				default:
					return nil
				}
			},
		}

		for _, param := range subcommand.Params {
			switch param.Type {
			case types.ParamTypeString:
				cmd.Flags().String(param.Name, "", param.Description)
			case types.ParamTypeBoolean:
				cmd.Flags().Bool(param.Name, false, param.Description)
			}

			if !param.Optional {
				_ = cmd.MarkFlagRequired(param.Name)
			}
		}

		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

func buildDoc(command *cobra.Command) (string, error) {
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

func LookupIntEnv(key string, fallback int) int {
	env, ok := os.LookupEnv(key)
	if !ok {
		return fallback

	}

	value, err := strconv.Atoi(env)
	if err != nil {
		return fallback
	}

	return value
}

func LookupBoolEnv(key string, fallback bool) bool {
	env, ok := os.LookupEnv(key)
	if !ok {
		return fallback

	}

	b, err := strconv.ParseBool(env)
	if err != nil {
		return fallback
	}

	return b
}

type CommandRef struct {
	Alias   string
	Command string
	Params  map[string]any
}

func ExtractCommand(shellCommand string) (CommandRef, error) {
	var ref CommandRef
	args, err := shlex.Split(shellCommand)
	if err != nil {
		return ref, err
	}

	if len(args) == 0 {
		return ref, fmt.Errorf("no command specified")
	}
	if args[0] != "sunbeam" {
		return ref, fmt.Errorf("invalid command: %s", args[0])
	}
	args = args[1:]

	if len(args) == 0 {
		return ref, fmt.Errorf("no extension specified")
	}

	ref.Alias = args[0]
	args = args[1:]

	if len(args) == 0 {
		return ref, fmt.Errorf("no command specified")
	}

	ref.Command = args[0]
	args = args[1:]

	if len(args) == 0 {
		return ref, nil
	}

	ref.Params = make(map[string]any)

	for len(args) > 0 {
		if !strings.HasPrefix(args[0], "--") {
			return ref, fmt.Errorf("invalid argument: %s", args[0])
		}

		arg := strings.TrimPrefix(args[0], "--")

		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			ref.Params[parts[0]] = parts[1]
			args = args[1:]
			continue
		}

		if len(args) == 1 {
			ref.Params[arg] = true
			args = args[1:]
			continue
		}

		if strings.HasPrefix(args[1], "--") {
			ref.Params[arg] = true
			args = args[1:]
			continue
		}

		ref.Params[arg] = args[1]
		args = args[2:]
	}

	return ref, nil
}
