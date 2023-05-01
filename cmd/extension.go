package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "embed"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"

	cp "github.com/otiai10/copy"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

var (
	extensionBinaryName = "sunbeam-extension"
)

func init() {
	if runtime.GOOS == "windows" {
		extensionBinaryName += ".exe"
	}
}

//go:embed templates/sunbeam-extension
var extensionTemplate []byte

func NewExtensionCmd(extensionDir string) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:     "extension",
		Short:   "Extension commands",
		GroupID: coreGroupID,
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
	extensionCmd.AddCommand(NewExtensionInstallCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionRenameCmd(extensionDir))
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
			generator := func() (*types.Page, error) {
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
								Type:  types.PushAction,
								Title: "View Readme",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"extension", "view", repo.FullName},
								},
							},
							{
								Type:  types.RunAction,
								Title: "Install",
								Key:   "i",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"extension", "install", "${input:alias}", repo.FullName},
								},
								Inputs: []types.Input{
									{
										Name:        "alias",
										Type:        types.TextFieldInput,
										Title:       "Alias",
										Placeholder: "my-command-alias",
									},
								},
							},
							{
								Type:   types.OpenAction,
								Title:  "Open in Browser",
								Target: repo.HtmlUrl,
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

			return Draw(generator)
		},
	}
}

func NewExtensionViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <repo>",
		Short: "View extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoUrl, err := utils.RepositoryUrl(args[0])
			if err != nil {
				return fmt.Errorf("could not parse repository: %s", err)
			}

			repo := utils.NewRepository(repoUrl)
			generator := func() (*types.Page, error) {
				res, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/readme", repo.FullName()))
				if err != nil {
					return nil, fmt.Errorf("could not fetch readme: %s", err)
				}
				defer res.Body.Close()

				if res.StatusCode != http.StatusOK {
					return &types.Page{
						Type: types.DetailPage,
						Preview: &types.Preview{
							HighLight: "markdown",
							Text:      fmt.Sprintf("Could not fetch readme: %s", res.Status),
						},
					}, nil
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

				page := types.Page{
					Type: types.DetailPage,
					Preview: &types.Preview{
						HighLight: "markdown",
						Text:      string(payload),
					},
					Actions: []types.Action{
						{
							Type:  types.RunAction,
							Title: "Install",
							Key:   "i",
							Command: &types.Command{
								Name: os.Args[0],
								Args: []string{"extension", "install", "${input:alias}", repo.FullName()},
							},
							Inputs: []types.Input{
								{
									Name:        "alias",
									Type:        types.TextFieldInput,
									Title:       "Alias",
									Placeholder: "my-command-alias",
								},
							},
						},
					},
				}

				return &page, nil
			}

			return Draw(generator)
		},
	}
}

func NewExtensionManageCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Manage installed extensions",

		RunE: func(cmd *cobra.Command, args []string) error {
			generator := func() (*types.Page, error) {
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
								Title: "Run Extension",
								Type:  types.PushAction,
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{extension},
								},
							},
							{
								Title: "Upgrade Extension",
								Type:  types.RunAction,
								Key:   "u",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"extension", "upgrade", extension},
								},
							},
							{
								Type:            types.RunAction,
								Title:           "Rename Extension",
								Key:             "r",
								ReloadOnSuccess: true,
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"extension", "rename", extension, "${input:name}"},
								},
								Inputs: []types.Input{
									{
										Name:        "name",
										Type:        types.TextFieldInput,
										Title:       "Name",
										Placeholder: "my-alias",
									},
								},
							},
							{
								Type:            types.RunAction,
								Title:           "Remove Extension",
								Key:             "d",
								ReloadOnSuccess: true,
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"extension", "remove", extension},
								},
							},
						},
					})
				}

				return &types.Page{
					Type:  types.ListPage,
					Items: listItems,
				}, nil
			}

			return Draw(generator)
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
			extensionDir := filepath.Join(cwd, extensionName)
			if _, err := os.Stat(extensionDir); !os.IsNotExist(err) {
				return fmt.Errorf("extension already exists: %s", extensionDir)
			}

			if err := os.MkdirAll(extensionDir, 0755); err != nil {
				return fmt.Errorf("could not create extension directory: %s", err)
			}

			extensionScriptPath := filepath.Join(extensionDir, extensionBinaryName)
			if err := os.WriteFile(extensionScriptPath, extensionTemplate, 0755); err != nil {
				return fmt.Errorf("could not write extension script: %s", err)
			}

			fmt.Println("Extension created successfully!")
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "Extension name")
	cmd.MarkFlagRequired("name")
	return cmd
}

func NewExtensionRenameCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "rename [old] [new]",
		Short: "Rename an extension",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldPath := filepath.Join(extensionDir, args[0])
			newPath := filepath.Join(extensionDir, args[1])

			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("could not rename extension: %s", err)
			}

			return nil
		},
	}
}

func NewExtensionInstallCmd(extensionDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [alias] [extension]",
		Short: "Install a sunbeam extension from a folder/gist/repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := filepath.Join(extensionDir, args[0])

			if args[1] == "." {
				// Return an error on Windows, as symlinks are not supported
				if runtime.GOOS == "windows" {
					return fmt.Errorf("local install are not supported on Windows")
				}

				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("could not get current working directory: %s", err)
				}

				if err := os.Symlink(cwd, targetDir); err != nil {
					return fmt.Errorf("could not symlink extension: %s", err)
				}

				fmt.Println("Extension installed successfully!")
				return nil
			}

			url, err := url.Parse(args[1])
			if err != nil {
				return fmt.Errorf("could not parse extension url: %s", err)
			}

			if url.Host == "gist.github.com" {
				parts := strings.Split(url.Path, "/")
				if len(parts) < 2 {
					return fmt.Errorf("invalid gist url")
				}

				gistID := parts[len(parts)-1]
				if err := gistInstall(gistID, targetDir); err != nil {
					return fmt.Errorf("could not install extension: %s", err)
				}

			}

			if url.Host == "github.com" {
				repository := utils.NewRepository(url)
				if err != nil {
					return fmt.Errorf("unable to parse repository: %s", err)
				}

				if err := installExtension(repository, targetDir); err != nil {
					return fmt.Errorf("could not install extension: %s", err)
				}

				open, _ := cmd.Flags().GetBool("open")
				if open {
					return Draw(internal.NewCommandGenerator(&types.Command{
						Name: os.Args[0],
						Args: []string{"run", repository.FullName()},
					}))
				}

				fmt.Println("Extension installed successfully!")
				return nil
			}

			return fmt.Errorf("unsupported extension url")
		},
	}

	cmd.Flags().BoolP("open", "o", false, "Open extension after installation")

	return cmd
}

func installExtension(repository *utils.Repository, targetDir string) error {
	if _, err := os.Stat(filepath.Join(targetDir, extensionBinaryName)); !os.IsNotExist(err) {
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
		bytes, err := json.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("unable to write extension manifest: %s", err)
		}

		manifestPath := filepath.Join(targetDir, "manifest.yml")
		if err := os.WriteFile(manifestPath, bytes, 0644); err != nil {
			return fmt.Errorf("unable to write extension manifest: %s", err)
		}

		return nil
	}

	if runtime.GOOS == "windows" {
		return fmt.Errorf("script based extensions are not supported on Windows")
	}

	if err := gitInstall(repository, targetDir); err != nil {
		return fmt.Errorf("unable to install extension: %s", err)
	}

	return nil
}

func releaseInstall(release *utils.Release, targetDir string) error {
	downloadUrl := fmt.Sprintf("https://github.com/pomdtr/sunbeam-vscode/releases/download/%s/sunbeam-extension-%s-%s", release.TagName, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		downloadUrl += ".exe"
	}
	res, err := http.Get(downloadUrl)
	if err != nil {
		return fmt.Errorf("unable to install extension: %s", err)
	}
	defer res.Body.Close()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("unable to install extension: %s", err)
	}
	out, err := os.OpenFile(filepath.Join(targetDir, extensionBinaryName), os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("unable to install extension: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return err
}

func gistInstall(gistID string, targetDir string) error {
	gist, err := utils.FetchGist(gistID)
	if err != nil {
		return fmt.Errorf("could not fetch gist: %s", err)
	}

	sunbeamFile, ok := gist.Files["sunbeam-extension"]
	if !ok {
		return fmt.Errorf("gist does not contain a sunbeam-extension file")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("unable to install extension: %s", err)
	}

	if err := os.WriteFile(filepath.Join(targetDir, extensionBinaryName), []byte(sunbeamFile.Content), 0755); err != nil {
		return fmt.Errorf("unable to install extension: %s", err)
	}

	return nil
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

	binPath := filepath.Join(tempDir, extensionBinaryName)
	if os.Stat(binPath); os.IsNotExist(err) {
		return fmt.Errorf("extension binary not found: %s", binPath)
	}

	return cp.Copy(tempDir, targetDir)
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
			targetDir := filepath.Join(extensionDir, args[0])
			if _, err := os.Stat(targetDir); os.IsNotExist(err) {
				return fmt.Errorf("extension %s not installed", args[0])
			}

			if err := os.RemoveAll(targetDir); err != nil {
				return fmt.Errorf("unable to remove extension: %s", err)
			}

			fmt.Println("extension removed")
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

			extensionPath := filepath.Join(extensionDir, extensionName)
			if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
				return fmt.Errorf("extension not installed: %s", args[0])
			}

			manifestPath := filepath.Join(extensionPath, "manifest.yml")
			if _, err := os.Stat(manifestPath); err == nil {
				bytes, err := os.ReadFile(manifestPath)
				if err != nil {
					return fmt.Errorf("unable to upgrade extension: %s", err)
				}
				var manifest ExtensionManifest
				if err := json.Unmarshal(bytes, &manifest); err != nil {
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

	extensionDirs, err := os.ReadDir(extensionDir)
	if err != nil {
		return nil, fmt.Errorf("unable to list extensions: %s", err)
	}

	extensions := make([]string, 0)
	for _, dir := range extensionDirs {
		binPath := filepath.Join(extensionDir, dir.Name(), extensionBinaryName)
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			continue
		}

		extensions = append(extensions, dir.Name())
	}

	return extensions, nil
}

type ExtensionManifest struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
	Host  string `json:"host"`
	Tag   string `json:"tag"`
	Path  string `json:"path"`
}
