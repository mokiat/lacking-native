package app

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"

	"github.com/mokiat/lacking/app"
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

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return fmt.Errorf("error initializing SDL2: %w", err)
	}
	defer sdl.Quit()

	if err := mix.Init(mix.INIT_MP3); err != nil {
		return fmt.Errorf("error initializing SDL2 MP3 mixer: %w", err)
	}
	defer mix.Quit()

	if err := mix.OpenAudio(44100, mix.DEFAULT_FORMAT, 2, 4096); err != nil {
		return fmt.Errorf("error opening audio device: %w", err)
	}
	defer mix.CloseAudio()

	var (
		windowWidth  = cfg.width
		windowHeight = cfg.height
		windowFlags  = sdl.WINDOW_SHOWN | sdl.WINDOW_OPENGL | sdl.WINDOW_RESIZABLE | sdl.WINDOW_ALLOW_HIGHDPI
	)
	if cfg.fullscreen {
		windowFlags |= sdl.WINDOW_FULLSCREEN_DESKTOP
	}
	if cfg.maximized {
		windowFlags |= sdl.WINDOW_MAXIMIZED
	}

	sdl.GLSetAttribute(sdl.GL_ACCELERATED_VISUAL, 1)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 1)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG, 1)
	sdl.GLSetAttribute(sdl.GL_FRAMEBUFFER_SRGB_CAPABLE, 1)

	window, err := sdl.CreateWindow(
		cfg.title,
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		int32(windowWidth),
		int32(windowHeight),
		uint32(windowFlags),
	)
	if err != nil {
		return fmt.Errorf("failed to create glfw window: %w", err)
	}
	defer window.Destroy()

	if cfg.minWidth != nil || cfg.minHeight != nil {
		minWidth := 1
		if cfg.minWidth != nil {
			minWidth = *cfg.minWidth
		}
		minHeight := 1
		if cfg.minHeight != nil {
			minHeight = *cfg.minHeight
		}
		window.SetMinimumSize(int32(minWidth), int32(minHeight))
	}
	if cfg.maxWidth != nil || cfg.maxHeight != nil {
		maxWidth := math.MaxInt32
		if cfg.maxWidth != nil {
			maxWidth = *cfg.maxWidth
		}
		maxHeight := math.MaxInt32
		if cfg.maxHeight != nil {
			maxHeight = *cfg.maxHeight
		}
		window.SetMaximumSize(int32(maxWidth), int32(maxHeight))
	}

	if cfg.icon != "" {
		img, err := openImage(cfg.locator, cfg.icon)
		if err != nil {
			return fmt.Errorf("failed to open icon %q: %w", cfg.icon, err)
		}
		bounds := img.Bounds()

		surface, err := sdl.CreateRGBSurfaceWithFormat(0, int32(bounds.Dx()), int32(bounds.Dy()), 32, sdl.PIXELFORMAT_RGBA8888)
		if err != nil {
			return fmt.Errorf("error creating surface: %v", err)
		}
		draw.Draw(surface, surface.Bounds(), img, image.Point{}, draw.Src)
		defer surface.Free()

		window.SetIcon(surface)
	}

	context, err := window.GLCreateContext()
	if err != nil {
		return fmt.Errorf("error creating gl context: %w", err)
	}
	defer sdl.GLDeleteContext(context)

	sdl.GLSetSwapInterval(cfg.swapInterval)

	if err := gl.Init(); err != nil {
		return fmt.Errorf("failed to initialize opengl: %w", err)
	}

	if glLogger.IsDebugEnabled() {
		gl.Enable(gl.DEBUG_OUTPUT)
		gl.DebugMessageCallback(func(source uint32, gltype uint32, id uint32, severity uint32, length int32, message string, userParam unsafe.Pointer) {
			switch severity {
			case gl.DEBUG_SEVERITY_LOW:
				glLogger.Debug(message)
			case gl.DEBUG_SEVERITY_MEDIUM:
				glLogger.Warn(message)
			case gl.DEBUG_SEVERITY_HIGH:
				glLogger.Error(message)
			default:
				glLogger.Debug(message)
			}
		}, gl.PtrOffset(0))
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

	return l.Run()
}
