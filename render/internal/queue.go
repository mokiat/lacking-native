package internal

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewQueue() *Queue {
	return &Queue{}
}

type Queue struct {
}

func (q *Queue) Invalidate() {
	// TODO
}

func (q *Queue) WriteBuffer(buffer render.Buffer, offset int, data []byte) {
	actualBuffer := buffer.(*Buffer)
	gl.BindBuffer(actualBuffer.kind, actualBuffer.id)
	gl.BufferSubData(actualBuffer.kind, offset, len(data), gl.Ptr(&data[0]))
	gl.BindBuffer(actualBuffer.kind, 0)
}
