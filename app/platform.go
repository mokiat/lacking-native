package app

import (
	"runtime"

	"github.com/mokiat/lacking/app"
)

func newPlatform() *platform {
	return &platform{
		os: determineOS(),
	}
}

type platform struct {
	os app.OS
}

var _ app.Platform = (*platform)(nil)

func (p *platform) Environment() app.Environment {
	return app.EnvironmentNative
}

func (p *platform) OS() app.OS {
	return p.os
}

func determineOS() app.OS {
	switch runtime.GOOS {
	case "linux":
		return app.OSLinux
	case "windows":
		return app.OSWindows
	case "darwin":
		return app.OSDarwin
	default:
		return app.OSUnknown
	}
}
