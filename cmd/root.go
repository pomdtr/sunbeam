package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/cli/v2/pkg/findsh"
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

const (
	coreGroupID      = "core"
	extensionGroupID = "extension"
)

func NewRootCmd(version string) *cobra.Command {

	dataDir := filepath.Join(xdg.DataHome, "sunbeam")
	extensionRoot := filepath.Join(dataDir, "extensions")

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
			var input string
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				input = string(b)
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
				Name:  commandArgs[0],
				Args:  commandArgs[1:],
				Input: input,
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
	rootCmd.AddCommand(NewExtensionCmd(extensionRoot))
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewReadCmd())
	rootCmd.AddCommand(NewTriggerCmd())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewCmdRun(extensionRoot))
	rootCmd.AddCommand(NewInfoCmd(extensionRoot, version))

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

	extensions, err := ListExtensions(extensionRoot)
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
				return runExtension(filepath.Join(extensionRoot, extension), args)
			},
		})
	}

	return rootCmd
}

func runExtension(extensionDir string, args []string) error {
	extensionBinary := filepath.Join(extensionDir, extensionBinaryName)
	extensionBinaryWin := fmt.Sprintf("%s.exe", extensionBinary)

	var command types.Command
	if runtime.GOOS != "windows" {
		command = types.Command{
			Name: extensionBinary,
			Args: args,
		}
	} else if _, err := os.Stat(extensionBinaryWin); err == nil {
		command = types.Command{
			Name: extensionBinaryWin,
			Args: args,
		}

	} else {
		shExe, err := findsh.Find()
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				return errors.New("the `sh.exe` interpreter is required. Please install Git for Windows and try again")
			}
			return err
		}
		forwardArgs := append([]string{"-c", `command "$@"`, "--", extensionBinary}, args...)

		command = types.Command{
			Name: shExe,
			Args: forwardArgs,
		}
	}

	return Draw(internal.NewCommandGenerator(&command))
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

	cmd := paginator.OutputCmd
	if cmd == nil {
		return nil
	}

	if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
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

func NewInfoCmd(extensionRoot string, version string) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print information about sunbeam",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Draw(func() (*types.Page, error) {
				return types.NewList("Info", []types.ListItem{
					{Title: "Version", Subtitle: version, Actions: []types.Action{
						types.NewCopyAction("Copy", version),
					}},
					{Title: "Extension Root", Subtitle: extensionRoot, Actions: []types.Action{
						types.NewCopyAction("Copy", extensionRoot),
					}},
				}), nil
			})
		},
	}

}
