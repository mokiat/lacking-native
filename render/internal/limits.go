package internal

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewLimits() *Limits {
	var uniformBufferOffsetAlignment int32
	gl.GetIntegerv(gl.UNIFORM_BUFFER_OFFSET_ALIGNMENT, &uniformBufferOffsetAlignment)

	return &Limits{
		uniformBufferOffsetAlignment: int(uniformBufferOffsetAlignment),
	}
}

type Limits struct {
	uniformBufferOffsetAlignment int
}

func (l Limits) UniformBufferOffsetAlignment() int {
	return l.uniformBufferOffsetAlignment
}

func (l Limits) Quality() render.Quality {
	return render.QualityHigh
}
