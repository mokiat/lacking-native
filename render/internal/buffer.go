package internal

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewVertexBuffer(info render.BufferInfo) *Buffer {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating vertex buffer (%v)", info.Label)()
	}
	return newBuffer(info, gl.ARRAY_BUFFER)
}

func NewIndexBuffer(info render.BufferInfo) *Buffer {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating index buffer (%v)", info.Label)()
	}
	return newBuffer(info, gl.ELEMENT_ARRAY_BUFFER)
}

func NewPixelTransferBuffer(info render.BufferInfo) render.Buffer {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating pixel transfer buffer (%v)", info.Label)()
	}

	var id uint32
	gl.GenBuffers(1, &id)
	gl.BindBuffer(gl.PIXEL_PACK_BUFFER, id)
	gl.BufferData(gl.PIXEL_PACK_BUFFER, int(info.Size), nil, gl.DYNAMIC_READ)

	result := &Buffer{
		label: info.Label,
		id:    id,
		kind:  gl.PIXEL_PACK_BUFFER,
	}
	buffers.Track(id, result)
	return result
}

func NewUniformBuffer(info render.BufferInfo) render.Buffer {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating uniform buffer (%v)", info.Label)()
	}
	return newBuffer(info, gl.UNIFORM_BUFFER)
}

func newBuffer(info render.BufferInfo, kind uint32) *Buffer {
	var id uint32
	gl.GenBuffers(1, &id)
	gl.BindBuffer(kind, id)

	if info.Data != nil {
		gl.BufferData(kind, len(info.Data), gl.Ptr(&info.Data[0]), glBufferUsage(info.Dynamic))
	} else {
		gl.BufferData(kind, int(info.Size), nil, glBufferUsage(info.Dynamic))
	}
	result := &Buffer{
		label: info.Label,
		id:    id,
		kind:  kind,
	}
	buffers.Track(id, result)
	return result
}

type Buffer struct {
	render.BufferMarker

	label string
	id    uint32
	kind  uint32
}

func (b *Buffer) Label() string {
	return b.label
}

func (b *Buffer) Release() {
	buffers.Release(b.id)
	gl.DeleteBuffers(1, &b.id)
	b.id = 0
	b.kind = 0
}

func glBufferUsage(dynamic bool) uint32 {
	if dynamic {
		return gl.DYNAMIC_DRAW
	} else {
		return gl.STATIC_DRAW
	}
}
