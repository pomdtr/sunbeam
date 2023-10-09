package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

func NewCmdExtension() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension",
		Aliases: []string{"ext", "extensions"},
		Short:   "Manage extensions",
		GroupID: CommandGroupCore,
	}

	cmd.AddCommand(NewCmdExtensionList())
	cmd.AddCommand(NewCmdExtensionInstall())
	cmd.AddCommand(NewCmdExtensionUpgrade())
	cmd.AddCommand(NewCmdExtensionRemove())
	cmd.AddCommand(NewCmdExtensionRename())
	cmd.AddCommand(NewCmdExtensionBrowse())

	return cmd
}

func NewCmdExtensionBrowse() *cobra.Command {
	return &cobra.Command{
		Use:   "browse",
		Short: "Browse extensions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return utils.Open("https://github.com/topics/sunbeam-extension")
		},
	}
}

func NewCmdExtensionList() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List installed extensions",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions, err := FindExtensions()
			if err != nil {
				return err
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				extensionMap := make(map[string]types.Manifest)
				for alias, extension := range extensions {
					extensionMap[alias] = extension.Manifest
				}

				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(extensionMap)
			}

			for alias, extension := range extensions {
				fmt.Printf("%s\t%s\t%s\n", alias, extension.Title, extension.Origin)
			}

			return nil
		},
	}
}

func NewCmdExtensionInstall() *cobra.Command {
	flags := struct {
		alias string
	}{}

	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"add"},
		Short:   "Install an extension",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			origin, err := url.Parse(args[0])
			if err != nil {
				return err
			}

			if origin.Scheme == "file" || origin.Scheme == "" {
				p, err := filepath.Abs(origin.Path)
				if err != nil {
					return err
				}

				origin.Path = p
			}

			extensionRoot := filepath.Join(dataHome(), "extensions")
			if err := os.MkdirAll(extensionRoot, 0755); err != nil {
				return err
			}

			var alias string
			if flags.alias != "" {
				alias = flags.alias
			} else {
				alias = filepath.Base(origin.Path)
				alias = strings.TrimSuffix(alias, filepath.Ext(alias))
				alias = strings.TrimPrefix(alias, "sunbeam-")
			}

			return installExtension(origin, filepath.Join(extensionRoot, alias))
		},
	}

	cmd.Flags().StringVarP(&flags.alias, "alias", "a", "", "alias for extension")
	return cmd
}

func installExtension(origin *url.URL, extensionDir string) (err error) {
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(extensionDir)
		}
	}()

	srcDir := filepath.Join(extensionDir, "src")
	var version string
	if origin.Scheme == "file" || origin.Scheme == "" {
		if err := installFromLocalDir(origin.Path, srcDir); err != nil {
			return err
		}
		version = time.Now().Format("20060102150405")
	} else if release, err := getLatestRelease(origin); err == nil {
		if err := installFromGithubRelease(release, srcDir); err != nil {
			return err
		}

		version = release.TagName
	} else if tag, err := getLatestTag(origin); err == nil {
		if err := installFromRepository(origin, tag, srcDir); err != nil {
			return err
		}

		version = tag
	} else {
		if err := installFromRepository(origin, "", srcDir); err != nil {
			return err
		}

		version = ""
	}

	entrypoint := filepath.Join(srcDir, "sunbeam-extension")
	if _, err := os.Stat(entrypoint); err != nil {
		return fmt.Errorf("extension %s not found", origin.String())
	}

	extension, err := tui.LoadExtension(entrypoint)
	if err != nil {
		return err
	}

	extension.Origin = origin.String()
	extension.Version = version

	f, err := os.Create(filepath.Join(extensionDir, "manifest.json"))
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(extension.Manifest); err != nil {
		return err
	}

	return nil
}

func installFromLocalDir(srcDir string, targetDir string) error {
	originDir, err := filepath.Abs(srcDir)
	if err != nil {
		return err
	}
	entrypoint := filepath.Join(originDir, "sunbeam-extension")
	if _, err := os.Stat(entrypoint); err != nil {
		return fmt.Errorf("extension %s not found", srcDir)
	}

	return os.Symlink(srcDir, targetDir)
}

func installFromRepository(origin *url.URL, tag string, targetDir string) error {
	// check if git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	var cloneCmd *exec.Cmd
	if tag != "" {
		cloneCmd = exec.Command("git", "clone", "--depth=1", fmt.Sprintf("--branch=%s", tag), tag, origin.String(), targetDir)
	} else {
		cloneCmd = exec.Command("git", "clone", "--depth=1", origin.String(), targetDir)
	}
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return err
	}

	return nil
}

func installFromGithubRelease(release GithubRelease, targetDir string) error {
	var assetUrl string
	for _, asset := range release.Assets {
		if asset.Name != fmt.Sprintf("sunbeam-extension-%s-%s", runtime.GOOS, runtime.GOARCH) {
			continue
		}

		assetUrl = asset.URL
		break
	}

	if assetUrl == "" {
		return fmt.Errorf("no asset found")
	}

	resp, err := http.Get(assetUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download extension")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	entrypointPath := filepath.Join(targetDir, "sunbeam-extension")
	entrypoint, err := os.Create(entrypointPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(entrypoint, resp.Body); err != nil {
		return err
	}

	if err := os.Chmod(entrypointPath, 0755); err != nil {
		return err
	}

	return nil
}

func getLatestTag(origin *url.URL) (string, error) {
	cmd := exec.Command("git", "ls-remote", "--tags", origin.String())
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	var tags []string
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}

		tags = append(tags, strings.Split(line, "\t")[1])
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found")
	}

	sort.SliceStable(tags, func(i, j int) bool {
		return semver.Compare(tags[i], tags[j]) == -1
	})

	return tags[len(tags)-1], nil
}

type GithubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func getLatestRelease(origin *url.URL) (GithubRelease, error) {
	var release GithubRelease
	if origin.Host != "github.com" {
		return release, fmt.Errorf("only github.com is supported")
	}

	latestReleaseUrl := fmt.Sprintf("https://api.github.com/repos%s/releases/latest", origin.Path)
	resp, err := http.Get(latestReleaseUrl)
	if err != nil {
		return release, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return release, fmt.Errorf("failed to get latest release")
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return release, err
	}

	return release, nil
}

func NewCmdExtensionUpgrade() *cobra.Command {
	flags := struct {
		all bool
	}{}
	cmd := &cobra.Command{
		Use:     "upgrade",
		Aliases: []string{"update"},
		Short:   "Upgrade an extension",
		Args:    cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if flags.all {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			extensions, err := FindExtensions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for alias := range extensions {
				completions = append(completions, alias)
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.all && len(args) > 0 {
				return fmt.Errorf("cannot use --all with an extension")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// check if git is installed
			toUpgrade := make([]string, 0)
			if flags.all {
				extensions, err := FindExtensions()
				if err != nil {
					return err
				}

				for alias := range extensions {
					toUpgrade = append(toUpgrade, alias)
				}
			} else {
				toUpgrade = append(toUpgrade, args[0])
			}

			for _, alias := range toUpgrade {
				cmd.PrintErrln()
				cmd.PrintErrf("Upgrading %s...\n", alias)
				extensionDir := filepath.Join(dataHome(), "extensions", alias)

				if err := upgradeExtension(extensionDir); err != nil {
					cmd.PrintErrln(err)
				} else {
					cmd.PrintErrln("done")
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&flags.all, "all", "a", false, "upgrade all extensions")

	return cmd
}

func upgradeExtension(extensionDir string) error {
	f, err := os.Open(filepath.Join(extensionDir, "manifest.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	var manifest types.Manifest
	if err := json.NewDecoder(f).Decode(&manifest); err != nil {
		return err
	}
	f.Close()

	origin, err := url.Parse(manifest.Origin)
	if err != nil {
		return err
	}

	var version string
	if origin.Scheme == "file" || origin.Scheme == "" {
		version = time.Now().Format("20060102150405")
	} else if release, err := getLatestRelease(origin); err == nil {
		if manifest.Version == release.TagName {
			return nil
		}

		// remove old entrypoint
		entrypoint := filepath.Join(extensionDir, "src", "sunbeam-extension")
		if err := os.Remove(entrypoint); err != nil {
			return err
		}

		if err := installFromGithubRelease(release, filepath.Join(extensionDir, "src")); err != nil {
			return err
		}

		version = release.TagName
	} else if tag, err := getLatestTag(origin); err == nil {
		if manifest.Version == tag {
			return nil
		}

		checkoutCmd := exec.Command("git", "checkout", tag)
		checkoutCmd.Dir = filepath.Join(extensionDir, "src")
		checkoutCmd.Stdout = os.Stdout
		checkoutCmd.Stderr = os.Stderr
		if err := checkoutCmd.Run(); err != nil {
			return err
		}

		version = tag
	} else {
		pullCmd := exec.Command("git", "pull")
		pullCmd.Dir = filepath.Join(extensionDir, "src")
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr

		if err := pullCmd.Run(); err != nil {
			return err
		}

		version = ""
	}

	// refresh manifest
	extension, err := tui.LoadExtension(filepath.Join(extensionDir, "src", "sunbeam-extension"))
	if err != nil {
		return err
	}
	extension.Origin = origin.String()
	extension.Version = version

	f, err = os.Create(filepath.Join(extensionDir, "manifest.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(extension.Manifest); err != nil {
		return err
	}

	return nil
}

func NewCmdExtensionRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove an extension",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			extensions, err := FindExtensions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for alias := range extensions {
				completions = append(completions, alias)
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionDir := filepath.Join(dataHome(), "extensions", args[0])
			return os.RemoveAll(extensionDir)
		},
	}

	return cmd
}

func NewCmdExtensionRename() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename an extension",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			extensions, err := FindExtensions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for alias := range extensions {
				completions = append(completions, alias)
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionDir := filepath.Join(dataHome(), "extensions", args[0])
			newExtensionDir := filepath.Join(dataHome(), "extensions", args[1])

			return os.Rename(extensionDir, newExtensionDir)
		},
	}

	return cmd
}

func FindExtensions() (map[string]tui.Extension, error) {
	extensionRoot := filepath.Join(dataHome(), "extensions")
	if _, err := os.Stat(extensionRoot); err != nil {
		return nil, nil
	}

	entries, err := os.ReadDir(extensionRoot)
	if err != nil {
		return nil, err
	}
	extensionMap := make(map[string]tui.Extension)
	for _, entry := range entries {
		manifestPath := filepath.Join(extensionRoot, entry.Name(), "manifest.json")
		f, err := os.Open(manifestPath)
		if err != nil {
			continue
		}
		defer f.Close()

		var manifest types.Manifest
		if err := json.NewDecoder(f).Decode(&manifest); err != nil {
			continue
		}

		extensionMap[entry.Name()] = tui.Extension{
			Manifest:   manifest,
			Entrypoint: filepath.Join(extensionRoot, entry.Name(), "src", "sunbeam-extension"),
		}
	}

	return extensionMap, nil
}
