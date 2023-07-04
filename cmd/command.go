package cmd

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/catalog"
	"github.com/pomdtr/sunbeam/types"
	"golang.org/x/term"

	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

var (
	commandBinaryName = "sunbeam-command"
)

type Manifest struct {
	path        string             `json:"-"`
	commandRoot string             `json:"-"`
	Commands    map[string]Command `json:"commands"`
}

type Command struct {
	*catalog.CommandMetadata
	Origin     string `json:"origin"`
	Version    string `json:"version"`
	Dir        string `json:"dir"`
	Entrypoint string `json:"entryPoint"`
}

func (c Command) IsLocal() bool {
	return c.Dir == c.Origin
}

func LoadManifest(manifestPath string) (*Manifest, error) {
	manifest := Manifest{
		path:        manifestPath,
		commandRoot: filepath.Join(filepath.Dir(manifestPath), "commands"),
	}

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		manifest.Commands = make(map[string]Command)

		if err := os.MkdirAll(manifest.commandRoot, 0755); err != nil {
			return nil, err
		}

		return &manifest, manifest.Save()
	}

	manifestFile, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(manifestFile, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (m Manifest) Save() error {
	manifestFile, err := os.OpenFile(m.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer manifestFile.Close()

	encoder := json.NewEncoder(manifestFile)
	encoder.SetIndent("", "  ")
	return encoder.Encode(m)
}

func (m Manifest) AddCommand(commandName string, command Command) error {
	if _, ok := m.Commands[commandName]; ok {
		return fmt.Errorf("command already exists")
	}

	m.Commands[commandName] = command
	return m.Save()
}

func (m Manifest) RemoveCommand(commandName string) error {
	if _, ok := m.Commands[commandName]; !ok {
		return fmt.Errorf("command does not exist")
	}

	delete(m.Commands, commandName)
	return m.Save()
}

func (m Manifest) UpdateCommand(commandName string, command Command) error {
	if _, ok := m.Commands[commandName]; !ok {
		return fmt.Errorf("command does not exist")
	}

	m.Commands[commandName] = command
	return m.Save()
}

func (m Manifest) RenameCommand(oldCommandName string, newCommandName string) error {
	if _, ok := m.Commands[oldCommandName]; !ok {
		return fmt.Errorf("command does not exist")
	}

	if _, ok := m.Commands[newCommandName]; ok {
		return fmt.Errorf("command already exists")
	}

	command := m.Commands[oldCommandName]
	command.Dir = filepath.Join(m.commandRoot, newCommandName)

	m.Commands[newCommandName] = command
	delete(m.Commands, oldCommandName)
	return m.Save()
}

func (m Manifest) ListCommands() []string {
	var commands []string
	for command := range m.Commands {
		commands = append(commands, command)
	}
	return commands
}

type CommandRemote interface {
	GetLatestVersion() (string, error)
	Download(targetDir string, version string) error
}

func GetRemote(origin *url.URL) CommandRemote {
	switch origin.Hostname() {
	case "github.com":
		return GithubRemote{
			origin: origin,
		}
	case "gist.github.com":
		return GistRemote{
			origin: origin,
		}
	default:
		return ScriptRemote{
			origin: origin,
		}
	}
}

type GithubRemote struct {
	origin *url.URL
}

func (r GithubRemote) GetLatestVersion() (string, error) {
	commit, err := utils.GetLastGitCommit(r.origin.String())
	if err != nil {
		return "", err
	}

	return commit.Sha, nil
}

func (r GithubRemote) Download(targetDir string, version string) error {
	if err := downloadAndExtractZip(fmt.Sprintf("%s/archive/%s.zip", r.origin.String(), version), filepath.Join(targetDir, "src")); err != nil {
		return fmt.Errorf("unable to download command: %s", err)
	}

	return nil
}

type GistRemote struct {
	origin *url.URL
}

func (r GistRemote) GetLatestVersion() (string, error) {
	commit, err := utils.GetLastGistCommit(r.origin.String())
	if err != nil {
		return "", err
	}

	return commit.Version, nil
}

func (r GistRemote) Download(targetDir string, version string) error {
	fmt.Fprintf(os.Stderr, "Installing command from gist...\n")
	zipUrl := fmt.Sprintf("%s/archive/%s.zip", r.origin.String(), version)
	if err := downloadAndExtractZip(zipUrl, filepath.Join(targetDir, "src")); err != nil {
		return fmt.Errorf("unable to install command: %s", err)
	}

	return nil
}

type ScriptRemote struct {
	origin *url.URL
}

func (r ScriptRemote) GetLatestVersion() (string, error) {
	// return utc timestamp
	return time.Now().UTC().Format(time.RFC3339), nil
}

func (r ScriptRemote) Download(targetDir string, _ string) error {
	res, err := http.Get(r.origin.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	entrypointPath := filepath.Join(targetDir, commandBinaryName)
	f, err := os.OpenFile(entrypointPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, res.Body); err != nil {
		return err
	}

	return nil
}

func NewCommandCmd(manifest *Manifest) *cobra.Command {
	commandCmd := &cobra.Command{
		Use:     "command",
		Short:   "Command commands",
		GroupID: coreGroupID,
	}

	commandCmd.AddCommand(NewCommandBrowseCmd())
	commandCmd.AddCommand(NewCommandManageCmd(manifest))
	commandCmd.AddCommand(NewCommandCreateCmd())
	commandCmd.AddCommand(NewCommandAddCmd(manifest))
	commandCmd.AddCommand(NewCommandRenameCmd(manifest))
	commandCmd.AddCommand(NewCommandListCmd(manifest))
	commandCmd.AddCommand(NewCommandRemoveCmd(manifest))
	commandCmd.AddCommand(NewCommandUpgradeCmd(manifest))

	return commandCmd
}

func NewCommandBrowseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "browse",
		Short: "Browse commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			generator := func() (*types.Page, error) {
				catalogItems, err := catalog.FetchCatalog()
				if err != nil {
					return nil, err
				}

				listItems := make([]types.ListItem, 0)
				for _, item := range catalogItems {
					listItems = append(listItems, types.ListItem{
						Title:    item.Metadata.Title,
						Subtitle: item.Metadata.Title,
						Accessories: []string{
							item.Metadata.Author,
						},
						Actions: []types.Action{
							{
								Type:  types.RunAction,
								Title: "Install",
								Key:   "i",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"command", "install", "{{input:name}}", item.Origin},
								},
								Inputs: []types.Input{
									{
										Name:        "name",
										Type:        types.TextFieldInput,
										Title:       "Name",
										Placeholder: "my-command-name",
									},
								},
							},
							{
								Type:   types.OpenAction,
								Title:  "Open in Browser",
								Target: item.Origin,
								Key:    "o",
							},
						},
					})
				}

				return &types.Page{
					Type:  types.ListPage,
					Items: listItems,
				}, nil
			}

			return Run(generator)
		},
	}
}

func NewCommandManageCmd(manifest *Manifest) *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Manage installed commands",

		RunE: func(cmd *cobra.Command, args []string) error {
			generator := func() (*types.Page, error) {
				listItems := make([]types.ListItem, 0)
				for command, manifest := range manifest.Commands {
					listItems = append(listItems, types.ListItem{
						Title:    command,
						Subtitle: manifest.Description,
						Actions: []types.Action{
							{
								Title: "Run Command",
								Type:  types.PushAction,
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{command},
								},
							},
							{
								Title: "Upgrade Command",
								Type:  types.RunAction,
								Key:   "u",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"command", "upgrade", command},
								},
							},
							{
								Type:  types.RunAction,
								Title: "Rename Command",
								Key:   "r",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"command", "rename", command, "{{input:name}}"},
								},
								Inputs: []types.Input{
									{
										Name:        "name",
										Type:        types.TextFieldInput,
										Default:     command,
										Title:       "Name",
										Placeholder: "my-command-name",
									},
								},
							},
							{
								Type:  types.RunAction,
								Title: "Remove Command",
								Key:   "d",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"command", "remove", command},
								},
							},
						},
					})
				}

				return &types.Page{
					Type: types.ListPage,
					EmptyView: &types.EmptyView{
						Text: "No commands installed",
						Actions: []types.Action{
							{
								Type:  types.PushAction,
								Title: "Browse Commands",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"command", "browse"},
								},
							},
						},
					},
					Items: listItems,
				}, nil
			}

			return Run(generator)
		},
	}
}

//go:embed sunbeam-command
var commandTemplate string

func NewCommandCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new command",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("name is required")
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("unable to get current working directory: %s", err)
			}

			commandDir := filepath.Join(cwd, name)
			if _, err := os.Stat(commandDir); err == nil {
				return fmt.Errorf("directory %s already exists", commandDir)
			}

			if err := os.Mkdir(commandDir, 0755); err != nil {
				return fmt.Errorf("unable to create directory %s: %s", commandDir, err)
			}

			if err := os.WriteFile(filepath.Join(commandDir, commandBinaryName), []byte(commandTemplate), 0755); err != nil {
				return fmt.Errorf("unable to write command binary: %s", err)
			}

			cmd.Printf("Created command %s\n", name)
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "command name")
	cmd.MarkFlagRequired("name") //nolint:errcheck

	return cmd
}

func NewCommandRenameCmd(manifest *Manifest) *cobra.Command {
	return &cobra.Command{
		Use:       "rename <command> <new-name>",
		Short:     "Rename an command",
		Args:      cobra.ExactArgs(2),
		ValidArgs: manifest.ListCommands(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			commands := cmd.Root().Commands()
			commandMap := make(map[string]struct{}, len(commands))
			for _, command := range commands {
				commandMap[command.Name()] = struct{}{}
			}

			if _, ok := commandMap[args[0]]; !ok {
				return fmt.Errorf("command does not exist: %s", args[0])
			}

			if _, ok := commandMap[args[1]]; ok {
				return fmt.Errorf("name conflicts with existing command: %s", args[0])
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			oldPath := filepath.Join(manifest.commandRoot, args[0])
			newPath := filepath.Join(manifest.commandRoot, args[1])

			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("could not rename command: %s", err)
			}

			if err := manifest.RenameCommand(args[0], args[1]); err != nil {
				return fmt.Errorf("could not rename command: %s", err)
			}

			cmd.Printf("Renamed command %s to %s\n", args[0], args[1])
			return nil
		},
	}
}

func NewCommandAddCmd(manifest *Manifest) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name> <origin>",
		Short:   "Install a sunbeam command from a folder/gist/repository",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"install"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			commandName := args[0]
			commands := cmd.Root().Commands()

			commandMap := make(map[string]struct{}, len(commands))
			for _, command := range commands {
				commandMap[command.Name()] = struct{}{}
			}

			if _, ok := commandMap[args[0]]; ok {
				return fmt.Errorf("name conflicts with existing command: %s", commandName)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			commandName := args[0]
			origin := args[1]

			if info, err := os.Stat(origin); err == nil {
				entrypoint, err := filepath.Abs(origin)
				if err != nil {
					return fmt.Errorf("could not get absolute path: %s", err)
				}

				if info.IsDir() {
					entrypoint = filepath.Join(entrypoint, commandBinaryName)
				}

				content, err := os.ReadFile(entrypoint)
				if err != nil {
					return fmt.Errorf("unable to read command binary: %s", err)
				}

				metadata, err := catalog.ExtractCommandMetadata(content)
				if err != nil {
					return fmt.Errorf("unable to extract metadata: %s", err)
				}

				if err := manifest.AddCommand(commandName, Command{
					Version:         "local",
					Origin:          origin,
					Dir:             filepath.Dir(entrypoint),
					Entrypoint:      filepath.Base(entrypoint),
					CommandMetadata: metadata,
				}); err != nil {
					return fmt.Errorf("could not add command: %s", err)
				}

				cmd.Printf("Added command %s\n", commandName)
				return nil

			}

			originUrl, err := url.Parse(origin)
			if err != nil {
				return fmt.Errorf("could not parse origin: %s", err)
			}

			remote := GetRemote(originUrl)

			version, err := remote.GetLatestVersion()
			if err != nil {
				return fmt.Errorf("could not get version: %s", err)
			}

			tempDir, err := os.MkdirTemp("", "sunbeam-install-*")
			if err != nil {
				return fmt.Errorf("could not create temporary directory: %s", err)
			}
			if err := remote.Download(tempDir, version); err != nil {
				return fmt.Errorf("could not install command: %s", err)
			}

			content, err := os.ReadFile(filepath.Join(tempDir, commandBinaryName))
			if err != nil {
				return fmt.Errorf("unable to read command binary: %s", err)
			}

			metadata, err := catalog.ExtractCommandMetadata(content)
			if err != nil {
				return fmt.Errorf("unable to extract metadata: %s", err)
			}

			if err := os.MkdirAll(manifest.commandRoot, 0755); err != nil {
				return fmt.Errorf("could not create command directory: %s", err)
			}

			commandDir := filepath.Join(manifest.commandRoot, commandName)

			if err := os.Rename(tempDir, commandDir); err != nil {
				return fmt.Errorf("could not install command: %s", err)
			}

			if err := manifest.AddCommand(commandName, Command{
				Origin:          origin,
				Entrypoint:      commandBinaryName,
				Dir:             commandDir,
				Version:         version,
				CommandMetadata: metadata,
			}); err != nil {
				return fmt.Errorf("could not add command: %s", err)
			}

			fmt.Printf("✓ Installed command %s\n", commandName)
			return nil
		},
	}

	return cmd
}

func NewCommandListCmd(manifest *Manifest) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed command commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			var isTTY bool
			var width int
			if isatty.IsTerminal(os.Stdout.Fd()) {
				isTTY = true
				w, _, err := term.GetSize(int(os.Stdout.Fd()))
				if err != nil {
					return err
				}
				width = w
			}

			printer := tableprinter.New(os.Stdout, isTTY, width)
			for commandName, command := range manifest.Commands {
				printer.AddField(commandName)
				printer.AddField(command.Description)
				printer.EndRow()
			}

			return printer.Render()
		},
	}

	return cmd
}

func NewCommandRemoveCmd(manifest *Manifest) *cobra.Command {
	return &cobra.Command{
		Use:       "remove <command> [commands...]",
		Short:     "Remove an installed command",
		Aliases:   []string{"rm", "uninstall"},
		Args:      cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
		ValidArgs: manifest.ListCommands(),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, commandName := range args {
				command, ok := manifest.Commands[commandName]
				if !ok {
					return fmt.Errorf("command %s not installed", commandName)
				}

				if command.IsLocal() {
					if err := manifest.RemoveCommand(commandName); err != nil {
						return fmt.Errorf("unable to remove command: %s", err)
					}

					fmt.Printf("✓ Removed command %s\n", commandName)
					continue
				}

				if err := os.RemoveAll(command.Dir); err != nil {
					return fmt.Errorf("unable to remove command: %s", err)
				}

				if err := manifest.RemoveCommand(commandName); err != nil {
					return fmt.Errorf("unable to remove command: %s", err)
				}

				fmt.Printf("✓ Removed command %s\n", commandName)
			}
			return nil
		},
	}
}

func NewCommandUpgradeCmd(manifest *Manifest) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "upgrade [--all] [<command>]",
		Short:     "Upgrade an installed command",
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: manifest.ListCommands(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			if all && len(args) > 0 {
				return fmt.Errorf("cannot specify both --all and an command name")
			}

			if !all && len(args) == 0 {
				return fmt.Errorf("must specify either --all or an command name")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var toUpgrade []string
			all, _ := cmd.Flags().GetBool("all")
			if all {
				toUpgrade = manifest.ListCommands()
			} else {
				toUpgrade = args
			}

			for _, commandName := range toUpgrade {
				command := manifest.Commands[commandName]
				if command.IsLocal() {
					fmt.Printf("Command %s is not installed from a remote, skipping\n", commandName)
					continue
				}

				originUrl, err := url.Parse(command.Origin)
				if err != nil {
					return fmt.Errorf("could not parse origin: %s", err)
				}

				remote := GetRemote(originUrl)

				version, err := remote.GetLatestVersion()
				if err != nil {
					return fmt.Errorf("could not get version: %s", err)
				}

				if version == command.Version {
					fmt.Printf("Command %s already up to date\n", commandName)
					continue
				}

				tempdir, err := os.MkdirTemp("", "sunbeam-*")
				if err != nil {
					return fmt.Errorf("unable to upgrade command: %s", err)
				}

				if err := remote.Download(tempdir, version); err != nil {
					return fmt.Errorf("unable to upgrade command: %s", err)
				}

				if err := os.RemoveAll(command.Dir); err != nil {
					return fmt.Errorf("unable to upgrade command: %s", err)
				}

				if err := os.Rename(tempdir, command.Dir); err != nil {
					return fmt.Errorf("unable to upgrade command: %s", err)
				}

				command.Version = version
				if err := manifest.UpdateCommand(commandName, command); err != nil {
					return fmt.Errorf("unable to upgrade command: %s", err)
				}

				fmt.Printf("✓ Upgraded command %s\n", commandName)
			}

			return nil
		},
	}

	cmd.Flags().BoolP("all", "a", false, "upgrade all commands")

	return cmd
}

func downloadAndExtractZip(zipUrl string, dst string) error {
	res, err := http.Get(zipUrl)
	if err != nil {
		return fmt.Errorf("unable to download command: %s", err)
	}
	defer res.Body.Close()

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to download command: %s", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(bs), int64(len(bs)))
	if err != nil {
		return fmt.Errorf("unable to download command: %s", err)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	rootDir := zipReader.File[0].Name
	for _, file := range zipReader.File[1:] {
		fpath := filepath.Join(dst, strings.TrimPrefix(file.Name, rootDir))

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, file.Mode()); err != nil {
				return err
			}
			continue
		}

		mode := file.Mode()
		if filepath.Base(fpath) == "sunbeam-command" {
			mode = 0755
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE, mode)
		if err != nil {
			return err
		}

		// Copy the file contents
		rc, err := file.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}

		outFile.Close()
		rc.Close()
	}

	return nil
}
