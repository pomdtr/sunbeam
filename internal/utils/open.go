package utils

import (
	"fmt"
	"runtime"
	"time"

	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/skratchdot/open-golang/open"
)

func OpenWith(target string, application *types.Application) error {
	if application == nil {
		return Open(target)
	}

	var applicationName string
	switch runtime.GOOS {
	case "darwin":
		applicationName = application.Macos
	case "linux":
		applicationName = application.Linux
	default:
		return fmt.Errorf("unsupported platform")
	}

	if err := open.RunWith(target, applicationName); err != nil {
		return err
	}

	// hack: wait for the application to open
	time.Sleep(100 * time.Millisecond)

	return nil
}

func Open(target string) error {
	if err := open.Run(target); err != nil {
		return err
	}

	// hack: wait for the application to open
	time.Sleep(100 * time.Millisecond)

	return nil
}
