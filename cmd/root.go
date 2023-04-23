package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	_ "embed"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/shlex"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	cobracompletefig "github.com/withfig/autocomplete-tools/integrations/cobra"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
)

func NewRootCmd(version string) *cobra.Command {

	dataDir := path.Join(xdg.DataHome, "sunbeam")
	extensionDir := path.Join(dataDir, "extensions")

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:           "sunbeam",
		Short:         "Command Line Launcher",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		Args: cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if os.Getenv("NO_COLOR") != "" {
				lipgloss.SetColorProfile(termenv.Ascii)
			} else {
				lipgloss.SetColorProfile(termenv.ANSI)
			}
			os.Setenv("SUNBEAM", "true")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				return Draw(internal.NewStaticGenerator(os.Stdin))
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

			return Draw(internal.NewCommandGenerator(&types.Command{
				Args: commandArgs,
			}))
		},
	}

	rootCmd.AddCommand(NewExtensionCmd(extensionDir))
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewCmdServe())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewTriggerCmd())
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

	_, err := p.Run()
	if err != nil {
		return err
	}
	return nil
}

func buildDoc(command *cobra.Command) (string, error) {
	if command.Hidden {
		return "", nil
	}
	if command.Name() == "help" {
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
