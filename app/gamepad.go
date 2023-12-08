package app

import (
	"math"
	"time"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/app"
	"github.com/veandco/go-sdl2/sdl"
)

func newGamepad(index int) *Gamepad {
	return &Gamepad{
		index:      index,
		controller: sdl.GameControllerOpen(index),

		isDirty:     true,
		isConnected: false,
		isSupported: false,

		deadzoneStick:   0.1,
		deadzoneTrigger: 0.0,
	}
}

type Gamepad struct {
	index      int
	controller *sdl.GameController

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

func (g *Gamepad) hasInstanceID(instanceID sdl.JoystickID) bool {
	if g.controller == nil {
		return false
	}
	return g.controller.Joystick().InstanceID() == instanceID
}

func (g *Gamepad) markDirty() {
	g.isDirty = true
}

func (g *Gamepad) refresh() {
	if !g.isDirty {
		return
	}
	g.isDirty = false

	g.isSupported = g.controller != nil
	g.isConnected = g.isSupported && g.controller.Attached()
	if g.isConnected {
		g.leftStickX = float64(g.controller.Axis(sdl.CONTROLLER_AXIS_LEFTX)) / float64(math.MaxInt16)
		g.leftStickY = float64(g.controller.Axis(sdl.CONTROLLER_AXIS_LEFTY)) / float64(math.MaxInt16)
		g.leftStickButton = g.controller.Button(sdl.CONTROLLER_BUTTON_LEFTSTICK) == 1
		g.rightStickX = float64(g.controller.Axis(sdl.CONTROLLER_AXIS_RIGHTX)) / float64(math.MaxInt16)
		g.rightStickY = float64(g.controller.Axis(sdl.CONTROLLER_AXIS_RIGHTY)) / float64(math.MaxInt16)
		g.rightStickButton = g.controller.Button(sdl.CONTROLLER_BUTTON_RIGHTSTICK) == 1
		g.leftBumperButton = g.controller.Button(sdl.CONTROLLER_BUTTON_LEFTSHOULDER) == 1
		g.leftTrigger = float64(g.controller.Axis(sdl.CONTROLLER_AXIS_TRIGGERLEFT)) / float64(math.MaxInt16)
		g.rightBumperButton = g.controller.Button(sdl.CONTROLLER_BUTTON_RIGHTSHOULDER) == 1
		g.rightTrigger = float64(g.controller.Axis(sdl.CONTROLLER_AXIS_TRIGGERRIGHT)) / float64(math.MaxInt16)
		g.dpadLeftButton = g.controller.Button(sdl.CONTROLLER_BUTTON_DPAD_LEFT) == 1
		g.dpadRightButton = g.controller.Button(sdl.CONTROLLER_BUTTON_DPAD_RIGHT) == 1
		g.dpadUpButton = g.controller.Button(sdl.CONTROLLER_BUTTON_DPAD_UP) == 1
		g.dpadDownButton = g.controller.Button(sdl.CONTROLLER_BUTTON_DPAD_DOWN) == 1
		g.actionLeftButton = g.controller.Button(sdl.CONTROLLER_BUTTON_X) == 1
		g.actionRightButton = g.controller.Button(sdl.CONTROLLER_BUTTON_B) == 1
		g.actionUpButton = g.controller.Button(sdl.CONTROLLER_BUTTON_Y) == 1
		g.actionDownButton = g.controller.Button(sdl.CONTROLLER_BUTTON_A) == 1
		g.forwardButton = g.controller.Button(sdl.CONTROLLER_BUTTON_START) == 1
		g.backButton = g.controller.Button(sdl.CONTROLLER_BUTTON_BACK) == 1
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
