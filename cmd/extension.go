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
	"strings"

	_ "embed"

	"github.com/pomdtr/sunbeam/types"

	"github.com/mattn/go-isatty"
	cp "github.com/otiai10/copy"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

//go:embed templates/extension.sh
var extensionTemplate []byte

func NewExtensionCmd(extensionDir string, validator tui.PageValidator) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:   "extension",
		Short: "Extension commands",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat(extensionDir); os.IsNotExist(err) {
				os.MkdirAll(extensionDir, 0755)
			}
		},
	}

	extensionCmd.AddCommand(NewExtensionBrowseCmd(extensionDir, validator))
	extensionCmd.AddCommand(NewExtensionViewCmd(validator))
	extensionCmd.AddCommand(NewExtensionManageCmd(extensionDir, validator))
	extensionCmd.AddCommand(NewExtensionCreateCmd())
	extensionCmd.AddCommand(NewExtensionExecCmd(extensionDir, validator))
	extensionCmd.AddCommand(NewExtensionInstallCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionListCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionRemoveCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionUpgradeCmd(extensionDir))
	extensionCmd.AddCommand(NewExtensionSearchCmd())

	return extensionCmd
}

func NewExtensionBrowseCmd(extensionDir string, validator tui.PageValidator) *cobra.Command {
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
								RawTitle:  "View Readme",
								OnSuccess: types.PushOnSuccess,
								Command:   fmt.Sprintf("sunbeam extension view %s", repo.HtmlUrl),
							},
							{
								Type:     types.RunAction,
								RawTitle: "Install",
								Command:  fmt.Sprintf("sunbeam extension install %s", repo.HtmlUrl),
							},
							{
								Type:     types.OpenAction,
								RawTitle: "Open in Browser",
								Shortcut: "ctrl+o",
								Url:      repo.HtmlUrl,
							},
						},
					})
				}

				page := types.Page{
					Type: types.ListPage,
					List: &types.List{
						Items: listItems,
					},
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

			cwd, _ := os.Getwd()
			runner := tui.NewRunner(generator, validator, cwd)
			tui.NewModel(runner).Draw()
		},
	}
}

func NewExtensionViewCmd(validator tui.PageValidator) *cobra.Command {
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
					Detail: &types.Detail{
						Text:     string(payload),
						Language: "markdown",
					},
				}

				return json.Marshal(page)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				content, err := generator("")
				if err != nil {
					exitWithErrorMsg("Could not generate page: %s", err)
				}

				fmt.Print(string(content))
				return
			}

			cwd, _ := os.Getwd()
			runner := tui.NewRunner(generator, validator, cwd)
			tui.NewModel(runner).Draw()
		},
	}
}

func NewExtensionManageCmd(extensionDir string, validator tui.PageValidator) *cobra.Command {
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
								RawTitle:  "Run Command",
								OnSuccess: types.PushOnSuccess,
								Command:   fmt.Sprintf("sunbeam extension exec %s", extension),
							},
							{
								RawTitle: "Upgrade Extension",
								Type:     types.RunAction,
								Command:  fmt.Sprintf("sunbeam extension upgrade %s", extension),
								Shortcut: "ctrl+u",
							},
							{
								Type:      types.RunAction,
								RawTitle:  "Remove Extension",
								Shortcut:  "ctrl+x",
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
							Type:     types.RunAction,
							RawTitle: "Create Extension",
							Command:  "sunbeam extension create ${input:extensionName}",
							Inputs: []types.FormInput{
								{
									Type:        types.TextField,
									Name:        "extensionName",
									Title:       "Extension Name",
									Placeholder: "my-extension",
								},
							},
							Shortcut: "ctrl+n",
						},
					},
					List: &types.List{
						Items: listItems,
					},
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

			cwd, _ := os.Getwd()
			runner := tui.NewRunner(generator, validator, cwd)
			tui.NewModel(runner).Draw()
		},
	}
}

func NewExtensionCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new extension",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			extensionID, err := ExtensionID(args[0])
			if err != nil {
				exitWithErrorMsg("Invalid extension name: %s", err)
			}

			cwd, _ := os.Getwd()
			extensionDir := path.Join(cwd, extensionID)
			if _, err := os.Stat(extensionDir); !os.IsNotExist(err) {
				exitWithErrorMsg("Extension already exists: %s", extensionDir)
			}

			if err := os.MkdirAll(extensionDir, 0755); err != nil {
				exitWithErrorMsg("Could not create extension directory: %s", err)
			}

			extensionScriptPath := path.Join(extensionDir, extensionID)
			if err := os.WriteFile(extensionScriptPath, extensionTemplate, 0755); err != nil {
				exitWithErrorMsg("Could not write extension script: %s", err)
			}

			fmt.Println("Extension created successfully!")
		},
	}
}

func NewExtensionExecCmd(extensionDir string, validator tui.PageValidator) *cobra.Command {
	return &cobra.Command{
		Use:   "exec",
		Short: "Execute an installed extension",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			extensionId, err := ExtensionID(args[0])
			if err != nil {
				exitWithErrorMsg("Invalid extension: %s", err)
			}

			binPath := path.Join(extensionDir, extensionId, extensionId)

			if _, err := os.Stat(binPath); os.IsNotExist(err) {
				exitWithErrorMsg("Extension not found: %s", extensionId)
			}

			extraArgs := []string{}
			if len(args) > 1 {
				extraArgs = args[1:]
			}

			cwd, _ := os.Getwd()
			generator := tui.NewCommandGenerator(binPath, extraArgs, cwd)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			runner := tui.NewRunner(generator, validator, cwd)

			tui.NewModel(runner).Draw()
		},
	}
}

func NewExtensionInstallCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install a sunbeam extension from a repository",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			if args[0] == "." {
				cwd, _ := os.Getwd()
				extensionName := path.Base(cwd)

				if !strings.HasPrefix(extensionName, "sunbeam-") {
					exitWithErrorMsg("Extension directory must be named sunbeam-<name> (e.g. sunbeam-foo)")
				}

				bin := path.Join(cwd, extensionName)
				if _, err := os.Stat(bin); os.IsNotExist(err) {
					exitWithErrorMsg("Extension binary not found: %s", bin)
				}

				return
			}

			_, err := url.Parse(args[0])
			if err != nil {
				exitWithErrorMsg("Invalid extension URL: %s", args[0])
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if args[0] == "." {
				cwd, _ := os.Getwd()
				targetDir := path.Join(extensionDir, path.Base(cwd))
				err := os.Symlink(cwd, targetDir)
				if err != nil {
					exitWithErrorMsg("Unable to install local extension: %s", err)
				}

				fmt.Sprintln("Extension installed:", path.Base(cwd))
				return
			}

			repository, err := utils.RepositoryFromString(args[0])
			if err != nil {
				exitWithErrorMsg("Unable to parse repository: %s", err)
			}

			if !strings.HasPrefix(repository.Name(), "sunbeam-") {
				exitWithErrorMsg("Extension name must be prefixed with sunbeam- (e.g. sunbeam-foo)")
			}

			targetDir := path.Join(extensionDir, repository.Name())
			if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
				exitWithErrorMsg("Extension already installed: %s", repository.Name())
			}

			tempDir, err := os.MkdirTemp("", "sunbeam-*")
			if err != nil {
				exitWithErrorMsg("Unable to install extension: %s", err)
			}
			defer os.RemoveAll(tempDir)

			if err = utils.GitClone(repository, tempDir); err != nil {
				exitWithErrorMsg("Unable to install extension: %s", err)
			}

			binPath := path.Join(tempDir, repository.Name())
			if os.Stat(binPath); os.IsNotExist(err) {
				exitWithErrorMsg("Extension binary not found: %s", binPath)
			}

			if err := cp.Copy(tempDir, targetDir); err != nil {
				exitWithErrorMsg("Unable to install extension: %s", err)
			}

			fmt.Println("Extension installed:", repository.Name())
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
			extensionId, err := ExtensionID(args[0])
			if err != nil {
				exitWithErrorMsg("Invalid extension: %s", err)
			}

			targetDir := path.Join(extensionDir, extensionId)
			if _, err := os.Stat(targetDir); os.IsNotExist(err) {
				exitWithErrorMsg("Extension not installed: %s", extensionId)
			}

			if err := os.RemoveAll(targetDir); err != nil {
				exitWithErrorMsg("Unable to remove extension: %s", err)
			}

			fmt.Println("Extension removed:", extensionId)
		},
	}
}

func NewExtensionUpgradeCmd(extensionDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade an installed extension",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			extensionId, err := ExtensionID(args[0])
			if err != nil {
				exitWithErrorMsg("Invalid extension: %s", err)
			}

			extensionPath := path.Join(extensionDir, extensionId)
			if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
				exitWithErrorMsg("Extension not installed: %s", args[0])
			}

			if err := utils.GitPull(extensionPath); err != nil {
				exitWithErrorMsg("Unable to upgrade extension: %s", err)
			}

			fmt.Sprintln("Extension upgraded:", args[0])
		},
	}
}

func ExtensionID(input string) (string, error) {
	tokens := strings.Split(input, "/")
	if len(tokens) == 1 {
		if strings.HasPrefix(input, "sunbeam-") {
			return input, nil
		}
		return fmt.Sprintf("sunbeam-%s", input), nil
	} else if len(tokens) == 2 {
		return tokens[1], nil
	} else {
		return "", fmt.Errorf("invalid extension ID: %s", input)
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
		if !strings.HasPrefix(entry.Name(), "sunbeam-") {
			continue
		}

		binary := path.Join(extensionDir, entry.Name(), entry.Name())
		if _, err := os.Stat(binary); os.IsNotExist(err) {
			continue
		}

		extensionId := strings.TrimPrefix(entry.Name(), "sunbeam-")
		extensions = append(extensions, extensionId)
	}

	return extensions, nil
}
