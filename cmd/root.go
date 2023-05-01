package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/shlex"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	cobracompletefig "github.com/withfig/autocomplete-tools/integrations/cobra"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
)

const (
	coreGroupID      = "core"
	extensionGroupID = "extension"
)

func NewRootCmd(version string) *cobra.Command {

	dataDir := filepath.Join(xdg.DataHome, "sunbeam")
	extensionDir := filepath.Join(dataDir, "extensions")

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		Version:      version,
		SilenceUsage: true,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		Args: cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if os.Getenv("NO_COLOR") != "" {
				lipgloss.SetColorProfile(termenv.Ascii)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				input, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("could not read input: %s", err)
				}

				var action types.Action
				if err := json.Unmarshal(input, &action); err != nil {
					return fmt.Errorf("could not parse input: %s", err)
				}

				inputsFlag, _ := cmd.Flags().GetStringArray("input")
				if len(inputsFlag) < len(action.Inputs) {
					return fmt.Errorf("not enough inputs provided")
				}

				inputs := make(map[string]string)
				for _, input := range inputsFlag {
					parts := strings.SplitN(input, "=", 2)
					if len(parts) != 2 {
						return fmt.Errorf("invalid input: %s", input)
					}
					inputs[parts[0]] = parts[1]
				}

				query, _ := cmd.Flags().GetString("query")
				return triggerAction(action, inputs, query)
			}

			defaultCommand, ok := os.LookupEnv("SUNBEAM_DEFAULT_CMD")
			if !ok {
				return cmd.Usage()
			}

			commandArgs, err := shlex.Split(defaultCommand)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not parse default command: %s", err)
				return err
			}

			if len(commandArgs) == 0 {
				return cmd.Usage()
			}

			return Draw(internal.NewCommandGenerator(&types.Command{
				Name: commandArgs[0],
				Args: commandArgs[1:],
			}))
		},
	}

	rootCmd.Flags().StringArrayP("input", "i", nil, "input to pass to the action")
	rootCmd.Flags().String("query", "", "query to pass to the action")
	rootCmd.Flags().MarkHidden("input")
	rootCmd.Flags().MarkHidden("query")

	rootCmd.AddGroup(
		&cobra.Group{ID: coreGroupID, Title: "Core Commands"},
		&cobra.Group{ID: extensionGroupID, Title: "Extension Commands"},
	)
	rootCmd.AddCommand(NewExtensionCmd(extensionDir))
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewReadCmd())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewCmdRun(extensionDir))

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

	extensions, err := ListExtensions(extensionDir)
	if err != nil {
		return rootCmd
	}

	for _, extension := range extensions {
		extension := extension
		rootCmd.AddCommand(&cobra.Command{
			Use:                extension,
			DisableFlagParsing: true,
			GroupID:            extensionGroupID,
			RunE: func(cmd *cobra.Command, args []string) error {
				extensionPath := filepath.Join(extensionDir, extension, extensionBinaryName)

				return Draw(internal.NewCommandGenerator(&types.Command{
					Name: extensionPath,
					Args: args,
				}))
			},
		})
	}

	return rootCmd
}

func Draw(generator internal.PageGenerator) error {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
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
	options := internal.SunbeamOptions{
		MaxHeight: utils.LookupInt("SUNBEAM_HEIGHT", 0),
		Padding:   utils.LookupInt("SUNBEAM_PADDING", 0),
	}
	paginator := internal.NewPaginator(runner, options)

	var p *tea.Program
	if options.MaxHeight == 0 {
		p = tea.NewProgram(paginator, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(paginator)
	}

	m, err := p.Run()
	if err != nil {
		return err
	}

	paginator, ok := m.(*internal.Paginator)
	if !ok {
		return fmt.Errorf("could not cast model to paginator")
	}

	if paginator.Output != "" {
		fmt.Print(paginator.Output)
	}

	return nil
}

func buildDoc(command *cobra.Command) (string, error) {
	if command.GroupID == extensionGroupID {
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

func triggerAction(action types.Action, inputs map[string]string, query string) error {
	for name, value := range inputs {
		action = internal.RenderAction(action, fmt.Sprintf("${input:%s}", name), value)
	}

	action = internal.RenderAction(action, "${query}", query)

	switch action.Type {
	case types.PushAction:
		if action.Command != nil {
			return Draw(internal.NewCommandGenerator(action.Command))
		}
		return Draw(internal.NewFileGenerator(action.Page))
	case types.RunAction:
		if _, err := action.Command.Output(); err != nil {
			return fmt.Errorf("command failed: %s", err)
		}

		return nil
	case types.OpenAction:
		err := browser.OpenURL(action.Target)
		if err != nil {
			return fmt.Errorf("unable to open link: %s", err)
		}
		return nil
	case types.CopyAction:
		if err := clipboard.WriteAll(action.Text); err != nil {
			return fmt.Errorf("unable to write to clipboard: %s", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}
