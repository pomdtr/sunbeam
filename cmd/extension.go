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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat(extensionDir); os.IsNotExist(err) {
				os.MkdirAll(extensionDir, 0755)
			}
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
		Run: func(cmd *cobra.Command, args []string) {
			generator := func(string) ([]byte, error) {
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
								Type:      types.RunAction,
								Title:     "View Readme",
								OnSuccess: types.PushOnSuccess,
								Command:   fmt.Sprintf("sunbeam extension view %s", repo.HtmlUrl),
							},
							newExtensionInstallAction(repo.HtmlUrl, "ctrl+i"),
							{
								Type:  types.OpenAction,
								Title: "Open in Browser",
								Url:   repo.HtmlUrl,
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
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("Could not generate page: %s", err)
				}
				fmt.Print(string(output))
				return
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
		},
	}
}

func newExtensionInstallAction(extensionUrl string, shortcut string) types.Action {
	return types.Action{
		Type:      types.RunAction,
		Title:     "Install",
		OnSuccess: types.PushOnSuccess,
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
		Run: func(cmd *cobra.Command, args []string) {
			repo, err := utils.RepositoryFromString(args[0])
			if err != nil {
				exitWithErrorMsg("Could not parse repository: %s", err)
			}

			generator := func(string) ([]byte, error) {
				res, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/readme", repo.FullName()))
				if err != nil {
					return nil, fmt.Errorf("could not fetch readme: %s", err)
				}
				defer res.Body.Close()

				if res.StatusCode != http.StatusOK {
					return json.Marshal(types.Page{
						Type:     types.DetailPage,
						Text:     fmt.Sprintf("Could not fetch readme: %s", res.Status),
						Language: "markdown",
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
					Type:     types.DetailPage,
					Text:     string(payload),
					Language: "markdown",
					Actions: []types.Action{
						newExtensionInstallAction(repo.Url().String(), "ctrl+i"),
					},
				})
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				content, err := generator("")
				if err != nil {
					exitWithErrorMsg("Could not generate page: %s", err)
				}

				fmt.Print(string(content))
				return
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
		},
	}
}

func NewExtensionManageCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Manage installed extensions",
		Run: func(cmd *cobra.Command, args []string) {
			generator := func(string) ([]byte, error) {
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
								Type:      types.RunAction,
								Title:     "Run Command",
								OnSuccess: types.PushOnSuccess,
								Command:   fmt.Sprintf("sunbeam extension exec %s", extension),
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
					Type: types.ListPage,
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
					Items: listItems,
				}

				return json.Marshal(page)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("Could not generate page: %s", err)
				}
				fmt.Print(string(output))
				return
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
		},
	}
}

func NewExtensionCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new extension",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			extensionName := args[0]
			cwd, _ := os.Getwd()
			extensionDir := path.Join(cwd, extensionName)
			if _, err := os.Stat(extensionDir); !os.IsNotExist(err) {
				exitWithErrorMsg("Extension already exists: %s", extensionDir)
			}

			if err := os.MkdirAll(extensionDir, 0755); err != nil {
				exitWithErrorMsg("Could not create extension directory: %s", err)
			}

			extensionScriptPath := path.Join(extensionDir, extensionBinaryName)
			if err := os.WriteFile(extensionScriptPath, extensionTemplate, 0755); err != nil {
				exitWithErrorMsg("Could not write extension script: %s", err)
			}

			fmt.Println("Extension created successfully!")
		},
	}
}

func NewExtensionExecCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "exec",
		Short: "Execute an installed extension",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			extensionName := args[0]

			binPath := path.Join(extensionDir, extensionName, extensionBinaryName)

			if _, err := os.Stat(binPath); os.IsNotExist(err) {
				exitWithErrorMsg("Extension not found: %s", extensionName)
			}

			command := fmt.Sprintf("%s %s", binPath, strings.Join(args[1:], " "))

			cwd, _ := os.Getwd()
			generator := internal.NewCommandGenerator(command, "", cwd)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			runner := internal.NewRunner(generator)

			internal.NewPaginator(runner).Draw()
		},
	}
}

func NewExtensionInstallCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install a sunbeam extension from a repository",
		Args:  cobra.ExactArgs(2),
		PreRun: func(cmd *cobra.Command, args []string) {
			extensionOrigin := args[1]
			if extensionOrigin == "." {
				cwd, _ := os.Getwd()

				bin := path.Join(cwd, extensionBinaryName)
				if _, err := os.Stat(bin); os.IsNotExist(err) {
					exitWithErrorMsg("Extension binary not found: %s", bin)
				}

				return
			}

			_, err := url.Parse(extensionOrigin)
			if err != nil {
				exitWithErrorMsg("Invalid extension URL: %s", args[0])
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			extensionName := args[0]
			extensionOrigin := args[1]
			if extensionOrigin == "." {
				cwd, _ := os.Getwd()
				targetDir := path.Join(extensionDir, extensionName)
				err := os.Symlink(cwd, targetDir)
				if err != nil {
					exitWithErrorMsg("Unable to install local extension: %s", err)
				}

				fmt.Sprintln("Extension installed:", path.Base(cwd))
				return
			}

			repository, err := utils.RepositoryFromString(extensionOrigin)
			if err != nil {
				exitWithErrorMsg("Unable to parse repository: %s", err)
			}

			targetDir := path.Join(extensionDir, extensionName)
			if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
				exitWithErrorMsg("Extension already installed: %s", repository.Name())
			}

			if release, err := utils.GetLatestRelease(repository); err == nil {
				if err := releaseInstall(release, targetDir); err != nil {
					exitWithErrorMsg("Unable to install extension: %s", err)
				}
				fmt.Println("Extension installed:", repository.Name())
			}

			if err := gitInstall(repository, targetDir); err != nil {
				exitWithErrorMsg("Unable to install extension: %s", err)
			}

			fmt.Println("Extension installed:", repository.Name())
		},
	}
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
		exitWithErrorMsg("Unable to install extension: %s", err)
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
		Run: func(cmd *cobra.Command, args []string) {
			src := path.Join(extensionDir, args[0])
			if _, err := os.Stat(src); os.IsNotExist(err) {
				exitWithErrorMsg("Source directory not found: %s", src)
			}

			dst := path.Join(extensionDir, args[1])
			if _, err := os.Stat(dst); !os.IsNotExist(err) {
				exitWithErrorMsg("Destination directory already exists: %s", dst)
			}

			if err := os.Rename(src, dst); err != nil {
				exitWithErrorMsg("Unable to rename extension: %s", err)
			}
		},
	}
}

func NewExtensionListCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed extension commands",
		Run: func(cmd *cobra.Command, args []string) {
			extensions, err := ListExtensions(extensionDir)
			if err != nil {
				exitWithErrorMsg("Unable to list extensions: %s", err)
			}

			for _, extension := range extensions {
				fmt.Println(extension)
			}

		},
	}
}

func NewExtensionRemoveCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove an installed extension",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			extensionName := args[0]
			targetDir := path.Join(extensionDir, extensionName)
			if _, err := os.Stat(targetDir); os.IsNotExist(err) {
				exitWithErrorMsg("Extension not installed: %s", extensionName)
			}

			if err := os.RemoveAll(targetDir); err != nil {
				exitWithErrorMsg("Unable to remove extension: %s", err)
			}

			fmt.Println("Extension removed:", extensionName)
		},
	}
}

func NewExtensionUpgradeCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade an installed extension",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			extensionName := args[0]

			extensionPath := path.Join(extensionDir, extensionName)
			if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
				exitWithErrorMsg("Extension not installed: %s", args[0])
			}

			if _, err := os.Stat(filepath.Join(extensionPath, ".git")); os.IsNotExist(err) {
				exitWithErrorMsg("Extension not installed from git: %s", args[0])
			}

			if err := utils.GitPull(extensionPath); err != nil {
				exitWithErrorMsg("Unable to upgrade extension: %s", err)
			}

			fmt.Sprintln("Extension upgraded:", args[0])
		},
	}
}

func NewExtensionSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search",
		Short: "Search for repositories with the sunbeam-extension topic",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var query string
			if len(args) == 1 {
				query = args[0]
			}
			extensionRepos, err := utils.SearchSunbeamExtensions(query)

			if err != nil {
				exitWithErrorMsg("Unable to search for extensions: %s", err)
			}

			for _, repo := range extensionRepos {
				fmt.Println(repo.Name)
			}
		},
	}
}

func ListExtensions(extensionDir string) ([]string, error) {
	if _, err := os.Stat(extensionDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(extensionDir)

	if err != nil {
		exitWithErrorMsg("Unable to list extensions: %s", err)
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
