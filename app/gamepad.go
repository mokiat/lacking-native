package app

import (
	"math"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/app"
)

func newGamepad(joystick glfw.Joystick) *Gamepad {
	return &Gamepad{
		joystick: joystick,

		isDirty:     true,
		isConnected: false,
		isSupported: false,

		deadzoneStick:   0.1,
		deadzoneTrigger: 0.0,
	}
}

type Gamepad struct {
	joystick glfw.Joystick

	isDirty     bool
	isConnected bool
	isSupported bool

	deadzoneStick   float64
	deadzoneTrigger float64

	leftStickX        float64
	leftStickY        float64
	leftStickButton   bool
	rightStickX       float64
	rightStickY       float64
	rightStickButton  bool
	leftBumperButton  bool
	leftTrigger       float64
	rightBumperButton bool
	rightTrigger      float64
	dpadLeftButton    bool
	dpadRightButton   bool
	dpadUpButton      bool
	dpadDownButton    bool
	actionLeftButton  bool
	actionRightButton bool
	actionUpButton    bool
	actionDownButton  bool
	forwardButton     bool
	backButton        bool
}

var _ app.Gamepad = (*Gamepad)(nil)

func (g *Gamepad) Connected() bool {
	g.refresh()
	return g.isConnected
}

func (g *Gamepad) Supported() bool {
	g.refresh()
	return g.isSupported
}

func (g *Gamepad) StickDeadzone() float64 {
	return g.deadzoneStick
}

func (g *Gamepad) SetStickDeadzone(deadzone float64) {
	g.deadzoneStick = deadzone
}

func (g *Gamepad) TriggerDeadzone() float64 {
	return g.deadzoneTrigger
}

func (g *Gamepad) SetTriggerDeadzone(deadzone float64) {
	g.deadzoneTrigger = deadzone
}

func (g *Gamepad) LeftStickX() float64 {
	g.refresh()
	return deadzoneValue(g.leftStickX, g.deadzoneStick)
}

func (g *Gamepad) LeftStickY() float64 {
	g.refresh()
	return deadzoneValue(g.leftStickY, g.deadzoneStick)
}

func (g *Gamepad) LeftStickButton() bool {
	g.refresh()
	return g.leftStickButton
}

func (g *Gamepad) RightStickX() float64 {
	g.refresh()
	return deadzoneValue(g.rightStickX, g.deadzoneStick)
}

func (g *Gamepad) RightStickY() float64 {
	g.refresh()
	return deadzoneValue(g.rightStickY, g.deadzoneStick)
}

func (g *Gamepad) RightStickButton() bool {
	g.refresh()
	return g.rightStickButton
}

func (g *Gamepad) LeftTrigger() float64 {
	g.refresh()
	return deadzoneValue(g.leftTrigger, g.deadzoneTrigger)
}

func (g *Gamepad) RightTrigger() float64 {
	g.refresh()
	return deadzoneValue(g.rightTrigger, g.deadzoneTrigger)
}

func (g *Gamepad) LeftBumper() bool {
	g.refresh()
	return g.leftBumperButton
}

func (g *Gamepad) RightBumper() bool {
	g.refresh()
	return g.rightBumperButton
}

func (g *Gamepad) DpadUpButton() bool {
	g.refresh()
	return g.dpadUpButton
}

func (g *Gamepad) DpadDownButton() bool {
	g.refresh()
	return g.dpadDownButton
}

func (g *Gamepad) DpadLeftButton() bool {
	g.refresh()
	return g.dpadLeftButton
}

func (g *Gamepad) DpadRightButton() bool {
	g.refresh()
	return g.dpadRightButton
}

func (g *Gamepad) ActionUpButton() bool {
	g.refresh()
	return g.actionUpButton
}

func (g *Gamepad) ActionDownButton() bool {
	g.refresh()
	return g.actionDownButton
}

func (g *Gamepad) ActionLeftButton() bool {
	g.refresh()
	return g.actionLeftButton
}

func (g *Gamepad) ActionRightButton() bool {
	g.refresh()
	return g.actionRightButton
}

func (g *Gamepad) ForwardButton() bool {
	g.refresh()
	return g.forwardButton
}

func (g *Gamepad) BackButton() bool {
	g.refresh()
	return g.backButton
}

func (g *Gamepad) Pulse(intensity float64, duration time.Duration) {
	// Haptic feedback is still not supported by glfw.
}

func (g *Gamepad) markDirty() {
	g.isDirty = true
}

func (g *Gamepad) refresh() {
	if !g.isDirty {
		return
	}
	g.isDirty = false
	g.isConnected = g.joystick.Present()
	if g.isConnected {
		g.isSupported = g.joystick.IsGamepad()
	} else {
		g.isSupported = false
	}
	if g.isSupported {
		state := g.joystick.GetGamepadState()
		g.leftStickX = float64(state.Axes[glfw.AxisLeftX])
		g.leftStickY = float64(state.Axes[glfw.AxisLeftY])
		g.leftStickButton = state.Buttons[glfw.ButtonLeftThumb] == glfw.Press
		g.rightStickX = float64(state.Axes[glfw.AxisRightX])
		g.rightStickY = float64(state.Axes[glfw.AxisRightY])
		g.rightStickButton = state.Buttons[glfw.ButtonRightThumb] == glfw.Press
		g.leftBumperButton = state.Buttons[glfw.ButtonLeftBumper] == glfw.Press
		g.leftTrigger = float64(state.Axes[glfw.AxisLeftTrigger]+1.0) / 2.0
		g.rightBumperButton = state.Buttons[glfw.ButtonRightBumper] == glfw.Press
		g.rightTrigger = float64(state.Axes[glfw.AxisRightTrigger]+1.0) / 2.0
		g.dpadLeftButton = state.Buttons[glfw.ButtonDpadLeft] == glfw.Press
		g.dpadRightButton = state.Buttons[glfw.ButtonDpadRight] == glfw.Press
		g.dpadUpButton = state.Buttons[glfw.ButtonDpadUp] == glfw.Press
		g.dpadDownButton = state.Buttons[glfw.ButtonDpadDown] == glfw.Press
		g.actionLeftButton = state.Buttons[glfw.ButtonSquare] == glfw.Press
		g.actionRightButton = state.Buttons[glfw.ButtonCircle] == glfw.Press
		g.actionUpButton = state.Buttons[glfw.ButtonTriangle] == glfw.Press
		g.actionDownButton = state.Buttons[glfw.ButtonCross] == glfw.Press
		g.forwardButton = state.Buttons[glfw.ButtonStart] == glfw.Press
		g.backButton = state.Buttons[glfw.ButtonBack] == glfw.Press
	} else {
		g.leftStickX = 0.0
		g.leftStickY = 0.0
		g.leftStickButton = false
		g.rightStickX = 0.0
		g.rightStickY = 0.0
		g.rightStickButton = false
		g.leftBumperButton = false
		g.leftTrigger = 0.0
		g.rightBumperButton = false
		g.rightTrigger = 0.0
		g.dpadLeftButton = false
		g.dpadRightButton = false
		g.dpadUpButton = false
		g.dpadDownButton = false
		g.actionLeftButton = false
		g.actionRightButton = false
		g.actionUpButton = false
		g.actionDownButton = false
		g.forwardButton = false
		g.backButton = false
	}
}

func deadzoneValue(value, deadzone float64) float64 {
	if math.Signbit(value) {
		// negative
		value = dprec.Max(-value, deadzone)
		value = value - deadzone
		return -value / (1.0 - deadzone)
	} else {
		// positive
		value = dprec.Max(value, deadzone)
		value = value - deadzone
		return value / (1.0 - deadzone)
	}
}
