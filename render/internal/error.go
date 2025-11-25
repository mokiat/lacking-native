package internal

import (
	"context"
	"fmt"
	"log/slog"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

var isDebugEnabled = logger.Enabled(context.Background(), slog.LevelDebug)

func trackError(msg, label string) func() {
	if !isDebugEnabled {
		return nopFunc
	}
	clearErrors()
	return func() {
		if err := getError(); err != "" {
			logger.Error(msg,
				slog.String("label", label),
				slog.String("error", err),
			)
		}
	}
}

func nopFunc() {}

func clearErrors() {
	for gl.GetError() != gl.NO_ERROR {
	}
}

func getError() string {
	switch code := gl.GetError(); code {
	case gl.NO_ERROR:
		return ""
	case gl.INVALID_ENUM:
		return "INVALID_ENUM"
	case gl.INVALID_VALUE:
		return "INVALID_VALUE"
	case gl.INVALID_OPERATION:
		return "INVALID_OPERATION"
	case gl.INVALID_FRAMEBUFFER_OPERATION:
		return "INVALID_FRAMEBUFFER_OPERATION"
	case gl.OUT_OF_MEMORY:
		return "OUT_OF_MEMORY"
	case gl.STACK_UNDERFLOW:
		return "STACK_UNDERFLOW"
	case gl.STACK_OVERFLOW:
		return "STACK_OVERFLOW"
	default:
		return fmt.Sprintf("UNKNOWN_ERROR(%x)", code)
	}
}

func LogDebug() {
	if isDebugEnabled {
		gl.Enable(gl.DEBUG_OUTPUT)
		gl.DebugMessageCallback(func(source uint32, gltype uint32, id uint32, severity uint32, length int32, message string, userParam unsafe.Pointer) {
			switch severity {
			case gl.DEBUG_SEVERITY_LOW:
				logger.Debug(message)
			case gl.DEBUG_SEVERITY_MEDIUM:
				logger.Warn(message)
			case gl.DEBUG_SEVERITY_HIGH:
				logger.Error(message)
			default:
				logger.Debug(message)
			}
		}, gl.PtrOffset(0))
	}
}
