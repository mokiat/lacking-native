package internal

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func trackError(format string, args ...any) func() {
	clearErrors()
	return func() {
		if err := getError(); err != "" {
			logger.Error(format+": "+err, args...)
		}
	}
}

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
