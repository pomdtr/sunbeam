package cmd

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
)

func NewCmdUpgrade(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "upgrade",
		Short:     "Upgrade extensions",
		GroupID:   CommandGroupCore,
		Args:      cobra.MatchAll(cobra.OnlyValidArgs, cobra.MaximumNArgs(1)),
		ValidArgs: cfg.Aliases(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if _, ok := cfg.Extensions[args[0]]; !ok {
					return fmt.Errorf("unknown extension: %s", args[0])
				}
			}

			for alias, ref := range cfg.Extensions {
				if len(args) > 0 && alias != args[0] {
					continue
				}

				extension, err := ExtractManifest(ref.Origin)
				if err != nil {
					return err
				}

				if _, err := cacheExtension(alias, extension); err != nil {
					return err
				}

				cmd.Printf("✅ Upgraded %s\n", alias)
			}

			return nil
		},
	}

	return cmd
}

func LoadExtension(alias string, origin string) (extensions.Extension, error) {
	cacheDir := filepath.Join(utils.CacheHome(), "extensions")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return extensions.Extension{}, err
	}

	cachePath := filepath.Join(cacheDir, alias+".json")
	if _, err := os.Stat(cachePath); err == nil {
		// check if cache is valid
		var extension extensions.Extension
		extensionBytes, err := os.ReadFile(cachePath)
		if err != nil {
			return extensions.Extension{}, err
		}

		if err := json.Unmarshal(extensionBytes, &extension); err != nil {
			return extensions.Extension{}, err
		}

		entrypointInfo, err := os.Stat(extension.Metadata.Entrypoint)
		if err != nil {
			return extensions.Extension{}, err
		}

		cacheInfo, err := os.Stat(cachePath)
		if err != nil {
			return extensions.Extension{}, err
		}

		if extension.Metadata.Origin == origin && entrypointInfo.ModTime().Before(cacheInfo.ModTime()) {
			return extension, nil
		}
	}

	extension, err := ExtractManifest(origin)
	if err != nil {
		return extensions.Extension{}, err
	}

	return cacheExtension(alias, extension)
}

func cacheExtension(alias string, extension extensions.Extension) (extensions.Extension, error) {
	extensionBytes, err := json.MarshalIndent(extension, "", "  ")
	if err != nil {
		return extensions.Extension{}, err
	}

	cacheDir := filepath.Join(utils.CacheHome(), "extensions")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return extensions.Extension{}, err
	}

	cachePath := filepath.Join(cacheDir, alias+".json")

	if err := os.WriteFile(cachePath, extensionBytes, 0644); err != nil {
		return extensions.Extension{}, err
	}

	return extension, nil
}

func ExtractManifest(origin string) (extensions.Extension, error) {
	var entrypoint string
	var extensionType extensions.ExtensionType
	if strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://") {
		extensionType = extensions.ExtensionTypeHttp
		resp, err := http.Get(origin)
		if err != nil {
			return extensions.Extension{}, err
		}

		h := sha1.New()
		h.Write([]byte(origin))
		sha1_hash := hex.EncodeToString(h.Sum(nil))

		cacheDir := filepath.Join(utils.CacheHome(), "scripts")
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return extensions.Extension{}, err
		}

		entrypoint = filepath.Join(utils.CacheHome(), "scripts", sha1_hash+filepath.Ext(origin))
		f, err := os.Create(entrypoint)
		if err != nil {
			return extensions.Extension{}, err
		}
		defer f.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			return extensions.Extension{}, err
		}

		if err := f.Close(); err != nil {
			return extensions.Extension{}, err
		}
	} else {
		extensionType = extensions.ExtensionTypeLocal
		entrypoint = origin
		if strings.HasPrefix(entrypoint, "~") {
			homedir, err := os.UserHomeDir()
			if err != nil {
				return extensions.Extension{}, err
			}

			entrypoint = filepath.Join(homedir, entrypoint[1:])
		}
	}

	if err := os.Chmod(entrypoint, 0755); err != nil {
		return extensions.Extension{}, err
	}

	var args []string
	if runtime.GOOS == "windows" {
		args = []string{"sunbeam", "shell", entrypoint}
	} else {
		args = []string{entrypoint}
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = filepath.Dir(entrypoint)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SUNBEAM=1")

	manifestBytes, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return extensions.Extension{}, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
		}

		return extensions.Extension{}, err
	}

	if err := schemas.ValidateManifest(manifestBytes); err != nil {
		return extensions.Extension{}, err
	}

	var manifest types.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return extensions.Extension{}, err
	}

	extension := extensions.Extension{
		Manifest: manifest,
		Metadata: extensions.Metadata{
			Type:       extensionType,
			Origin:     origin,
			Entrypoint: entrypoint,
		},
	}

	return extension, nil
}
