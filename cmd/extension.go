package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	_ "embed"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"gopkg.in/yaml.v3"

	"github.com/mattn/go-isatty"
	cp "github.com/otiai10/copy"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

const (
	extensionBinaryName = "sunbeam-extension"
)

//go:embed templates/sunbeam-extension
var extensionTemplate []byte

func NewExtensionCmd(extensionDir string) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:   "extension",
		Short: "Extension commands",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(extensionDir); os.IsNotExist(err) {
				os.MkdirAll(extensionDir, 0755)
			}
			return nil
		},
	}

	extensionCmd.AddCommand(NewExtensionBrowseCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionViewCmd())
	extensionCmd.AddCommand(NewExtensionManageCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionCreateCmd())
	extensionCmd.AddCommand(NewExtensionRenameCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionExecCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionInstallCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionListCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionRemoveCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionUpgradeCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionSearchCmd())

	return extensionCmd
}

func NewExtensionBrowseCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "browse",
		Short: "Browse extensions",
		RunE: func(cmd *cobra.Command, args []string) error {
			generator := func() ([]byte, error) {
				repos, err := utils.SearchSunbeamExtensions("")
				if err != nil {
					return nil, err
				}

				listItems := make([]types.ListItem, 0)

				for _, repo := range repos {
					listItems = append(listItems, types.ListItem{
						Title:    repo.FullName,
						Subtitle: repo.Description,
						Accessories: []string{
							fmt.Sprintf("%d *", repo.StargazersCount),
						},
						Actions: []types.Action{
							{
								Type:  types.PushPageAction,
								Title: "View Readme",
								Page: &types.Target{
									Type:    "dynamic",
									Command: fmt.Sprintf("sunbeam extension view %s", repo.HtmlUrl),
								},
							},
							newExtensionInstallAction(repo.HtmlUrl, "i"),
							{
								Type:  types.OpenUrlAction,
								Title: "Open in Browser",
								Url:   repo.HtmlUrl,
								Key:   "o",
							},
						},
					})
				}

				page := types.Page{
					Type:  types.ListPage,
					Items: listItems,
				}

				return json.Marshal(page)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator()
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}
				fmt.Print(string(output))
				return nil
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
			return nil
		},
	}
}

func newExtensionInstallAction(extensionUrl string, key string) types.Action {
	return types.Action{
		Type:      types.RunAction,
		Title:     "Install",
		OnSuccess: types.ExitOnSuccess,
		Key:       key,
		Command:   fmt.Sprintf("sunbeam extension install ${input:name} %s", extensionUrl),
		Inputs: []types.Input{
			{
				Type:        types.TextFieldInput,
				Name:        "name",
				Title:       "Extension Name",
				Placeholder: "Extension Name",
			},
		},
	}
}

func NewExtensionViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <repo>",
		Short: "View extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := utils.RepositoryFromString(args[0])
			if err != nil {
				return fmt.Errorf("could not parse repository: %s", err)
			}

			generator := func() ([]byte, error) {
				res, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/readme", repo.FullName()))
				if err != nil {
					return nil, fmt.Errorf("could not fetch readme: %s", err)
				}
				defer res.Body.Close()

				if res.StatusCode != http.StatusOK {
					return json.Marshal(types.Page{
						Type: types.DetailPage,
						Preview: &types.Preview{
							Text: fmt.Sprintf("Could not fetch readme: %s", res.Status),
						},
					})
				}

				content, err := io.ReadAll(res.Body)
				if err != nil {
					return nil, fmt.Errorf("could not read readme: %s", err)
				}

				var ReadmePayload struct {
					Content string `json:"content"`
				}
				if err := json.Unmarshal(content, &ReadmePayload); err != nil {
					return nil, fmt.Errorf("could not parse readme: %s", err)
				}

				payload, err := base64.StdEncoding.DecodeString(ReadmePayload.Content)
				if err != nil {
					return nil, fmt.Errorf("could not decode readme: %s", err)
				}

				return json.Marshal(types.Page{
					Type: types.DetailPage,
					Preview: &types.Preview{
						Text: string(payload),
					},
					Actions: []types.Action{
						newExtensionInstallAction(repo.Url().String(), "i"),
					},
				})
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				content, err := generator()
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}

				fmt.Print(string(content))
				return nil
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()

			return nil
		},
	}
}

func NewExtensionManageCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Manage installed extensions",

		RunE: func(cmd *cobra.Command, args []string) error {
			generator := func() ([]byte, error) {
				extensions, err := ListExtensions(extensionDir)
				if err != nil {
					return nil, fmt.Errorf("could not list extensions: %s", err)
				}

				listItems := make([]types.ListItem, 0)
				for _, extension := range extensions {
					listItems = append(listItems, types.ListItem{
						Title: extension,
						Actions: []types.Action{
							{
								Type:  types.PushPageAction,
								Title: "Run Command",
								Page: &types.Target{
									Command: fmt.Sprintf("sunbeam extension exec %s", extension),
								},
							},
							{
								Title:     "Upgrade Extension",
								Type:      types.RunAction,
								OnSuccess: types.ExitOnSuccess,
								Command:   fmt.Sprintf("sunbeam extension upgrade %s", extension),
							},
							{
								Type:      types.RunAction,
								Title:     "Remove Extension",
								OnSuccess: types.ReloadOnSuccess,
								Command:   fmt.Sprintf("sunbeam extension remove %s", extension),
							},
						},
					})
				}

				page := types.Page{
					Type:  types.ListPage,
					Items: listItems,
					Actions: []types.Action{
						{
							Type:      types.RunAction,
							OnSuccess: types.ExitOnSuccess,
							Title:     "Create Extension",
							Command:   "sunbeam extension create ${input:extensionName}",
							Inputs: []types.Input{
								{
									Type:        types.TextFieldInput,
									Name:        "extensionName",
									Title:       "Extension Name",
									Placeholder: "my-extension",
								},
							},
						},
					},
				}

				return json.Marshal(page)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator()
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}
				fmt.Print(string(output))
				return nil
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
			return nil
		},
	}
}

func NewExtensionCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new extension",
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionName, _ := cmd.Flags().GetString("name")
			cwd, _ := os.Getwd()
			extensionDir := path.Join(cwd, extensionName)
			if _, err := os.Stat(extensionDir); !os.IsNotExist(err) {
				return fmt.Errorf("extension already exists: %s", extensionDir)
			}

			if err := os.MkdirAll(extensionDir, 0755); err != nil {
				return fmt.Errorf("could not create extension directory: %s", err)
			}

			extensionScriptPath := path.Join(extensionDir, extensionBinaryName)
			if err := os.WriteFile(extensionScriptPath, extensionTemplate, 0755); err != nil {
				return fmt.Errorf("could not write extension script: %s", err)
			}

			fmt.Println("Extension created successfully!")
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "Extension name")
	return cmd
}

func NewExtensionExecCmd(extensionDir string) *cobra.Command {
	validArgs, _ := ListExtensions(extensionDir)
	return &cobra.Command{
		Use:       "exec",
		Short:     "Execute an installed extension",
		Args:      cobra.MinimumNArgs(1),
		ValidArgs: validArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionName := args[0]

			binPath := path.Join(extensionDir, extensionName, extensionBinaryName)

			if _, err := os.Stat(binPath); os.IsNotExist(err) {
				return fmt.Errorf("extension not found: %s", extensionName)
			}

			command := fmt.Sprintf("%s %s", binPath, strings.Join(args[1:], " "))

			cwd, _ := os.Getwd()
			generator := internal.NewCommandGenerator(command, "", cwd)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator()
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return nil
			}

			runner := internal.NewRunner(generator)

			internal.NewPaginator(runner).Draw()
			return nil
		},
	}
}

func NewExtensionInstallCmd(extensionDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a sunbeam extension from a repository",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			extensionOrigin := args[1]
			if extensionOrigin == "." {
				cwd, _ := os.Getwd()

				bin := path.Join(cwd, extensionBinaryName)
				if _, err := os.Stat(bin); os.IsNotExist(err) {
					return fmt.Errorf("extension binary not found: %s", bin)
				}

				return nil
			}

			_, err := url.Parse(extensionOrigin)
			if err != nil {
				return fmt.Errorf("invalid extension URL: %s", args[0])
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionName, _ := cmd.Flags().GetString("name")
			extensionOrigin := args[0]
			if extensionOrigin == "." {
				cwd, _ := os.Getwd()
				targetDir := path.Join(extensionDir, extensionName)
				err := os.Symlink(cwd, targetDir)
				if err != nil {
					return fmt.Errorf("unable to install local extension: %s", err)
				}

				fmt.Sprintln("Extension installed:", path.Base(cwd))
				return nil
			}

			repository, err := utils.RepositoryFromString(extensionOrigin)
			if err != nil {
				return fmt.Errorf("unable to parse repository: %s", err)
			}

			targetDir := path.Join(extensionDir, extensionName)
			if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
				return fmt.Errorf("extension already installed")
			}

			if release, err := utils.GetLatestRelease(repository); err == nil {
				if err := releaseInstall(release, targetDir); err != nil {
					return fmt.Errorf("unable to install extension: %s", err)
				}
				manifest := ExtensionManifest{
					Name:  repository.Name(),
					Owner: repository.Owner(),
					Host:  "github.com",
					Tag:   release.TagName,
					Path:  targetDir,
				}
				bytes, err := yaml.Marshal(manifest)
				if err != nil {
					return fmt.Errorf("unable to write extension manifest: %s", err)
				}

				manifestPath := path.Join(targetDir, "manifest.yml")
				if err := os.WriteFile(manifestPath, bytes, 0644); err != nil {
					return fmt.Errorf("unable to write extension manifest: %s", err)
				}

				fmt.Println("Extension installed!")
				return nil
			}

			if err := gitInstall(repository, targetDir); err != nil {
				return fmt.Errorf("unable to install extension: %s", err)
			}

			fmt.Println("Extension installed!")
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "Extension name (defaults to repository name)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func releaseInstall(release *utils.Release, targetDir string) error {
	downloadUrl := fmt.Sprintf("https://github.com/pomdtr/sunbeam-vscode/releases/download/%s/sunbeam-extension-%s-%s", release.TagName, runtime.GOOS, runtime.GOARCH)
	res, err := http.Get(downloadUrl)
	if err != nil {
		return fmt.Errorf("unable to install extension: %s", err)
	}
	defer res.Body.Close()

	os.MkdirAll(targetDir, 0755)
	out, err := os.OpenFile(path.Join(targetDir, extensionBinaryName), os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("unable to install extension: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return err
}

func gitInstall(repository *utils.Repository, targetDir string) error {
	tempDir, err := os.MkdirTemp("", "sunbeam-*")
	if err != nil {
		return fmt.Errorf("unable to install extension: %s", err)
	}
	defer os.RemoveAll(tempDir)

	if err = utils.GitClone(repository, tempDir); err != nil {
		return fmt.Errorf("unable to install extension: %s", err)
	}

	binPath := path.Join(tempDir, extensionBinaryName)
	if os.Stat(binPath); os.IsNotExist(err) {
		return fmt.Errorf("extension binary not found: %s", binPath)
	}

	return cp.Copy(tempDir, targetDir)
}

func NewExtensionRenameCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename an installed extension",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := path.Join(extensionDir, args[0])
			if _, err := os.Stat(src); os.IsNotExist(err) {
				return fmt.Errorf("source directory not found: %s", src)
			}

			dst := path.Join(extensionDir, args[1])
			if _, err := os.Stat(dst); !os.IsNotExist(err) {
				return fmt.Errorf("destination directory already exists: %s", dst)
			}

			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("unable to rename extension: %s", err)
			}
			return nil
		},
	}
}

func NewExtensionListCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed extension commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions, err := ListExtensions(extensionDir)
			if err != nil {
				return fmt.Errorf("unable to list extensions: %s", err)
			}

			for _, extension := range extensions {
				fmt.Println(extension)
			}

			return nil
		},
	}
}

func NewExtensionRemoveCmd(extensionDir string) *cobra.Command {
	validArgs, _ := ListExtensions(extensionDir)
	return &cobra.Command{
		Use:       "remove",
		Short:     "Remove an installed extension",
		Args:      cobra.ExactArgs(1),
		ValidArgs: validArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionName := args[0]
			targetDir := path.Join(extensionDir, extensionName)
			if _, err := os.Stat(targetDir); os.IsNotExist(err) {
				return fmt.Errorf("extension not installed: %s", extensionName)
			}

			if err := os.RemoveAll(targetDir); err != nil {
				return fmt.Errorf("unable to remove extension: %s", err)
			}

			fmt.Println("Extension removed:", extensionName)
			return nil
		},
	}
}

func NewExtensionUpgradeCmd(extensionDir string) *cobra.Command {
	validArgs, _ := ListExtensions(extensionDir)
	return &cobra.Command{
		Use:       "upgrade",
		Short:     "Upgrade an installed extension",
		Args:      cobra.ExactArgs(1),
		ValidArgs: validArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionName := args[0]

			extensionPath := path.Join(extensionDir, extensionName)
			if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
				return fmt.Errorf("extension not installed: %s", args[0])
			}

			manifestPath := path.Join(extensionPath, "manifest.yml")
			if _, err := os.Stat(manifestPath); err == nil {
				bytes, err := os.ReadFile(manifestPath)
				if err != nil {
					return fmt.Errorf("unable to upgrade extension: %s", err)
				}
				var manifest ExtensionManifest
				if err := yaml.Unmarshal(bytes, &manifest); err != nil {
					return fmt.Errorf("unable to upgrade extension: %s", err)
				}

				repositoryUrl := &url.URL{
					Scheme: "https",
					Host:   manifest.Host,
					Path:   fmt.Sprintf("/%s/%s", manifest.Owner, manifest.Name),
				}

				release, err := utils.GetLatestRelease(utils.NewRepository(repositoryUrl))
				if err != nil {
					return fmt.Errorf("unable to upgrade extension: %s", err)
				}

				if release.TagName == manifest.Tag {
					fmt.Println("Extension already up to date")
					return nil
				}

				if err := releaseInstall(release, extensionPath); err != nil {
					return fmt.Errorf("unable to upgrade extension: %s", err)
				}

				return nil
			}

			if _, err := os.Stat(filepath.Join(extensionPath, ".git")); os.IsNotExist(err) {
				return fmt.Errorf("extension not installed from git: %s", args[0])
			}

			if err := utils.GitPull(extensionPath); err != nil {
				return fmt.Errorf("unable to upgrade extension: %s", err)
			}

			fmt.Sprintln("Extension upgraded:", args[0])
			return nil
		},
	}
}

func NewExtensionSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search",
		Short: "Search for repositories with the sunbeam-extension topic",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var query string
			if len(args) == 1 {
				query = args[0]
			}
			extensionRepos, err := utils.SearchSunbeamExtensions(query)

			if err != nil {
				return fmt.Errorf("unable to search for extensions: %s", err)
			}

			for _, repo := range extensionRepos {
				fmt.Println(repo.Name)
			}
			return nil
		},
	}
}

func ListExtensions(extensionDir string) ([]string, error) {
	if _, err := os.Stat(extensionDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(extensionDir)

	if err != nil {
		return nil, fmt.Errorf("unable to list extensions: %s", err)
	}

	extensions := make([]string, 0)
	for _, entry := range entries {
		binary := path.Join(extensionDir, entry.Name(), extensionBinaryName)
		if _, err := os.Stat(binary); os.IsNotExist(err) {
			continue
		}

		extensions = append(extensions, entry.Name())
	}

	return extensions, nil
}

type ExtensionManifest struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
	Host  string `yaml:"host"`
	Tag   string `yaml:"tag"`
	Path  string `yaml:"path"`
}
