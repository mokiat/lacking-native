package app

import (
	"github.com/mokiat/lacking/app"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	keyboardActionMapping map[sdl.EventType]app.KeyboardAction
	keyboardKeyMapping    map[sdl.Scancode]app.KeyCode
)

func init() {
	keyboardActionMapping = make(map[sdl.EventType]app.KeyboardAction)
	keyboardActionMapping[sdl.KEYDOWN] = app.KeyboardActionDown
	keyboardActionMapping[sdl.KEYUP] = app.KeyboardActionUp

	keyboardKeyMapping = make(map[sdl.Scancode]app.KeyCode)
	keyboardKeyMapping[sdl.SCANCODE_ESCAPE] = app.KeyCodeEscape
	keyboardKeyMapping[sdl.SCANCODE_RETURN] = app.KeyCodeEnter
	keyboardKeyMapping[sdl.SCANCODE_SPACE] = app.KeyCodeSpace
	keyboardKeyMapping[sdl.SCANCODE_TAB] = app.KeyCodeTab
	keyboardKeyMapping[sdl.SCANCODE_CAPSLOCK] = app.KeyCodeCaps
	keyboardKeyMapping[sdl.SCANCODE_LSHIFT] = app.KeyCodeLeftShift
	keyboardKeyMapping[sdl.SCANCODE_RSHIFT] = app.KeyCodeRightShift
	keyboardKeyMapping[sdl.SCANCODE_LCTRL] = app.KeyCodeLeftControl
	keyboardKeyMapping[sdl.SCANCODE_RCTRL] = app.KeyCodeRightControl
	keyboardKeyMapping[sdl.SCANCODE_LALT] = app.KeyCodeLeftAlt
	keyboardKeyMapping[sdl.SCANCODE_RALT] = app.KeyCodeRightAlt
	keyboardKeyMapping[sdl.SCANCODE_LGUI] = app.KeyCodeLeftSuper
	keyboardKeyMapping[sdl.SCANCODE_RGUI] = app.KeyCodeRightSuper
	keyboardKeyMapping[sdl.SCANCODE_BACKSPACE] = app.KeyCodeBackspace
	keyboardKeyMapping[sdl.SCANCODE_INSERT] = app.KeyCodeInsert
	keyboardKeyMapping[sdl.SCANCODE_DELETE] = app.KeyCodeDelete
	keyboardKeyMapping[sdl.SCANCODE_HOME] = app.KeyCodeHome
	keyboardKeyMapping[sdl.SCANCODE_END] = app.KeyCodeEnd
	keyboardKeyMapping[sdl.SCANCODE_PAGEUP] = app.KeyCodePageUp
	keyboardKeyMapping[sdl.SCANCODE_PAGEDOWN] = app.KeyCodePageDown
	keyboardKeyMapping[sdl.SCANCODE_LEFT] = app.KeyCodeArrowLeft
	keyboardKeyMapping[sdl.SCANCODE_RIGHT] = app.KeyCodeArrowRight
	keyboardKeyMapping[sdl.SCANCODE_UP] = app.KeyCodeArrowUp
	keyboardKeyMapping[sdl.SCANCODE_DOWN] = app.KeyCodeArrowDown
	keyboardKeyMapping[sdl.SCANCODE_MINUS] = app.KeyCodeMinus
	keyboardKeyMapping[sdl.SCANCODE_EQUALS] = app.KeyCodeEqual
	keyboardKeyMapping[sdl.SCANCODE_LEFTBRACKET] = app.KeyCodeLeftBracket
	keyboardKeyMapping[sdl.SCANCODE_RIGHTBRACKET] = app.KeyCodeRightBracket
	keyboardKeyMapping[sdl.SCANCODE_SEMICOLON] = app.KeyCodeSemicolon
	keyboardKeyMapping[sdl.SCANCODE_COMMA] = app.KeyCodeComma
	keyboardKeyMapping[sdl.SCANCODE_PERIOD] = app.KeyCodePeriod
	keyboardKeyMapping[sdl.SCANCODE_SLASH] = app.KeyCodeSlash
	keyboardKeyMapping[sdl.SCANCODE_BACKSLASH] = app.KeyCodeBackslash
	keyboardKeyMapping[sdl.SCANCODE_APOSTROPHE] = app.KeyCodeApostrophe
	keyboardKeyMapping[sdl.SCANCODE_GRAVE] = app.KeyCodeGraveAccent
	keyboardKeyMapping[sdl.SCANCODE_A] = app.KeyCodeA
	keyboardKeyMapping[sdl.SCANCODE_B] = app.KeyCodeB
	keyboardKeyMapping[sdl.SCANCODE_C] = app.KeyCodeC
	keyboardKeyMapping[sdl.SCANCODE_D] = app.KeyCodeD
	keyboardKeyMapping[sdl.SCANCODE_E] = app.KeyCodeE
	keyboardKeyMapping[sdl.SCANCODE_F] = app.KeyCodeF
	keyboardKeyMapping[sdl.SCANCODE_G] = app.KeyCodeG
	keyboardKeyMapping[sdl.SCANCODE_H] = app.KeyCodeH
	keyboardKeyMapping[sdl.SCANCODE_I] = app.KeyCodeI
	keyboardKeyMapping[sdl.SCANCODE_J] = app.KeyCodeJ
	keyboardKeyMapping[sdl.SCANCODE_K] = app.KeyCodeK
	keyboardKeyMapping[sdl.SCANCODE_L] = app.KeyCodeL
	keyboardKeyMapping[sdl.SCANCODE_M] = app.KeyCodeM
	keyboardKeyMapping[sdl.SCANCODE_N] = app.KeyCodeN
	keyboardKeyMapping[sdl.SCANCODE_O] = app.KeyCodeO
	keyboardKeyMapping[sdl.SCANCODE_P] = app.KeyCodeP
	keyboardKeyMapping[sdl.SCANCODE_Q] = app.KeyCodeQ
	keyboardKeyMapping[sdl.SCANCODE_R] = app.KeyCodeR
	keyboardKeyMapping[sdl.SCANCODE_S] = app.KeyCodeS
	keyboardKeyMapping[sdl.SCANCODE_T] = app.KeyCodeT
	keyboardKeyMapping[sdl.SCANCODE_U] = app.KeyCodeU
	keyboardKeyMapping[sdl.SCANCODE_V] = app.KeyCodeV
	keyboardKeyMapping[sdl.SCANCODE_W] = app.KeyCodeW
	keyboardKeyMapping[sdl.SCANCODE_X] = app.KeyCodeX
	keyboardKeyMapping[sdl.SCANCODE_Y] = app.KeyCodeY
	keyboardKeyMapping[sdl.SCANCODE_Z] = app.KeyCodeZ
	keyboardKeyMapping[sdl.SCANCODE_0] = app.KeyCode0
	keyboardKeyMapping[sdl.SCANCODE_1] = app.KeyCode1
	keyboardKeyMapping[sdl.SCANCODE_2] = app.KeyCode2
	keyboardKeyMapping[sdl.SCANCODE_3] = app.KeyCode3
	keyboardKeyMapping[sdl.SCANCODE_4] = app.KeyCode4
	keyboardKeyMapping[sdl.SCANCODE_5] = app.KeyCode5
	keyboardKeyMapping[sdl.SCANCODE_6] = app.KeyCode6
	keyboardKeyMapping[sdl.SCANCODE_7] = app.KeyCode7
	keyboardKeyMapping[sdl.SCANCODE_8] = app.KeyCode8
	keyboardKeyMapping[sdl.SCANCODE_9] = app.KeyCode9
	keyboardKeyMapping[sdl.SCANCODE_F1] = app.KeyCodeF1
	keyboardKeyMapping[sdl.SCANCODE_F2] = app.KeyCodeF2
	keyboardKeyMapping[sdl.SCANCODE_F3] = app.KeyCodeF3
	keyboardKeyMapping[sdl.SCANCODE_F4] = app.KeyCodeF4
	keyboardKeyMapping[sdl.SCANCODE_F5] = app.KeyCodeF5
	keyboardKeyMapping[sdl.SCANCODE_F6] = app.KeyCodeF6
	keyboardKeyMapping[sdl.SCANCODE_F7] = app.KeyCodeF7
	keyboardKeyMapping[sdl.SCANCODE_F8] = app.KeyCodeF8
	keyboardKeyMapping[sdl.SCANCODE_F9] = app.KeyCodeF9
	keyboardKeyMapping[sdl.SCANCODE_F10] = app.KeyCodeF10
	keyboardKeyMapping[sdl.SCANCODE_F11] = app.KeyCodeF11
	keyboardKeyMapping[sdl.SCANCODE_F12] = app.KeyCodeF12
}
