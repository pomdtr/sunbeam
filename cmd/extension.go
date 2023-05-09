package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"embed"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leaanthony/gosod"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"

	cp "github.com/otiai10/copy"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

var (
	extensionBinaryName = "sunbeam-extension"
	manifestName        = "manifest.json"
)

//go:embed templates/bash/*
var bashTemplate embed.FS

//go:embed templates/go/*
var golangTemplate embed.FS

//go:embed templates/deno/*
var denoTemplate embed.FS

var templates = map[string]embed.FS{
	"bash": bashTemplate,
	"go":   golangTemplate,
	"deno": denoTemplate,
}

type ExtensionType string

const (
	ExtensionTypeBinary ExtensionType = "binary"
	ExtensionTypeGit    ExtensionType = "git"
	ExtensionTypeGist   ExtensionType = "gist"
	ExtentionTypeLocal  ExtensionType = "local"
)

type ExtensionManifest struct {
	Type        ExtensionType `json:"type"`
	Entrypoint  string        `json:"entrypoint"`
	Description string        `json:"description"`
	Remote      string        `json:"remote,omitempty"`
	Version     string        `json:"version,omitempty"`
}

func ReadManifest(manifestPath string) (*ExtensionManifest, error) {
	bytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load extension manifest: %s", err)
	}

	var manifest ExtensionManifest
	if err := json.Unmarshal(bytes, &manifest); err != nil {
		return nil, fmt.Errorf("unable to load extension manifest: %s", err)
	}

	return &manifest, nil
}

func (m *ExtensionManifest) Write(manifestPath string) error {
	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to write extension manifest: %s", err)
	}

	if err := os.WriteFile(manifestPath, bytes, 0644); err != nil {
		return fmt.Errorf("unable to write extension manifest: %s", err)
	}

	return nil
}

func NewExtensionCmd(extensionRoot string, extensions map[string]*ExtensionManifest) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:     "extension",
		Short:   "Extension commands",
		GroupID: coreGroupID,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(extensionRoot); os.IsNotExist(err) {
				os.MkdirAll(extensionRoot, 0755)
			}
			return nil
		},
	}

	extensionCmd.AddCommand(NewExtensionBrowseCmd(extensionRoot))
	extensionCmd.AddCommand(NewExtensionManageCmd(extensionRoot))
	extensionCmd.AddCommand(NewExtensionCreateCmd())
	extensionCmd.AddCommand(NewExtensionInstallCmd(extensionRoot))
	extensionCmd.AddCommand(NewExtensionRenameCmd(extensionRoot, extensions))
	extensionCmd.AddCommand(NewExtensionListCmd(extensionRoot, extensions))
	extensionCmd.AddCommand(NewExtensionRemoveCmd(extensionRoot, extensions))
	extensionCmd.AddCommand(NewExtensionUpgradeCmd(extensionRoot, extensions))

	return extensionCmd
}

func NewExtensionBrowseCmd(extensionRoot string) *cobra.Command {
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

			return Run(generator)
		},
	}
}

func NewExtensionManageCmd(extensionRoot string) *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Manage installed extensions",

		RunE: func(cmd *cobra.Command, args []string) error {
			generator := func() (*types.Page, error) {
				extensions, err := ListExtensions(extensionRoot)
				if err != nil {
					return nil, fmt.Errorf("unable to list extensions: %s", err)
				}

				listItems := make([]types.ListItem, 0)
				for extension, manifest := range extensions {
					listItems = append(listItems, types.ListItem{
						Title:       extension,
						Subtitle:    manifest.Description,
						Accessories: []string{manifest.Version},
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
										Default:     extension,
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
					Type: types.ListPage,
					EmptyView: &types.EmptyView{
						Text: "No extensions installed",
						Actions: []types.Action{
							{
								Type:  types.PushAction,
								Title: "Browse Extensions",
								Command: &types.Command{
									Name: os.Args[0],
									Args: []string{"extension", "browse"},
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

func NewExtensionCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new extension",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			language, _ := cmd.Flags().GetString("language")

			var missingInputs []types.Input
			createArgs := []string{"extension", "create"}

			if !cmd.Flags().Changed("name") {
				missingInputs = append(missingInputs, types.Input{Type: types.TextFieldInput, Name: "name", Title: "Extension Name", Placeholder: "my-extension"})
			} else {
				createArgs = append(createArgs, fmt.Sprintf("--name=%s", name))
			}

			if !cmd.Flags().Changed("language") {
				missingInputs = append(missingInputs, types.Input{
					Type:  types.DropDownInput,
					Name:  "language",
					Title: "Language",
					Items: []types.DropDownItem{
						{Title: "Bash", Value: "bash"},
						{Title: "Deno", Value: "deno"},
						{Title: "Go", Value: "go"},
					}})
			} else {
				createArgs = append(createArgs, fmt.Sprintf("--language=%s", language))
			}

			if len(missingInputs) > 0 {
				form := internal.NewForm(
					"Create Extension",
					func(values map[string]string) tea.Msg {
						for name, value := range values {
							createArgs = append(createArgs, fmt.Sprintf("--%s=%s", name, value))
						}

						return internal.ExitMsg{
							Cmd: exec.Command(os.Args[0], createArgs...),
						}
					},
					missingInputs...)

				return internal.Draw(form)
			}

			template, ok := templates[language]
			if !ok {
				return fmt.Errorf("unknown template: %s", language)
			}

			sub, err := fs.Sub(template, fmt.Sprint("templates/", language))
			if err != nil {
				return fmt.Errorf("unable to load template: %s", err)
			}

			templateDir := gosod.New(sub)
			templateDir.SetTemplateFilters([]string{".tmpl"})
			if err := templateDir.Extract(name, nil); err != nil {
				return fmt.Errorf("unable to extract template: %s", err)
			}

			fmt.Printf("Created extension %s\n", name)
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "extension name")
	cmd.Flags().StringP("language", "l", "", "extension language")

	return cmd
}

func NewExtensionRenameCmd(extensionRoot string, extensions map[string]*ExtensionManifest) *cobra.Command {
	validArgs := make([]string, 0, len(extensions))
	for extension := range extensions {
		validArgs = append(validArgs, extension)
	}

	return &cobra.Command{
		Use:       "rename [old] [new]",
		Short:     "Rename an extension",
		Args:      cobra.ExactArgs(2),
		ValidArgs: validArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			commands := cmd.Root().Commands()
			commandMap := make(map[string]struct{}, len(commands))
			for _, command := range commands {
				commandMap[command.Name()] = struct{}{}
			}

			if _, ok := commandMap[args[0]]; !ok {
				return fmt.Errorf("extension does not exist: %s", args[0])
			}

			if _, ok := commandMap[args[1]]; ok {
				return fmt.Errorf("alias conflicts with existing command: %s", args[0])
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			oldPath := filepath.Join(extensionRoot, args[0])
			newPath := filepath.Join(extensionRoot, args[1])

			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("could not rename extension: %s", err)
			}

			return nil
		},
	}
}

func NewExtensionInstallCmd(extensionRoot string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [alias] [extension]",
		Short: "Install a sunbeam extension from a folder/gist/repository",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			commands := cmd.Root().Commands()
			commandMap := make(map[string]struct{}, len(commands))
			for _, command := range commands {
				commandMap[command.Name()] = struct{}{}
			}

			if _, ok := commandMap[args[0]]; ok {
				return fmt.Errorf("alias conflicts with existing command: %s", args[0])
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := filepath.Join(extensionRoot, args[0])

			if err := installExtension(args[1], targetDir); err != nil {
				return fmt.Errorf("could not install extension: %s", err)
			}

			open, _ := cmd.Flags().GetBool("open")
			if open {
				return Run(internal.NewCommandGenerator(&types.Command{
					Name: os.Args[0],
					Args: []string{args[0]},
				}))
			} else {
				fmt.Printf("Installed extension %s\n", args[0])
			}

			return nil
		},
	}

	cmd.Flags().BoolP("open", "o", false, "open extension after installation")

	return cmd
}

func installExtension(origin string, targetDir string) error {
	if origin == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current working directory: %s", err)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not get home directory: %s", err)
		}

		entrypoint := filepath.Join(cwd, extensionBinaryName)
		if _, err := os.Stat(entrypoint); os.IsNotExist(err) {
			return fmt.Errorf("no extension found in current directory")
		}

		if err := os.Chmod(entrypoint, 0755); err != nil {
			return fmt.Errorf("could not make extension executable: %s", err)
		}

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("could not create extension directory: %s", err)
		}

		manifest := ExtensionManifest{
			Type:        ExtentionTypeLocal,
			Entrypoint:  entrypoint,
			Description: strings.Replace(cwd, homeDir, "~", 1),
		}

		if err := manifest.Write(filepath.Join(targetDir, manifestName)); err != nil {
			return fmt.Errorf("unable to write extension manifest: %s", err)
		}

		return nil
	}

	extensionUrl, err := url.Parse(origin)
	if err != nil {
		return fmt.Errorf("invalid extension url: %s", err)
	}

	if extensionUrl.Host == "gist.github.com" {
		gist, err := utils.FetchGithubGist(origin)
		if err != nil {
			return fmt.Errorf("unable to install extension: %s", err)
		}

		commit, err := utils.GetLastGistCommit(origin)
		if err != nil {
			return fmt.Errorf("unable to install extension: %s", err)
		}

		zipUrl := fmt.Sprintf("%s/archive/%s.zip", origin, commit.Version)
		if err := downloadAndExtractZip(zipUrl, filepath.Join(targetDir, "src")); err != nil {
			return fmt.Errorf("unable to install extension: %s", err)
		}

		manifest := ExtensionManifest{
			Type:        ExtensionTypeGist,
			Remote:      origin,
			Description: gist.Description,
			Version:     commit.Version,
			Entrypoint:  filepath.Join("src", "sunbeam-extension"),
		}

		if err := manifest.Write(filepath.Join(targetDir, manifestName)); err != nil {
			return fmt.Errorf("unable to write extension manifest: %s", err)
		}

		return nil
	}

	repository, err := utils.FetchGithubRepository(origin)
	if err != nil {
		return fmt.Errorf("could not fetch extension metadata: %s", err)
	}

	if release, err := utils.GetLatestRelease(origin); err == nil {
		binaryName, err := downloadRelease(release, targetDir)
		if err != nil {
			return fmt.Errorf("unable to install extension: %s", err)
		}

		manifest := ExtensionManifest{
			Type:        ExtensionTypeBinary,
			Remote:      origin,
			Description: repository.Description,
			Entrypoint:  binaryName,
			Version:     release.TagName,
		}

		if err := manifest.Write(filepath.Join(targetDir, manifestName)); err != nil {
			return fmt.Errorf("unable to write extension manifest: %s", err)
		}

		return nil
	}

	commit, err := utils.GetLastGitCommit(origin)
	if err != nil {
		return fmt.Errorf("could not fetch extension metadata: %s", err)
	}

	if err := downloadAndExtractZip(fmt.Sprintf("%s/archive/%s.zip", extensionUrl, commit.Sha), filepath.Join(targetDir, "src")); err != nil {
		return fmt.Errorf("unable to download extension: %s", err)
	}

	manifest := ExtensionManifest{
		Type:        ExtensionTypeGit,
		Remote:      origin,
		Version:     commit.Sha,
		Description: repository.Description,
		Entrypoint:  filepath.Join("src", extensionBinaryName),
	}

	if err := manifest.Write(filepath.Join(targetDir, manifestName)); err != nil {
		return fmt.Errorf("unable to write extension manifest: %s", err)
	}

	return nil
}

func downloadRelease(release *utils.Release, targetDir string) (string, error) {
	downloadUrl := fmt.Sprintf("https://github.com/pomdtr/sunbeam-vscode/releases/download/%s/sunbeam-extension-%s-%s", release.TagName, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		downloadUrl += ".exe"
	}
	res, err := http.Get(downloadUrl)
	if err != nil {
		return "", fmt.Errorf("unable to install extension: %s", err)
	}
	defer res.Body.Close()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("unable to install extension: %s", err)
	}

	binaryName := extensionBinaryName
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	out, err := os.OpenFile(filepath.Join(targetDir, binaryName), os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return "", fmt.Errorf("unable to install extension: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return binaryName, err
}

func NewExtensionListCmd(extensionRoot string, extensions map[string]*ExtensionManifest) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed extension commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			delimiter, _ := cmd.Flags().GetString("delimiter")
			for extension, manifest := range extensions {
				var version string
				if len(manifest.Version) > 8 {
					version = manifest.Version[:8]
				} else {
					version = manifest.Version
				}
				fmt.Printf("%s%s%s%s%s\n", extension, delimiter, manifest.Description, delimiter, version)
			}

			return nil
		},
	}

	cmd.Flags().StringP("delimiter", "d", "\t", "delimiter to use between extension name and description")

	return cmd
}

func NewExtensionRemoveCmd(extensionRoot string, extensions map[string]*ExtensionManifest) *cobra.Command {
	validArgs := make([]string, 0, len(extensions))
	for extension := range extensions {
		validArgs = append(validArgs, extension)
	}
	return &cobra.Command{
		Use:       "remove",
		Short:     "Remove an installed extension",
		Args:      cobra.ExactArgs(1),
		ValidArgs: validArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := filepath.Join(extensionRoot, args[0])
			if _, err := os.Stat(targetDir); os.IsNotExist(err) {
				return fmt.Errorf("extension %s not installed", args[0])
			}

			if err := os.RemoveAll(targetDir); err != nil {
				return fmt.Errorf("unable to remove extension: %s", err)
			}

			fmt.Printf("Extension %s removed\n", args[0])
			return nil
		},
	}
}

func NewExtensionUpgradeCmd(extensionRoot string, extensions map[string]*ExtensionManifest) *cobra.Command {
	validArgs := make([]string, 0, len(extensions))
	for extension := range extensions {
		validArgs = append(validArgs, extension)
	}

	return &cobra.Command{
		Use:       "upgrade",
		Short:     "Upgrade an installed extension",
		Args:      cobra.ExactArgs(1),
		ValidArgs: validArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionName := args[0]
			extensionPath := filepath.Join(extensionRoot, extensionName)
			if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
				return fmt.Errorf("extension not installed: %s", args[0])
			}

			if err := upgradeExtension(extensionPath); err != nil {
				return fmt.Errorf("unable to upgrade extension: %s", err)
			}

			fmt.Sprintln("Extension upgraded:", args[0])
			return nil

		},
	}
}

func upgradeExtension(extensionPath string) error {

	manifestPath := filepath.Join(extensionPath, manifestName)
	manifest, err := ReadManifest(filepath.Join(extensionPath, manifestName))
	if err != nil {
		return fmt.Errorf("unable to upgrade extension: %s", err)
	}

	switch manifest.Type {
	case ExtensionTypeBinary:
		release, err := utils.GetLatestRelease(manifest.Remote)
		if err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}

		if release.TagName == manifest.Version {
			fmt.Println("Extension already up to date")
			return nil
		}

		if _, err := downloadRelease(release, extensionPath); err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}

		manifest.Version = release.TagName
		if err := manifest.Write(manifestPath); err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}

		return nil
	case ExtensionTypeGit:
		commit, err := utils.GetLastGitCommit(manifest.Remote)
		if err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}

		if commit.Sha == manifest.Version {
			fmt.Println("Extension already up to date")
			return nil
		}

		tempdir, err := os.MkdirTemp("", "sunbeam-*")
		if err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}
		defer os.RemoveAll(tempdir)

		zipUrl := fmt.Sprintf("%s/archive/%s.zip", manifest.Remote, commit.Sha)
		if err := downloadAndExtractZip(zipUrl, tempdir); err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}

		srcDir := filepath.Join(extensionPath, "src")
		if err := os.RemoveAll(srcDir); err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}

		if err := cp.Copy(tempdir, srcDir); err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}
	case ExtensionTypeGist:
		commit, err := utils.GetLastGistCommit(manifest.Remote)
		if err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}

		if commit.Version == manifest.Version {
			fmt.Println("Extension already up to date")
			return nil
		}

		tempdir, err := os.MkdirTemp("", "sunbeam-*")
		if err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}
		defer os.RemoveAll(tempdir)

		zipUrl := fmt.Sprintf("%s/archive/%s.zip", manifest.Remote, commit.Version)
		if err := downloadAndExtractZip(zipUrl, tempdir); err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}

		srcDir := filepath.Join(extensionPath, "src")
		if err := os.RemoveAll(srcDir); err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}

		if err := cp.Copy(tempdir, srcDir); err != nil {
			return fmt.Errorf("unable to upgrade extension: %s", err)
		}
	case ExtentionTypeLocal:
		return fmt.Errorf("upgrade not supported for local extensions")
	}

	return nil
}

func downloadAndExtractZip(zipUrl string, dst string) error {
	res, err := http.Get(zipUrl)
	if err != nil {
		return fmt.Errorf("unable to download extension: %s", err)
	}
	defer res.Body.Close()

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to download extension: %s", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(bs), int64(len(bs)))
	if err != nil {
		return fmt.Errorf("unable to download extension: %s", err)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	rootDir := zipReader.File[0].Name
	for _, file := range zipReader.File[1:] {
		fpath := filepath.Join(dst, strings.TrimPrefix(file.Name, rootDir))

		if file.FileInfo().IsDir() {
			os.MkdirAll(fpath, file.Mode())
			continue
		}

		mode := file.Mode()
		if filepath.Base(fpath) == "sunbeam-extension" {
			mode = 0755
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE, mode)
		if err != nil {
			return err
		}
		defer outFile.Close()

		// Copy the file contents
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

func ListExtensions(extensionRoot string) (map[string]*ExtensionManifest, error) {
	if _, err := os.Stat(extensionRoot); os.IsNotExist(err) {
		return nil, nil
	}

	extensionDirs, err := os.ReadDir(extensionRoot)
	if err != nil {
		return nil, fmt.Errorf("unable to list extensions: %s", err)
	}

	extensions := make(map[string]*ExtensionManifest)
	for _, extensionDir := range extensionDirs {
		if !extensionDir.IsDir() {
			continue
		}
		manifestPath := filepath.Join(extensionRoot, extensionDir.Name(), manifestName)
		if _, err := os.Stat(filepath.Join(extensionRoot, extensionDir.Name(), manifestName)); os.IsNotExist(err) {
			return nil, fmt.Errorf("unable to list extensions: %s", err)
		}

		bytes, err := os.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("unable to list extensions: %s", err)
		}

		var manifest ExtensionManifest
		if err := json.Unmarshal(bytes, &manifest); err != nil {
			return nil, fmt.Errorf("unable to list extensions: %s", err)
		}

		extensions[extensionDir.Name()] = &manifest
	}

	return extensions, nil
}
