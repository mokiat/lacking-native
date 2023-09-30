package app

import (
	"fmt"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	glrender "github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/resource"
)

const (
	taskQueueSize         = 1024
	taskProcessingTimeout = 30 * time.Millisecond
)

func newLoop(locator resource.ReadLocator, title string, window *glfw.Window, controller app.Controller) *loop {
	return &loop{
		platform:      newPlatform(),
		locator:       locator,
		title:         title,
		window:        window,
		controller:    controller,
		renderAPI:     glrender.NewAPI(),
		tasks:         make(chan func(), taskQueueSize),
		shouldStop:    false,
		shouldDraw:    true,
		cursorVisible: true,
		cursorLocked:  false,
		gamepads: [4]*Gamepad{
			newGamepad(glfw.Joystick1),
			newGamepad(glfw.Joystick2),
			newGamepad(glfw.Joystick3),
			newGamepad(glfw.Joystick4),
		},
	}
}

var _ app.Window = (*loop)(nil)

type loop struct {
	platform      *platform
	locator       resource.ReadLocator
	title         string
	window        *glfw.Window
	controller    app.Controller
	renderAPI     render.API
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

	l.window.SetRefreshCallback(l.onGLFWRefresh)

	l.window.SetSizeCallback(l.onGLFWSize)
	width, height := l.window.GetSize()
	l.onGLFWSize(l.window, width, height)

	l.window.SetFramebufferSizeCallback(l.onGLFWFramebufferSize)
	width, height = l.window.GetFramebufferSize()
	l.onGLFWFramebufferSize(l.window, width, height)

	l.window.SetKeyCallback(l.onGLFWKey)
	l.window.SetCharCallback(l.onGLFWChar)

	l.window.SetCursorPosCallback(l.onGLFWCursorPos)
	l.window.SetCursorEnterCallback(l.onGLFWCursorEnter)
	l.window.SetMouseButtonCallback(l.onGLFWMouseButton)
	l.window.SetScrollCallback(l.onGLFWScroll)
	l.window.SetDropCallback(l.onGLFWMouseDrop)

	for !l.shouldStop {
		if l.shouldWake {
			l.shouldWake = false
			glfw.PollEvents()
		} else {
			glfw.WaitEvents()
		}

		if l.window.ShouldClose() {
			l.controller.OnCloseRequested(l)
			l.window.SetShouldClose(false)
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
			l.controller.OnRender(l)
			l.window.SwapBuffers()
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
	return l.window.GetSize()
}

func (l *loop) SetSize(width, height int) {
	l.window.SetSize(width, height)
}

func (l *loop) FramebufferSize() (int, int) {
	return l.window.GetFramebufferSize()
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
		glfw.PostEmptyEvent()
	default:
		panic(fmt.Errorf("failed to queue task; queue is full"))
	}
}

func (l *loop) Invalidate() {
	if !l.shouldDraw {
		l.shouldDraw = true
		if !l.shouldWake {
			l.shouldWake = true
			glfw.PostEmptyEvent()
		}
	}
}

func (l *loop) CreateCursor(definition app.CursorDefinition) app.Cursor {
	img, err := openImage(l.locator, definition.Path)
	if err != nil {
		panic(fmt.Errorf("failed to open cursor %q: %w", definition.Path, err))
	}
	return &customCursor{
		cursor: glfw.CreateCursor(img, definition.HotspotX, definition.HotspotY),
	}
}

func (l *loop) UseCursor(cursor app.Cursor) {
	switch cursor := cursor.(type) {
	case *customCursor:
		l.window.SetCursor(cursor.cursor)
	default:
		l.window.SetCursor(nil)
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

func (l *loop) RenderAPI() render.API {
	return l.renderAPI
}

func (l *loop) AudioAPI() audio.API {
	return nil // TODO
}

func (l *loop) Close() {
	if !l.shouldStop {
		l.shouldStop = true
		glfw.PostEmptyEvent()
	}
}

func (l *loop) updateCursorMode() {
	switch {
	case l.cursorLocked:
		l.window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	case l.cursorVisible:
		l.window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	default:
		l.window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
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

func (l *loop) onGLFWRefresh(w *glfw.Window) {
	l.controller.OnRender(l)
	l.window.SwapBuffers()
}

func (l *loop) onGLFWSize(w *glfw.Window, width int, height int) {
	l.controller.OnResize(l, width, height)
}

func (l *loop) onGLFWFramebufferSize(w *glfw.Window, width int, height int) {
	l.controller.OnFramebufferResize(l, width, height)
}

func (l *loop) onGLFWKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	eventType, ok := keyboardEventTypeMapping[action]
	if !ok {
		return
	}
	keyCode, ok := keyboardKeyMapping[key]
	if !ok {
		return
	}
	var modifiers app.KeyModifierSet
	if (mods & glfw.ModControl) != 0b0 {
		modifiers = modifiers | app.KeyModifierSet(app.KeyModifierControl)
	}
	if (mods & glfw.ModShift) != 0b0 {
		modifiers = modifiers | app.KeyModifierSet(app.KeyModifierShift)
	}
	if (mods & glfw.ModAlt) != 0b0 {
		modifiers = modifiers | app.KeyModifierSet(app.KeyModifierAlt)
	}
	if (mods & glfw.ModCapsLock) != 0b0 {
		modifiers = modifiers | app.KeyModifierSet(app.KeyModifierCapsLock)
	}
	if (mods & glfw.ModSuper) != 0b0 {
		modifiers = modifiers | app.KeyModifierSet(app.KeyModifierSuper)
	}
	l.controller.OnKeyboardEvent(l, app.KeyboardEvent{
		Type:      eventType,
		Code:      keyCode,
		Modifiers: modifiers,
	})
}

func (l *loop) onGLFWChar(w *glfw.Window, char rune) {
	l.controller.OnKeyboardEvent(l, app.KeyboardEvent{
		Type: app.KeyboardEventTypeType,
		Rune: char,
	})
}

func (l *loop) onGLFWCursorPos(w *glfw.Window, xpos float64, ypos float64) {
	l.controller.OnMouseEvent(l, app.MouseEvent{
		Index: 0,
		X:     int(xpos),
		Y:     int(ypos),
		Type:  app.MouseEventTypeMove,
	})
}

func (l *loop) onGLFWCursorEnter(w *glfw.Window, entered bool) {
	var eventType app.MouseEventType
	if entered {
		eventType = app.MouseEventTypeEnter
	} else {
		eventType = app.MouseEventTypeLeave
	}
	xpos, ypos := l.window.GetCursorPos()
	l.controller.OnMouseEvent(l, app.MouseEvent{
		Index: 0,
		X:     int(xpos),
		Y:     int(ypos),
		Type:  eventType,
	})
}

func (l *loop) onGLFWMouseButton(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	xpos, ypos := l.window.GetCursorPos()
	var eventType app.MouseEventType
	switch action {
	case glfw.Press:
		eventType = app.MouseEventTypeDown
	case glfw.Release:
		eventType = app.MouseEventTypeUp
	}
	var eventButton app.MouseButton
	switch button {
	case glfw.MouseButton1:
		eventButton = app.MouseButtonLeft
	case glfw.MouseButton2:
		eventButton = app.MouseButtonRight
	case glfw.MouseButton3:
		eventButton = app.MouseButtonMiddle
	}
	l.controller.OnMouseEvent(l, app.MouseEvent{
		Index:  0,
		X:      int(xpos),
		Y:      int(ypos),
		Type:   eventType,
		Button: eventButton,
	})
}

func (l *loop) onGLFWScroll(w *glfw.Window, xoff float64, yoff float64) {
	xpos, ypos := l.window.GetCursorPos()
	l.controller.OnMouseEvent(l, app.MouseEvent{
		Index:   0,
		X:       int(xpos),
		Y:       int(ypos),
		Type:    app.MouseEventTypeScroll,
		ScrollX: xoff,
		ScrollY: yoff,
	})
}

func (l *loop) onGLFWMouseDrop(w *glfw.Window, names []string) {
	xpos, ypos := l.window.GetCursorPos()
	l.controller.OnMouseEvent(l, app.MouseEvent{
		Index: 0,
		X:     int(xpos),
		Y:     int(ypos),
		Type:  app.MouseEventTypeDrop,
		Payload: app.FilepathPayload{
			Paths: names,
		},
	})
}
