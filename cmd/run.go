package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/spf13/cobra"
)

func NewCmdRun(config tui.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <origin> [args...]",
		Short:              "Run an extension without installing it",
		Args:               cobra.MinimumNArgs(1),
		GroupID:            coreGroupID,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			origin, err := resolveSpecifiers(args[0])
			if err != nil {
				return err
			}

			rootCmd, err := NewCustomCommand(origin, config)
			if err != nil {
				return err
			}

			rootCmd.Use = args[0]
			rootCmd.SetArgs(args[1:])
			return rootCmd.Execute()
		},
	}

	return cmd
}

var (
	githubRegexp = regexp.MustCompile("^github:([A-Za-z0-9_-]+/[A-Za-z0-9_-]+)@([A-Za-z0-9_-]+)/(.+)$")
	gistRegexp   = regexp.MustCompile("^gist:([A-Za-z0-9_-]+/[A-Za-z0-9_-]+)(?:@([A-Za-z0-9_-]+))?/(.+)$")
)

func downloadToTempfile(rawUrl string) (string, error) {
	res, err := http.Get(rawUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch extension manifest: %s", res.Status)
	}

	f, err := os.CreateTemp("", "sunbeam-")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(f, res.Body); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	// chmod +x
	if err := os.Chmod(f.Name(), 0755); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func resolveSpecifiers(rawOrigin string) (string, error) {
	// check if origin match github regexp
	if matches := githubRegexp.FindStringSubmatch(rawOrigin); len(matches) > 0 {
		repo, rev, path := matches[1], matches[2], matches[3]
		rawUrl := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repo, rev, path)

		scriptPath, err := downloadToTempfile(rawUrl)
		if err != nil {
			return "", err
		}
		rawOrigin = fmt.Sprintf("file://%s", scriptPath)
	} else if matches := gistRegexp.FindStringSubmatch(rawOrigin); len(matches) > 0 {
		gist, rev, path := matches[1], matches[2], matches[3]
		var rawUrl string
		if rev != "" {
			rawUrl = fmt.Sprintf("https://gist.githubusercontent.com/%s/raw/%s/%s", gist, rev, path)
		} else {
			rawUrl = fmt.Sprintf("https://gist.githubusercontent.com/%s/raw/%s", gist, path)
		}

		scriptPath, err := downloadToTempfile(rawUrl)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("file://%s", scriptPath), nil
	}

	return rawOrigin, nil
}
