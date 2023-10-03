package utils

import (
	"fmt"
	"runtime"
	"time"

	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/skratchdot/open-golang/open"
)

func OpenWith(target string, application types.Applications) error {
	if len(application) == 0 {
		return open.Run(target)
	}

	var platform string
	switch runtime.GOOS {
	case "windows":
		platform = types.PlatformWindows
	case "darwin":
		platform = types.PlatformMac
	case "linux":
		platform = types.PlatformLinux
	default:
		return fmt.Errorf("unsupported platform")
	}

	for _, app := range application {
		if app.Platform != "" && app.Platform != types.Platform(platform) {
			continue
		}
		if err := open.RunWith(target, app.Name); err != nil {
			return err
		}

		// hack: wait for the application to open
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	return fmt.Errorf("no application found")
}

func Open(target string) error {
	return OpenWith(target, nil)
}
