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
	"github.com/cli/cli/v2/pkg/findsh"
	"github.com/mattn/go-isatty"
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

var options internal.SunbeamOptions

func init() {
	options = internal.SunbeamOptions{
		MaxHeight:  utils.LookupIntEnv("SUNBEAM_HEIGHT", 35),
		MaxWidth:   utils.LookupIntEnv("SUNBEAM_WIDTH", 110),
		FullScreen: utils.LookupBoolEnv("SUNBEAM_FULLSCREEN", true),
		Border:     utils.LookupBoolEnv("SUNBEAM_BORDER", true),
	}
}

func NewCmdRoot(version string) (*cobra.Command, error) {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not get current working dir: %w", err)
			}

			extensionPath := filepath.Join(cwd, extensionBinaryName)
			if _, err := os.Stat(extensionPath); err == nil {
				var input string
				if !isatty.IsTerminal(os.Stdin.Fd()) {
					bs, err := io.ReadAll(os.Stdin)
					if err != nil {
						return err
					}

					input = string(bs)
				}

				return runExtension(extensionPath, args, input)
			}

			manifestPath := filepath.Join(cwd, "sunbeam.json")
			if _, err := os.Stat(manifestPath); err == nil {
				return Run(internal.NewFileGenerator(manifestPath))
			}

			return cmd.Usage()
		},
	}

	extensions, err := ListExtensions(extensionRoot)
	if err != nil {
		return nil, fmt.Errorf("could not list extensions: %w", err)
	}

	rootCmd.AddGroup(
		&cobra.Group{ID: coreGroupID, Title: "Core Commands"},
		&cobra.Group{ID: extensionGroupID, Title: "Extension Commands"},
	)
	rootCmd.AddCommand(NewExtensionCmd(extensionRoot, extensions))
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewArgsCmd())
	rootCmd.AddCommand(NewFetchCmd())
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

	for extension, manifest := range extensions {
		rootCmd.AddCommand(NewExtensionExecCmd(extensionRoot, extension, manifest))
	}

	return rootCmd, nil
}

func NewExtensionExecCmd(extensionRoot string, extensionName string, manifest *ExtensionManifest) *cobra.Command {
	return &cobra.Command{
		Use:                extensionName,
		Short:              manifest.Description,
		DisableFlagParsing: true,
		GroupID:            extensionGroupID,

		RunE: func(cmd *cobra.Command, args []string) error {
			var input string
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				inputBytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				input = string(inputBytes)
			}

			if manifest.Type == ExtentionTypeLocal {
				return runExtension(manifest.Entrypoint, args, input)
			}

			return runExtension(filepath.Join(extensionRoot, extensionName, manifest.Entrypoint), args, input)
		},
	}
}

func runExtension(extensionBin string, args []string, input string) error {
	var command types.Command
	if runtime.GOOS != "windows" {
		if err := os.Chmod(extensionBin, 0755); err != nil {
			return err
		}

		command = types.Command{
			Name: extensionBin,
			Args: args,
		}
		return Run(internal.NewCommandGenerator(&command))
	}

	shExe, err := findsh.Find()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("the `sh.exe` interpreter is required. Please install Git for Windows and try again")
		}
		return err
	}
	forwardArgs := append([]string{"-c", `command "$@"`, "--", extensionBin}, args...)

	command = types.Command{
		Name: shExe,
		Args: forwardArgs,
	}

	return Run(internal.NewCommandGenerator(&command))
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
	return internal.Draw(runner, options)
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
			return Run(func() (*types.Page, error) {
				return &types.Page{
					Title: "Info",
					Type:  types.ListPage,
					Items: []types.ListItem{
						{Title: "Version", Subtitle: version, Actions: []types.Action{
							types.NewCopyAction("Copy", version),
						}},
						{Title: "Extension Root", Subtitle: extensionRoot, Actions: []types.Action{
							types.NewCopyAction("Copy", extensionRoot),
						}},
					}}, nil
			})
		},
	}

}
