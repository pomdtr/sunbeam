package cmd

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/mapstructure"
	cp "github.com/otiai10/copy"
	"github.com/pomdtr/sunbeam/types"
	"golang.org/x/term"

	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

const (
	manifestName = "sunbeam.json"
)

var (
	commandRoot                      = filepath.Join(xdg.DataHome, "sunbeam", "commands")
	commands     map[string]Manifest = make(map[string]Manifest)
	commandNames                     = []string{}
)

func init() {
	dataHome := xdg.DataHome
	if runtime.GOOS == "darwin" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		dataHome = filepath.Join(homeDir, ".local", "share")
	}
	commandRoot = filepath.Join(dataHome, "sunbeam", "commands")

	if _, err := os.Stat(commandRoot); err != nil {
		if err := os.MkdirAll(commandRoot, 0755); err != nil {
			panic(err)
		}
		return
	}

	dirs, err := os.ReadDir(commandRoot)
	if err != nil {
		return
	}

	for _, dir := range dirs {
		manifestPath := filepath.Join(commandRoot, dir.Name(), manifestName)
		if _, err := os.Stat(manifestPath); err != nil {
			continue
		}

		f, err := os.Open(manifestPath)
		if err != nil {
			continue
		}

		var cmd Manifest
		if err := json.NewDecoder(f).Decode(&cmd); err != nil {
			continue
		}

		commands[dir.Name()] = cmd
		commandNames = append(commandNames, dir.Name())
	}
}

type Manifest struct {
	Origin  string `json:"origin"`
	Version string `json:"version"`
	Command
}

type Command struct {
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	Entrypoint  string             `json:"entrypoint,omitempty"`
	SubCommands map[string]Command `json:"subcommands,omitempty"`
}

var MetadataRegexp = regexp.MustCompile(`@(sunbeam|raycast)\.(?P<key>[A-Za-z0-9]+)\s(?P<value>[\S ]+)`)

func ExtractScriptMetadata(script []byte) (Command, error) {
	var cmd Command
	matches := MetadataRegexp.FindAllSubmatch(script, -1)
	if len(matches) == 0 {
		return cmd, fmt.Errorf("no metadata found")
	}

	metadata := Command{}
	for _, match := range matches {
		key := string(match[2])
		value := string(match[3])

		switch key {
		case "title":
			metadata.Title = value
		case "description":
			metadata.Description = value
		default:
			log.Printf("unsupported metadata key: %s", key)
		}
	}

	if metadata.Title == "" {
		return cmd, fmt.Errorf("no title found")
	}

	return metadata, nil
}

type CommandRemote interface {
	GetLatestVersion() (string, error)
	Download(targetDir string, version string) error
}

func GetRemote(origin string) (CommandRemote, error) {
	if info, err := os.Stat(origin); err == nil {
		abs, err := filepath.Abs(origin)
		if err != nil {
			return nil, fmt.Errorf("could not get absolute path: %s", err)
		}

		if info.IsDir() {
			return LocalDir{
				path: abs,
			}, nil
		}

		return LocalScript{
			path: abs,
		}, nil
	}

	originUrl, err := url.Parse(origin)
	if err != nil {
		return nil, fmt.Errorf("could not parse origin: %s", err)
	}

	switch originUrl.Hostname() {
	case "github.com":
		return GithubRemote{
			origin: originUrl,
		}, nil
	default:
		return ScriptRemote{
			origin: originUrl,
		}, nil
	}
}

func ParseCommand(commandDir string, prefix string) (Command, error) {
	manifestPath := filepath.Join(commandDir, prefix, manifestName)

	var cmd Command
	f, err := os.Open(manifestPath)
	if err != nil {
		return cmd, fmt.Errorf("unable to open manifest: %s", err)
	}

	var raw struct {
		Title       string         `json:"title"`
		Description string         `json:"description"`
		Entrypoint  string         `json:"entrypoint"`
		SubCommands map[string]any `json:"subcommands"`
	}

	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return cmd, fmt.Errorf("unable to decode manifest: %s", err)
	}

	if raw.Entrypoint == "" && len(raw.SubCommands) == 0 {
		return cmd, fmt.Errorf("no entrypoint or subcommands found")
	}

	cmd.Title = raw.Title
	cmd.Description = raw.Description

	if cmd.Entrypoint != "" {
		cmd.Entrypoint = filepath.Join(prefix, cmd.Entrypoint)
		return cmd, nil
	}

	cmd.SubCommands = make(map[string]Command)
	for name, subcommand := range raw.SubCommands {
		switch subcommand := subcommand.(type) {
		case string:
			cmdPath := filepath.Join(commandDir, prefix, subcommand)
			info, err := os.Stat(cmdPath)
			if err != nil {
				return cmd, fmt.Errorf("unable to find subcommand: %s", err)
			}

			if info.IsDir() {
				return cmd, fmt.Errorf("subcommand cannot be a directory")
			}
			bs, err := os.ReadFile(cmdPath)
			if err != nil {
				return cmd, fmt.Errorf("unable to read subcommand: %s", err)
			}

			subCmd, err := ExtractScriptMetadata(bs)
			if err != nil {
				return cmd, fmt.Errorf("unable to extract metadata: %s", err)
			}
			subCmd.Entrypoint = filepath.Join(prefix, subcommand)

			cmd.SubCommands[name] = subCmd
		case map[string]any:
			var subCmd Command
			if err := mapstructure.Decode(subcommand, &subCmd); err != nil {
				return cmd, fmt.Errorf("unable to decode subcommand: %s", err)
			}

			if len(subCmd.SubCommands) > 0 {
				return cmd, fmt.Errorf("subcommand cannot have subcommands")
			}

			if subCmd.Entrypoint == "" {
				return cmd, fmt.Errorf("subcommand must have entrypoint")
			}

			subCmd.Entrypoint = filepath.Join(prefix, subCmd.Entrypoint)
			cmd.SubCommands[name] = subCmd
		default:
			return cmd, fmt.Errorf("unsupported subcommand type: %T", subcommand)
		}
	}

	return cmd, nil
}

type LocalScript struct {
	path string
}

func (r LocalScript) Download(targetDir string, version string) error {
	if _, err := os.Stat(r.path); err != nil {
		return fmt.Errorf("unable to find script: %s", err)
	}

	bs, err := os.ReadFile(r.path)
	if err != nil {
		return fmt.Errorf("unable to read script: %s", err)
	}

	command, err := ExtractScriptMetadata(bs)
	if err != nil {
		return fmt.Errorf("unable to extract metadata: %s", err)
	}
	command.Entrypoint = filepath.Base(r.path)

	sourceDir := filepath.Join(targetDir, "src")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return fmt.Errorf("unable to create source directory: %s", err)
	}

	entrypoint := filepath.Join(sourceDir, filepath.Base(r.path))
	if err := cp.Copy(r.path, entrypoint); err != nil {
		return fmt.Errorf("unable to copy script: %s", err)
	}

	manifest := Manifest{
		Origin:  r.path,
		Version: version,
		Command: command,
	}

	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal manifest: %s", err)
	}

	manifestPath := filepath.Join(targetDir, manifestName)
	if err := os.WriteFile(manifestPath, content, 0644); err != nil {
		return fmt.Errorf("unable to write manifest: %s", err)
	}

	return nil
}

func (r LocalScript) GetLatestVersion() (string, error) {
	return time.Now().UTC().Format(time.RFC3339), nil
}

type LocalDir struct {
	path string
}

func (r LocalDir) Download(targetDir string, version string) error {
	srcDir := filepath.Join(targetDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("unable to create source directory: %s", err)
	}

	if err := cp.Copy(r.path, srcDir); err != nil {
		return fmt.Errorf("unable to copy script: %s", err)
	}

	command, err := ParseCommand(targetDir, "src")
	if err != nil {
		return fmt.Errorf("unable to parse command: %s", err)
	}

	manifest := Manifest{
		Origin:  r.path,
		Version: version,
		Command: command,
	}

	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal manifest: %s", err)
	}

	manifestPath := filepath.Join(targetDir, manifestName)
	if err := os.WriteFile(manifestPath, content, 0644); err != nil {
		return fmt.Errorf("unable to write manifest: %s", err)
	}

	return nil
}

func (r LocalDir) GetLatestVersion() (string, error) {
	return time.Now().UTC().Format(time.RFC3339), nil
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
	srcDir := filepath.Join(targetDir, "src")
	if err := downloadAndExtractZip(fmt.Sprintf("%s/archive/%s.zip", r.origin.String(), version), srcDir); err != nil {
		return fmt.Errorf("unable to download command: %s", err)
	}

	command, err := ParseCommand(targetDir, "src")
	if err != nil {
		return fmt.Errorf("unable to parse command: %s", err)
	}

	manifest := Manifest{
		Origin:  r.origin.String(),
		Version: version,
		Command: command,
	}

	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal manifest: %s", err)
	}

	manifestPath := filepath.Join(targetDir, manifestName)
	if err := os.WriteFile(manifestPath, content, 0644); err != nil {
		return fmt.Errorf("unable to write manifest: %s", err)
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

func (r ScriptRemote) Download(targetDir string, version string) error {
	res, err := http.Get(r.origin.String())
	if err != nil {
		return fmt.Errorf("unable to download script: %s", err)
	}
	defer res.Body.Close()

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %s", err)
	}

	command, err := ExtractScriptMetadata(bs)
	if err != nil {
		return fmt.Errorf("unable to extract metadata: %s", err)
	}

	entrypointPath := filepath.Join(targetDir, "sunbeam-command")
	if err := os.WriteFile(entrypointPath, bs, 0755); err != nil {
		return fmt.Errorf("unable to write script: %s", err)
	}
	command.Entrypoint = entrypointPath

	manifest := Manifest{
		Origin:  r.origin.String(),
		Version: version,
		Command: command,
	}

	bs, err = json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("unable to encode manifest: %s", err)
	}

	manifestPath := filepath.Join(targetDir, manifestName)
	if err := os.WriteFile(manifestPath, bs, 0644); err != nil {
		return fmt.Errorf("unable to write manifest: %s", err)
	}

	return nil
}

func NewCommandCmd() *cobra.Command {
	commandCmd := &cobra.Command{
		Use:     "command",
		Short:   "Manage, install, and run commands",
		GroupID: coreGroupID,
	}

	commandCmd.AddCommand(NewCommandManageCmd())
	commandCmd.AddCommand(NewCommandAddCmd())
	commandCmd.AddCommand(NewCommandRenameCmd())
	commandCmd.AddCommand(NewCommandListCmd())
	commandCmd.AddCommand(NewCommandRemoveCmd())
	commandCmd.AddCommand(NewCommandUpgradeCmd())

	return commandCmd
}

func NewCommandManageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Manage installed commands",

		RunE: func(cmd *cobra.Command, args []string) error {
			generator := func() (*types.Page, error) {
				listItems := make([]types.ListItem, 0)
				for name, command := range commands {
					listItems = append(listItems, types.ListItem{
						Title:    name,
						Subtitle: command.Description,
						Actions: []types.Action{
							{
								Title: "Run Command",
								Type:  types.PushAction,
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{name},
								},
							},
							{
								Title: "Upgrade Command",
								Type:  types.RunAction,
								Key:   "u",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"command", "upgrade", name},
								},
							},
							{
								Type:  types.RunAction,
								Title: "Rename Command",
								Key:   "r",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"command", "rename", name, "{{input:name}}"},
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
									Args: []string{"command", "remove", name},
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

func NewCommandRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "rename <command> <new-name>",
		Short:     "Rename an command",
		Args:      cobra.ExactArgs(2),
		ValidArgs: commandNames,
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
			oldPath := filepath.Join(commandRoot, args[0])
			newPath := filepath.Join(commandRoot, args[1])

			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("could not rename command: %s", err)
			}

			if err := os.Rename(args[0], args[1]); err != nil {
				return fmt.Errorf("could not rename command: %s", err)
			}

			cmd.Printf("Renamed command %s to %s\n", args[0], args[1])
			return nil
		},
	}
}

func NewCommandAddCmd() *cobra.Command {
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

			remote, err := GetRemote(origin)
			if err != nil {
				return fmt.Errorf("could not get remote: %s", err)
			}

			version, err := remote.GetLatestVersion()
			if err != nil {
				return fmt.Errorf("could not get version: %s", err)
			}

			tempDir, err := os.MkdirTemp("", "sunbeam-install-*")
			if err != nil {
				return fmt.Errorf("could not create temporary directory: %s", err)
			}
			defer os.RemoveAll(tempDir)

			if err := remote.Download(tempDir, version); err != nil {
				return fmt.Errorf("could not install command: %s", err)
			}

			commandDir := filepath.Join(commandRoot, commandName)
			if err := os.Rename(tempDir, commandDir); err != nil {
				return fmt.Errorf("could not install command: %s", err)
			}

			fmt.Printf("✓ Installed command %s\n", commandName)
			return nil
		},
	}

	return cmd
}

func NewCommandListCmd() *cobra.Command {
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
			for commandName, command := range commands {
				printer.AddField(commandName)
				printer.AddField(command.Description)
				printer.EndRow()
			}

			return printer.Render()
		},
	}

	return cmd
}

func NewCommandRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "remove <command> [commands...]",
		Short:     "Remove an installed command",
		Aliases:   []string{"rm", "uninstall"},
		Args:      cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
		ValidArgs: commandNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, commandName := range args {
				commandDir := filepath.Join(commandRoot, commandName)

				if err := os.RemoveAll(commandDir); err != nil {
					return fmt.Errorf("unable to remove command: %s", err)
				}

				fmt.Printf("✓ Removed command %s\n", commandName)
			}
			return nil
		},
	}
}

func NewCommandUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "upgrade [--all] [<command>]",
		Short:     "Upgrade an installed command",
		Aliases:   []string{"up", "update"},
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: commandNames,
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
				toUpgrade = commandNames
			} else {
				toUpgrade = args
			}

			for _, commandName := range toUpgrade {
				command, ok := commands[commandName]
				if !ok {
					return fmt.Errorf("command %s not found", commandName)
				}

				remote, err := GetRemote(command.Origin)
				if err != nil {
					return fmt.Errorf("could not get remote: %s", err)
				}

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
				defer os.RemoveAll(tempdir)

				if err := remote.Download(tempdir, version); err != nil {
					return fmt.Errorf("unable to upgrade command: %s", err)
				}

				commandDir := filepath.Join(commandRoot, commandName)
				if err := os.RemoveAll(commandDir); err != nil {
					return fmt.Errorf("unable to upgrade command: %s", err)
				}

				if err := os.Rename(tempdir, commandDir); err != nil {
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

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE, file.Mode())
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
