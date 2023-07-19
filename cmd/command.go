package cmd

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/adrg/xdg"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/types"
	"golang.org/x/term"

	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

var (
	commandRoot  string
	commands     map[string]CommandManifest = make(map[string]CommandManifest)
	commandNames                            = []string{}
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
		manifest, err := LoadManifest(filepath.Join(commandRoot, dir.Name()))
		if err != nil {
			continue
		}
		commands[dir.Name()] = manifest
		commandNames = append(commandNames, dir.Name())
	}

	if err != nil {
		panic(err)
	}
}

type CommandManifest struct {
	Origin   string `json:"origin"`
	Version  string `json:"version"`
	RootDir  string `json:"root"`
	Metadata `json:"metadata"`
}

func LoadManifest(dir string) (CommandManifest, error) {
	f, err := os.Open(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return CommandManifest{}, err
	}
	defer f.Close()

	var m CommandManifest
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return CommandManifest{}, err
	}

	return m, nil
}

func (m CommandManifest) Save(dir string) error {
	f, err := os.OpenFile(filepath.Join(dir, "manifest.json"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(m)
}

type Metadata struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Entrypoint  string `json:"entrypoint,omitempty"`
}

var MetadataRegexp = regexp.MustCompile(`@sunbeam.(?P<key>[A-Za-z0-9]+)\s(?P<value>[\S ]+)`)

func ExtractScriptMetadata(script []byte) (Metadata, error) {
	var cmd Metadata
	matches := MetadataRegexp.FindAllSubmatch(script, -1)
	if len(matches) == 0 {
		return cmd, fmt.Errorf("no metadata found")
	}

	metadata := Metadata{}
	for _, match := range matches {
		key := string(match[1])
		value := string(match[2])

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
	originUrl, err := url.Parse(origin)
	if err != nil {
		return nil, fmt.Errorf("could not parse origin: %s", err)
	}

	if originUrl.Scheme == "" || originUrl.Scheme == "file" {
		remotePath, err := filepath.Abs(originUrl.Path)
		if err != nil {
			return nil, fmt.Errorf("could not get absolute path: %s", err)
		}

		info, err := os.Stat(origin)
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("unable to find command: %s", err)
		} else if err != nil {
			return nil, fmt.Errorf("unable to stat command: %s", err)
		}

		if info.IsDir() {
			return LocalRemote{
				path: filepath.Join(remotePath, "sunbeam.json"),
			}, nil
		}

		return LocalRemote{
			path: remotePath,
		}, nil
	}

	switch originUrl.Hostname() {
	case "www.github.com", "github.com":
		return GithubRemote{
			origin: originUrl,
		}, nil
	case "raw.githubusercontent.com":
		return ScriptRemote{
			origin: originUrl,
		}, nil
	case "www.val.town", "val.town":
		return NewValTownRemote(originUrl)
	default:
		return RestRemote{
			origin: originUrl,
		}, nil
	}
}

type LocalRemote struct {
	path string
}

func (r LocalRemote) Download(targetDir string, version string) error {
	info, err := os.Stat(r.path)
	if err != nil {
		return fmt.Errorf("unable to stat command: %s", err)
	}

	if info.Name() == "sunbeam.json" {
		f, err := os.Open(r.path)
		if err != nil {
			return fmt.Errorf("unable to open manifest: %s", err)
		}

		var metadata Metadata
		if err := json.NewDecoder(f).Decode(&metadata); err != nil {
			return fmt.Errorf("unable to load metadata: %s", err)
		}

		manifest := CommandManifest{
			Origin:   r.path,
			Version:  version,
			RootDir:  filepath.Dir(r.path),
			Metadata: metadata,
		}

		if err := manifest.Save(targetDir); err != nil {
			return fmt.Errorf("unable to save manifest: %s", err)
		}

		return nil
	}

	bs, err := os.ReadFile(r.path)
	if err != nil {
		return fmt.Errorf("unable to read command: %s", err)
	}

	metadata, err := ExtractScriptMetadata(bs)
	if err != nil {
		return fmt.Errorf("unable to extract script metadata: %s", err)
	}
	metadata.Entrypoint = filepath.Base(r.path)

	manifest := CommandManifest{
		Origin:   r.path,
		Version:  version,
		RootDir:  filepath.Dir(r.path),
		Metadata: metadata,
	}

	return manifest.Save(targetDir)
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

	root := filepath.Join(srcDir, "sunbeam.json")

	var metadata Metadata
	if err := json.Unmarshal([]byte(root), &metadata); err != nil {
		return fmt.Errorf("unable to load metadata: %s", err)
	}

	manifest := CommandManifest{
		Origin:   r.origin.String(),
		Version:  version,
		RootDir:  "src",
		Metadata: metadata,
	}

	if err := manifest.Save(targetDir); err != nil {
		return fmt.Errorf("unable to save manifest: %s", err)
	}

	return nil
}

type ScriptRemote struct {
	origin *url.URL
}

func (r ScriptRemote) GetLatestVersion() (string, error) {
	res, err := http.Get(r.origin.String())
	if err != nil {
		return "", fmt.Errorf("unable to download script: %s", err)
	}
	defer res.Body.Close()

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body: %s", err)
	}

	hash := sha256.Sum256(bs)
	return hex.EncodeToString(hash[:])[:8], nil
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

	entrypointPath := filepath.Join(targetDir, "sunbeam-command")
	if err := os.WriteFile(entrypointPath, bs, 0755); err != nil {
		return fmt.Errorf("unable to write script: %s", err)
	}

	metadata, err := ExtractScriptMetadata(bs)
	if err != nil {
		return fmt.Errorf("unable to extract script metadata: %s", err)
	}

	manifest := CommandManifest{
		Origin:   r.origin.String(),
		Version:  version,
		RootDir:  ".",
		Metadata: metadata,
	}

	if err := manifest.Save(targetDir); err != nil {
		return fmt.Errorf("unable to save manifest: %s", err)
	}

	return nil
}

var valTownRegexp = regexp.MustCompile(`\/v\/([^\/]+)\.([^\/]+)$`)

type ValTownRemote struct {
	origin *url.URL
	author string
	name   string
}

type Author struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Output struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Val struct {
	Author     Author `json:"author"`
	Code       string `json:"code"`
	Error      string `json:"error"`
	ID         string `json:"id"`
	Logs       []int  `json:"logs"`
	Name       string `json:"name"`
	Output     Output `json:"output"`
	Public     bool   `json:"public"`
	RunEndAt   string `json:"runEndAt"`
	RunStartAt string `json:"runStartAt"`
	Version    int    `json:"version"`
}

func NewValTownRemote(origin *url.URL) (*ValTownRemote, error) {
	matches := valTownRegexp.FindStringSubmatch(origin.Path)

	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid val.town url")
	}

	return &ValTownRemote{
		origin: origin,
		author: matches[1],
		name:   matches[2],
	}, nil
}

func (t ValTownRemote) Val() string {
	return fmt.Sprintf("%s.%s", t.author, t.name)
}

func (r ValTownRemote) FetchVal() (Val, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.val.town/v1/alias/%s/%s", r.author, r.name), nil)
	if err != nil {
		return Val{}, fmt.Errorf("unable to fetch val: %s", err)
	}

	if env, ok := os.LookupEnv("VALTOWN_TOKEN"); ok {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", env))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Val{}, fmt.Errorf("unable to fetch val: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Val{}, fmt.Errorf("unable to fetch val: %s", resp.Status)
	}

	var val Val
	if err := json.NewDecoder(resp.Body).Decode(&val); err != nil {
		return Val{}, fmt.Errorf("unable to decode val: %s", err)
	}

	return val, nil
}

func (r ValTownRemote) GetLatestVersion() (string, error) {
	val, err := r.FetchVal()
	if err != nil {
		return "", err
	}

	return strconv.Itoa(val.Version), nil
}

//go:embed templates/run-val.sh
var rawRunValTemplate string
var runValTemplate = template.Must(template.New("run-val.sh").Parse(rawRunValTemplate))

func (r ValTownRemote) Download(targetDir string, version string) error {
	val, err := r.FetchVal()
	if err != nil {
		return err
	}

	metadata, err := ExtractScriptMetadata([]byte(val.Code))
	if err != nil {
		return fmt.Errorf("unable to extract metadata: %s", err)
	}

	f, err := os.OpenFile(filepath.Join(targetDir, "sunbeam-command"), os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("unable to write script: %s", err)
	}
	defer f.Close()

	if err := runValTemplate.Execute(f, map[string]any{
		"Val": r.Val(),
	}); err != nil {
		return fmt.Errorf("unable to render template: %s", err)
	}

	manifest := CommandManifest{
		Origin:   r.origin.String(),
		Version:  version,
		RootDir:  ".",
		Metadata: metadata,
	}

	if err := manifest.Save(targetDir); err != nil {
		return fmt.Errorf("unable to save manifest: %s", err)
	}

	return nil
}

type RestRemote struct {
	origin *url.URL
}

func (r RestRemote) GetLatestVersion() (string, error) {
	resp, err := http.Get(r.origin.String())
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

//go:embed templates/run-rest.sh
var rawRestTemplate string
var restTemplate = template.Must(template.New("run-server.sh").Parse(rawRestTemplate))

func (r RestRemote) Download(targetDir string, version string) error {
	resp, err := http.Get(r.origin.String())
	if err != nil {
		return fmt.Errorf("unable to fetch latest version: %s", err)
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(filepath.Join(targetDir, "sunbeam-command"), os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("unable to write script: %s", err)
	}
	if err := restTemplate.Execute(f, map[string]any{
		"Remote": r.origin.String(),
	}); err != nil {
		return fmt.Errorf("unable to render template: %s", err)
	}

	var metadata Metadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return fmt.Errorf("unable to decode metadata: %s", err)
	}
	metadata.Entrypoint = "sunbeam-command"

	manifest := CommandManifest{
		Origin:   r.origin.String(),
		Version:  version,
		RootDir:  ".",
		Metadata: metadata,
	}

	if err := manifest.Save(targetDir); err != nil {
		return fmt.Errorf("unable to save manifest: %s", err)
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
								Type:  types.ExecAction,
								Key:   "u",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"command", "upgrade", name},
								},
							},
							{
								Type:  types.ExecAction,
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
								Type:  types.ExecAction,
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

			for commandName, command := range commands {
				printer.AddField(commandName)
				printer.AddField(command.Title)
				printer.AddField(command.Version)
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
