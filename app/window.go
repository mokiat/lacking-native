package app

import (
	"fmt"
	"image"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/constr"
	"github.com/mokiat/lacking/app"
	_ "github.com/mokiat/lacking/debug/log" // for side effects
)

// Run starts a new application and opens a single window.
//
// The specified configuration is used to determine how the
// window is initialized.
//
// The specified controller will be used to send notifications
// on window state changes.
func Run(cfg *Config, controller app.Controller) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := glfw.Init(); err != nil {
		return fmt.Errorf("failed to initialize glfw: %w", err)
	}
	defer glfw.Terminate()

	monitor := glfw.GetPrimaryMonitor()

	var windowWidth, windowHeight int
	if cfg.fullscreen {
		videoMode := monitor.GetVideoMode()
		windowWidth = videoMode.Width
		windowHeight = videoMode.Height
	} else {
		scaleX, scaleY := monitor.GetContentScale()
		windowWidth = scaled(cfg.width, scaleX)
		windowHeight = scaled(cfg.height, scaleY)
	}

	glfw.WindowHint(glfw.ScaleToMonitor, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.SRGBCapable, glfw.True)
	if cfg.maximized {
		glfw.WindowHint(glfw.Maximized, glfw.True)
	}

	windowMonitor := gog.Ternary(cfg.fullscreen, monitor, nil)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, cfg.title, windowMonitor, nil)
	if err != nil {
		return fmt.Errorf("failed to create glfw window: %w", err)
	}
	defer window.Destroy()

	if cfg.minWidth != nil || cfg.maxWidth != nil || cfg.minHeight != nil || cfg.maxHeight != nil {
		scaleX, scaleY := window.GetContentScale()
		minWidth := glfw.DontCare
		if cfg.minWidth != nil {
			minWidth = scaled(*cfg.minWidth, scaleX)
		}
		minHeight := glfw.DontCare
		if cfg.minHeight != nil {
			minHeight = scaled(*cfg.minHeight, scaleY)
		}
		maxWidth := glfw.DontCare
		if cfg.maxWidth != nil {
			maxWidth = scaled(*cfg.maxWidth, scaleX)
		}
		maxHeight := glfw.DontCare
		if cfg.maxHeight != nil {
			maxHeight = scaled(*cfg.maxHeight, scaleY)
		}
		window.SetSizeLimits(minWidth, minHeight, maxWidth, maxHeight)
	}

	if cfg.icon != "" {
		img, err := openImage(cfg.locator, cfg.icon)
		if err != nil {
			return fmt.Errorf("failed to open icon %q: %w", cfg.icon, err)
		}
		window.SetIcon([]image.Image{img})
	}

	window.MakeContextCurrent()
	defer glfw.DetachCurrentContext()
	glfw.SwapInterval(cfg.swapInterval)

	if err := gl.Init(); err != nil {
		return fmt.Errorf("failed to initialize opengl: %w", err)
	}

	l := newLoop(cfg.locator, cfg.title, window, controller)

	if cfg.cursor != nil {
		cursor := l.CreateCursor(*cfg.cursor)
		defer cursor.Destroy()
		l.UseCursor(cursor)
		defer l.UseCursor(nil)
	}

	if !cfg.cursorVisible {
		l.SetCursorVisible(false)
	}

	return l.Run(cfg.audioEnabled)
}

func scaled[T constr.Numeric](size T, scale float32) T {
	return T(float64(size) * float64(scale))
}

func invScaled[T constr.Numeric](size T, scale float32) T {
	return T(float64(size) / max(1e-6, float64(scale)))
}
