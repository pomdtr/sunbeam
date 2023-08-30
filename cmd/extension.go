package cmd

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/types"
	"golang.org/x/term"

	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

var (
	extensions Extensions
)

func init() {
	m, err := LoadExtensions()
	if err != nil {
		panic(err)
	}

	extensions = m
}

type Extensions map[string]Extension

func (e Extensions) Names() []string {
	var names []string
	for name := range e {
		names = append(names, name)
	}
	return names
}

func LoadExtensions() (Extensions, error) {
	f, err := os.Open(filepath.Join(xdg.DataHome, "sunbeam", "manifest.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return Extensions{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var metadatas map[string]Metadata
	if err := json.NewDecoder(f).Decode(&extensions); err != nil {
		return nil, err
	}

	for name, metadata := range metadatas {
		manifestPath := filepath.Join(metadata.RootDir, "sunbeam.json")
		f, err := os.Open(manifestPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		var command Extension
		if err := json.NewDecoder(f).Decode(&command); err != nil {
			return nil, err
		}

		command.Metadata = metadata
		extensions[name] = command
	}

	return extensions, nil
}

func (c Extensions) Save() error {
	metadatas := make(map[string]Metadata)
	for name, command := range c {
		metadatas[name] = command.Metadata
	}
	f, err := os.OpenFile(filepath.Join(xdg.DataHome, "sunbeam", "manifest.json"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(metadatas)
}

type CommandMode string

const (
	CommandModeView   CommandMode = "view"
	CommandModeNoView CommandMode = "no-view"
)

type Metadata struct {
	Origin  string `json:"origin"`
	RootDir string `json:"root_dir"`
	Version string `json:"version"`
}

type Manifest struct {
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Entrypoint  string    `json:"entrypoint,omitempty"`
	Commands    []Command `json:"subcommands,omitempty"`
}

type Command struct {
	Name        string      `json:"name"`
	Hidden      bool        `json:"hidden,omitempty"`
	Description string      `json:"description"`
	Entrypoint  string      `json:"entrypoint"`
	Output      CommandMode `json:"output,omitempty"`
	Arguments   []Argument  `json:"arguments,omitempty"`
}

type Extension struct {
	Manifest
	Metadata
}

type ArgumentType string

const (
	ArgumentTypeString ArgumentType = "string"
	ArgumentTypeBool   ArgumentType = "bool"
)

type Argument struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Optional    bool   `json:"required"`
}

func LoadManifest(rootDir string) (Manifest, error) {
	var manifest Manifest
	f, err := os.Open(filepath.Join(rootDir, filepath.Join(rootDir, "sunbeam.json")))
	if err != nil {
		return manifest, err
	}
	defer f.Close()

	manifest.Entrypoint = filepath.Join(rootDir, manifest.Entrypoint)
	return manifest, json.NewDecoder(f).Decode(&manifest)
}

func (m Manifest) Save(targetDir string) error {
	f, err := os.OpenFile(filepath.Join(targetDir, "sunbeam.json"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(m)
}

type CommandRemote interface {
	GetLatestVersion() (string, error)
	Download(targetDir string, version string) (Extension, error)
}

var gistRegexp = regexp.MustCompile(`^https?://gist\.github\.com/([A-Za-z0-9_-]+)`)
var repoRegexp = regexp.MustCompile(`^https?://github\.com/([A-Za-z0-9_-]+)/([A-Za-z0-9_-]+)$`)

func GetRemote(origin *url.URL) (CommandRemote, error) {
	origin.Host = strings.TrimPrefix(origin.Host, "www.")

	switch origin.Host {
	case "github.com":
		matches := repoRegexp.FindStringSubmatch(origin.String())

		if len(matches) != 3 {
			return nil, fmt.Errorf("invalid repo url: %s", origin)
		}

		return &GithubRemote{
			Owner: matches[1],
			Name:  matches[2],
		}, nil
	case "gist.github.com":
		matches := gistRegexp.FindStringSubmatch(origin.String())

		if len(matches) != 2 {
			return nil, fmt.Errorf("invalid gist url: %s", origin.String())
		}

		return GistRemote{
			GistID: matches[1],
		}, nil

	default:
		return HttpRemote{
			origin: origin.String(),
		}, nil

	}

}

type LocalRemote struct {
	path string
}

func (r LocalRemote) GetLatestVersion() (string, error) {
	bs, err := os.ReadFile(r.path)
	if err != nil {
		return "", fmt.Errorf("unable to read command: %s", err)
	}

	hash := sha256.Sum256(bs)

	// return utc timestamp

	return hex.EncodeToString(hash[:])[:8], nil
}

func (r LocalRemote) Download(targetDir string, version string) (Extension, error) {
	manifest, err := LoadManifest(r.path)
	if err != nil {
		return Extension{}, err
	}

	return Extension{
		Manifest: manifest,
		Metadata: Metadata{
			Origin:  r.path,
			RootDir: targetDir,
			Version: version,
		},
	}, nil
}

type GithubRemote struct {
	Owner string
	Name  string
}

func (r GithubRemote) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

func (r GithubRemote) Url() string {
	return fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Name)
}

func (r GithubRemote) GetLatestVersion() (string, error) {
	commit, err := utils.GetLastGitCommit(r.Owner, r.Name)
	if err != nil {
		return "", err
	}

	return commit.Sha, nil
}

func (r GithubRemote) Download(targetDir string, version string) (Extension, error) {
	if version == "" {
		v, err := r.GetLatestVersion()
		if err != nil {
			return Extension{}, fmt.Errorf("unable to get latest version: %s", err)
		}
		version = v
	}

	if err := downloadAndExtractZip(fmt.Sprintf("%s/archive/%s.zip", r.Url(), version), targetDir); err != nil {
		return Extension{}, fmt.Errorf("unable to download command: %s", err)
	}

	manifest, err := LoadManifest(targetDir)
	if err != nil {
		return Extension{}, fmt.Errorf("unable to load manifest: %s", err)
	}

	return Extension{
		Manifest: manifest,
		Metadata: Metadata{
			Origin:  fmt.Sprintf("npm:%s", r.FullName()),
			RootDir: targetDir,
			Version: version,
		},
	}, nil
}

type GistRemote struct {
	GistID string
}

func (r GistRemote) GetLatestVersion() (string, error) {
	commit, err := utils.GetLastGistCommit(r.GistID)
	if err != nil {
		return "", err
	}

	return commit.Version, nil
}

func (r GistRemote) Download(targetDir string, version string) (Extension, error) {
	if err := downloadAndExtractZip(fmt.Sprintf("https://gist.github.com/%s/archive/%s.zip", r.GistID, version), targetDir); err != nil {
		return Extension{}, err
	}

	manifest, err := LoadManifest(targetDir)
	if err != nil {
		return Extension{}, err
	}

	return Extension{
		Manifest: manifest,
		Metadata: Metadata{
			Origin:  fmt.Sprintf("https://gist.github.com/%s", r.GistID),
			RootDir: targetDir,
			Version: version,
		},
	}, nil
}

type HttpRemote struct {
	origin string
}

func (r HttpRemote) GetLatestVersion() (string, error) {
	resp, err := http.Get(r.origin)
	if err != nil {
		return "", fmt.Errorf("unable to fetch latest version: %s", err)
	}
	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to fetch latest version: %s", err)
	}

	hash := sha256.Sum256(bs)
	return hex.EncodeToString(hash[:])[:8], nil
}

func (r HttpRemote) Download(targetDir string, version string) (Extension, error) {
	resp, err := http.Get(r.origin)
	if err != nil {
		return Extension{}, fmt.Errorf("unable to fetch latest version: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Extension{}, fmt.Errorf("unable to fetch latest version: %s", resp.Status)
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return Extension{}, fmt.Errorf("unable to fetch latest version: %s", err)
	}

	if err := manifest.Save(targetDir); err != nil {
		return Extension{}, err
	}

	return Extension{
		Manifest: manifest,
		Metadata: Metadata{
			Origin:  r.origin,
			RootDir: targetDir,
			Version: version,
		},
	}, nil
}

func (r HttpRemote) Origin() string {
	return r.origin
}

func NewExtensionCmd() *cobra.Command {
	commandCmd := &cobra.Command{
		Use:     "extension",
		Short:   "Manage, install, and run extensions",
		GroupID: coreGroupID,
	}

	commandCmd.AddCommand(NewExtensionManageCmd())
	commandCmd.AddCommand(NewExtensionAddCmd())
	commandCmd.AddCommand(NewExtensionRenameCmd())
	commandCmd.AddCommand(NewExtensionListCmd())
	commandCmd.AddCommand(NewExtensionRemoveCmd())
	commandCmd.AddCommand(NewExtensionUpgradeCmd())

	return commandCmd
}

func NewExtensionManageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Manage installed extensions",

		RunE: func(cmd *cobra.Command, args []string) error {
			generator := func() (*types.Page, error) {
				listItems := make([]types.ListItem, 0)
				for name, command := range extensions {
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

func NewExtensionRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "rename <command> <new-name>",
		Short:     "Rename an command",
		Args:      cobra.ExactArgs(2),
		ValidArgs: extensions.Names(),
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
			command := extensions[args[0]]
			extensions[args[1]] = command
			delete(extensions, args[0])

			if err := extensions.Save(); err != nil {
				return err
			}

			cmd.Printf("Renamed command %s to %s\n", args[0], args[1])
			return nil
		},
	}
}

func NewExtensionAddCmd() *cobra.Command {
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
			if _, err := os.Stat(origin); err == nil {
				manifest, err := LoadManifest(origin)
				if err != nil {
					return err
				}
				extensions[commandName] = Extension{
					Manifest: manifest,
					Metadata: Metadata{
						Origin:  args[1],
						RootDir: args[1],
					},
				}

				return nil
			}

			originUrl, err := originToUrl(origin)
			if err != nil {
				return fmt.Errorf("could not parse origin: %s", err)
			}

			remote, err := GetRemote(originUrl)
			if err != nil {
				return fmt.Errorf("could not get remote: %s", err)
			}

			version, err := remote.GetLatestVersion()
			if err != nil {
				return fmt.Errorf("could not get version: %s", err)
			}

			targetDir := filepath.Join(xdg.CacheHome, "sunbeam", utils.UrlToFilename(originUrl), version)
			command, err := remote.Download(targetDir, version)
			if err != nil {
				return fmt.Errorf("could not install command: %s", err)
			}

			extensions[commandName] = command
			if err := extensions.Save(); err != nil {
				return fmt.Errorf("could not save manifest: %s", err)
			}

			fmt.Printf("✓ Installed command %s\n", commandName)
			return nil
		},
	}

	return cmd
}

func originToUrl(origin string) (*url.URL, error) {
	if strings.HasPrefix(origin, "github:") {
		return url.Parse(fmt.Sprintf("https://github.com/%s", strings.TrimPrefix(origin, "github:")))
	}

	if strings.HasPrefix(origin, "gist:") {
		return url.Parse(fmt.Sprintf("https://gist.github.com/%s", strings.TrimPrefix(origin, "gist:")))

	}

	return url.Parse(origin)
}

func NewExtensionListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed command commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			var printer tableprinter.TablePrinter
			if isatty.IsTerminal(os.Stdout.Fd()) {
				width, _, err := term.GetSize(int(os.Stdout.Fd()))
				if err != nil {
					return err
				}
				printer = tableprinter.New(os.Stdout, true, width)
			} else {
				printer = tableprinter.New(os.Stdout, false, 0)
			}

			for commandName, command := range extensions {
				printer.AddField(commandName)
				printer.AddField(command.Title)
				printer.EndRow()
			}

			return printer.Render()
		},
	}

	return cmd
}

func NewExtensionRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "remove <command> [commands...]",
		Short:     "Remove an installed command",
		Aliases:   []string{"rm", "uninstall"},
		Args:      cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
		ValidArgs: extensions.Names(),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, commandName := range args {
				command := extensions[commandName]

				if err := os.RemoveAll(command.Metadata.RootDir); err != nil {
					return fmt.Errorf("unable to remove command: %s", err)
				}

				delete(extensions, commandName)
				extensions.Save()

				fmt.Printf("✓ Removed command %s\n", commandName)
			}
			return nil
		},
	}
}

func NewExtensionUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "upgrade [--all] [<command>]",
		Short:     "Upgrade an installed command",
		Aliases:   []string{"up", "update"},
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: extensions.Names(),
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
				toUpgrade = extensions.Names()
			} else {
				toUpgrade = args
			}

			for _, commandName := range toUpgrade {
				command, ok := extensions[commandName]
				if !ok {
					return fmt.Errorf("command %s not found", commandName)
				}

				origin, err := originToUrl(command.Origin)
				if err != nil {
					return fmt.Errorf("could not parse origin: %s", err)
				}

				remote, err := GetRemote(origin)
				if err != nil {
					return fmt.Errorf("could not get remote: %s", err)
				}

				version, err := remote.GetLatestVersion()
				if err != nil {
					return fmt.Errorf("could not get version: %s", err)
				}

				if version == command.Metadata.Version {
					fmt.Printf("Command %s already up to date\n", commandName)
					continue
				}

				originUrl, err := url.Parse(command.Origin)
				if err != nil {
					return err
				}
				targetDir := filepath.Join(xdg.CacheHome, "sunbeam", utils.UrlToFilename(originUrl), version)

				newCommand, err := remote.Download(targetDir, version)
				if err != nil {
					_ = os.RemoveAll(targetDir)
					return fmt.Errorf("unable to upgrade command: %s", err)
				}

				extensions[commandName] = newCommand
				extensions.Save()

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
