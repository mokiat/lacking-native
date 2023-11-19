package app

import (
	"fmt"
	"image"
	"image/draw"
	"time"

	nataudio "github.com/mokiat/lacking-native/audio"
	natrender "github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/debug/log"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/resource"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	taskQueueSize         = 1024
	taskProcessingTimeout = 30 * time.Millisecond
)

func newLoop(locator resource.ReadLocator, title string, window *sdl.Window, controller app.Controller) *loop {
	return &loop{
		platform:      newPlatform(),
		locator:       locator,
		title:         title,
		window:        window,
		controller:    controller,
		renderAPI:     natrender.NewAPI(),
		audioAPI:      nataudio.NewAPI(),
		tasks:         make(chan func(), taskQueueSize),
		shouldStop:    false,
		shouldDraw:    true,
		cursorVisible: true,
		cursorLocked:  false,
		gamepads: [4]*Gamepad{
			newGamepad(0),
			newGamepad(1),
			newGamepad(2),
			newGamepad(3),
		},
	}
}

var _ app.Window = (*loop)(nil)

type loop struct {
	platform      *platform
	locator       resource.ReadLocator
	title         string
	window        *sdl.Window
	controller    app.Controller
	renderAPI     render.API
	audioAPI      audio.API
	tasks         chan func()
	shouldStop    bool
	shouldDraw    bool
	shouldWake    bool
	cursorVisible bool
	cursorLocked  bool
	gamepads      [4]*Gamepad
}

func (l *loop) Run() error {
	l.controller.OnCreate(l)

	width, height := l.window.GetSize()
	l.onSizeChanged(int(width), int(height))

	width, height = l.window.GLGetDrawableSize()
	l.onFramebufferSizeChanged(int(width), int(height))

	for !l.shouldStop {
		if l.shouldWake {
			l.shouldWake = false
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				l.handleEvent(event)
			}
		} else {
			for event := sdl.WaitEvent(); event != nil; event = sdl.PollEvent() {
				l.handleEvent(event)
			}
		}

		for _, gamepad := range l.gamepads {
			gamepad.markDirty()
		}

		if !l.processTasks(taskProcessingTimeout) {
			// Not all events were processed, loop should not
			// block on next iteration.
			l.shouldWake = true
		}

		if l.shouldDraw {
			l.shouldDraw = false
			metric.BeginFrame()

			ctrlRegion := metric.BeginRegion("controller")
			l.controller.OnRender(l)
			ctrlRegion.End()

			swapRegion := metric.BeginRegion("swap")
			l.window.GLSwap()
			swapRegion.End()

			metric.EndFrame()
		}
	}

	l.controller.OnDestroy(l)

	// Give any async tasks a chance to complete.
	if !l.processTasks(5 * time.Second) {
		return fmt.Errorf("failed to cleanup within timeout")
	}

	return nil
}

func (l *loop) Platform() app.Platform {
	return l.platform
}

func (l *loop) Title() string {
	return l.title
}

func (l *loop) SetTitle(title string) {
	l.title = title
	l.window.SetTitle(title)
}

func (l *loop) Size() (int, int) {
	width, height := l.window.GetSize()
	return int(width), int(height)
}

func (l *loop) SetSize(width, height int) {
	l.window.SetSize(int32(width), int32(height))
}

func (l *loop) FramebufferSize() (int, int) {
	width, height := l.window.GLGetDrawableSize()
	return int(width), int(height)
}

func (l *loop) Gamepads() [4]app.Gamepad {
	var result [4]app.Gamepad
	for i := range result {
		result[i] = l.gamepads[i]
	}
	return result
}

func (l *loop) Schedule(fn func()) {
	select {
	case l.tasks <- fn:
		sdl.PushEvent(&sdl.UserEvent{})
	default:
		panic(fmt.Errorf("failed to queue task; queue is full"))
	}
}

func (l *loop) Invalidate() {
	if !l.shouldDraw {
		l.shouldDraw = true
		if !l.shouldWake {
			l.shouldWake = true
			sdl.PushEvent(&sdl.UserEvent{})
		}
	}
}

func (l *loop) CreateCursor(definition app.CursorDefinition) app.Cursor {
	img, err := openImage(l.locator, definition.Path)
	if err != nil {
		panic(fmt.Errorf("failed to open cursor %q: %w", definition.Path, err))
	}
	bounds := img.Bounds()

	surface, err := sdl.CreateRGBSurfaceWithFormat(0, int32(bounds.Dx()), int32(bounds.Dy()), 32, sdl.PIXELFORMAT_RGBA8888)
	if err != nil {
		panic(fmt.Errorf("error creating surface: %v", err))
	}
	draw.Draw(WrapSurface(surface), surface.Bounds(), img, image.Point{}, draw.Src)

	return &customCursor{
		surface: surface,
		cursor:  sdl.CreateColorCursor(surface, int32(definition.HotspotX), int32(definition.HotspotY)),
	}
}

func (l *loop) UseCursor(cursor app.Cursor) {
	switch cursor := cursor.(type) {
	case *customCursor:
		sdl.SetCursor(cursor.cursor)
	default:
		sdl.SetCursor(nil)
	}
}

func (l *loop) CursorVisible() bool {
	return l.cursorVisible && !l.cursorLocked
}

func (l *loop) SetCursorVisible(visible bool) {
	l.cursorVisible = visible
	l.updateCursorMode()
}

func (l *loop) SetCursorLocked(locked bool) {
	l.cursorLocked = locked
	l.updateCursorMode()
}

func (l *loop) RequestCopy(text string) {
	sdl.SetClipboardText(text)
}

func (l *loop) RequestPaste() {
	text, err := sdl.GetClipboardText()
	if err != nil {
		log.Error("Error getting clipboard text: %v", err)
		return
	}
	l.Schedule(func() {
		l.controller.OnClipboardEvent(l, app.ClipboardEvent{
			Text: text,
		})
	})
}

func (l *loop) RenderAPI() render.API {
	return l.renderAPI
}

func (l *loop) AudioAPI() audio.API {
	return l.audioAPI
}

func (l *loop) Close() {
	if !l.shouldStop {
		l.shouldStop = true
		sdl.PushEvent(&sdl.UserEvent{})
	}
}

func (l *loop) updateCursorMode() {
	switch {
	case l.cursorLocked:
		l.window.SetGrab(true)
		sdl.ShowCursor(0)
	case l.cursorVisible:
		l.window.SetGrab(false)
		sdl.ShowCursor(1)
	default:
		l.window.SetGrab(false)
		sdl.ShowCursor(0)
	}
}

func (l *loop) processTasks(limit time.Duration) bool {
	startTime := time.Now()
	for time.Since(startTime) < limit {
		select {
		case task := <-l.tasks:
			// There was a task in the queue so run it.
			task()
		default:
			// No more tasks, we have consumed everything there
			// is for now.
			return true
		}
	}
	// We did not consume all available tasks within our time window.
	return false
}

func (l *loop) handleEvent(event sdl.Event) {
	switch event := event.(type) {
	case *sdl.QuitEvent:
		l.onQuit()
	case *sdl.WindowEvent:
		l.onWindowEvent(event)
	case *sdl.MouseMotionEvent:
		l.onMouseMotion(event)
	case *sdl.MouseButtonEvent:
		l.onMouseButton(event)
	case *sdl.MouseWheelEvent:
		l.onMouseWheel(event)
	case *sdl.DropEvent:
		l.onDrop(event)
	case *sdl.KeyboardEvent:
		l.onKeyboard(event)
	case *sdl.TextInputEvent:
		l.onTextInput(event)
	case *sdl.ControllerDeviceEvent:
		l.onControllerEvent(event)
	}
}

func (l *loop) onQuit() {
	l.shouldStop = l.controller.OnCloseRequested(l)
}

func (l *loop) onWindowEvent(event *sdl.WindowEvent) {
	switch event.Event {
	case sdl.WINDOWEVENT_RESIZED:
		width, height := l.window.GetSize()
		l.onSizeChanged(int(width), int(height))
		width, height = l.window.GLGetDrawableSize()
		l.onFramebufferSizeChanged(int(width), int(height))

	case sdl.WINDOWEVENT_ENTER:
		xpos, ypos, _ := sdl.GetMouseState()
		l.controller.OnMouseEvent(l, app.MouseEvent{
			Index:  0,
			X:      int(xpos),
			Y:      int(ypos),
			Action: app.MouseActionEnter,
		})

	case sdl.WINDOWEVENT_LEAVE:
		xpos, ypos, _ := sdl.GetMouseState()
		l.controller.OnMouseEvent(l, app.MouseEvent{
			Index:  0,
			X:      int(xpos),
			Y:      int(ypos),
			Action: app.MouseActionLeave,
		})

	case sdl.WINDOWEVENT_SHOWN:
		l.controller.OnRender(l)
		l.window.GLSwap()
	}
}

func (l *loop) onMouseMotion(event *sdl.MouseMotionEvent) {
	l.controller.OnMouseEvent(l, app.MouseEvent{
		Index:  0,
		X:      int(event.X),
		Y:      int(event.Y),
		Action: app.MouseActionMove,
	})
}

func (l *loop) onMouseButton(event *sdl.MouseButtonEvent) {
	var eventType app.MouseAction
	switch event.Type {
	case sdl.MOUSEBUTTONDOWN:
		eventType = app.MouseActionDown
	case sdl.MOUSEBUTTONUP:
		eventType = app.MouseActionUp
	}

	var eventButton app.MouseButton
	switch event.Button {
	case sdl.BUTTON_LEFT:
		eventButton = app.MouseButtonLeft
	case sdl.BUTTON_RIGHT:
		eventButton = app.MouseButtonRight
	case sdl.BUTTON_MIDDLE:
		eventButton = app.MouseButtonMiddle
	}

	l.controller.OnMouseEvent(l, app.MouseEvent{
		Index:  0,
		X:      int(event.X),
		Y:      int(event.Y),
		Action: eventType,
		Button: eventButton,
	})
}

func (l *loop) onMouseWheel(event *sdl.MouseWheelEvent) {
	xpos, ypos, _ := sdl.GetMouseState()

	l.controller.OnMouseEvent(l, app.MouseEvent{
		Index:   0,
		X:       int(xpos),
		Y:       int(ypos),
		Action:  app.MouseActionScroll,
		ScrollX: -float64(event.PreciseX) * 20.0,
		ScrollY: float64(event.PreciseY) * 20.0,
	})
}

func (l *loop) onDrop(event *sdl.DropEvent) {
	if event.Type != sdl.DROPFILE {
		return
	}
	xpos, ypos, _ := sdl.GetMouseState()
	l.controller.OnMouseEvent(l, app.MouseEvent{
		Index:  0,
		X:      int(xpos),
		Y:      int(ypos),
		Action: app.MouseActionDrop,
		Payload: app.FilepathPayload{
			Paths: []string{event.File},
		},
	})
}

func (l *loop) onKeyboard(event *sdl.KeyboardEvent) {
	eventType, ok := keyboardActionMapping[event.Type]
	if !ok {
		return
	}
	keyCode, ok := keyboardKeyMapping[event.Keysym.Scancode]
	if !ok {
		return
	}
	l.controller.OnKeyboardEvent(l, app.KeyboardEvent{
		Action: eventType,
		Code:   keyCode,
	})
}

func (l *loop) onTextInput(event *sdl.TextInputEvent) {
	for _, char := range event.GetText() {
		l.controller.OnKeyboardEvent(l, app.KeyboardEvent{
			Action:    app.KeyboardActionType,
			Character: char,
		})
	}
}

func (l *loop) onControllerEvent(event *sdl.ControllerDeviceEvent) {
	switch event.Type {
	case sdl.CONTROLLERDEVICEADDED:
		if event.Which >= 0 && event.Which < 4 {
			l.gamepads[0].controller = sdl.GameControllerOpen(int(event.Which))
		}
	case sdl.CONTROLLERDEVICEREMOVED:
		for _, gamepad := range l.gamepads {
			if gamepad.hasInstanceID(event.Which) {
				gamepad.controller.Close()
				gamepad.controller = nil
			}
		}
	}
}

func (l *loop) onSizeChanged(width int, height int) {
	l.controller.OnResize(l, width, height)
}

func (l *loop) onFramebufferSizeChanged(width int, height int) {
	l.controller.OnFramebufferResize(l, width, height)
}
