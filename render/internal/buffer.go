package internal

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewVertexBuffer(info render.BufferInfo) *Buffer {
	return newBuffer(info, gl.ARRAY_BUFFER)
}

func NewIndexBuffer(info render.BufferInfo) *Buffer {
	return newBuffer(info, gl.ELEMENT_ARRAY_BUFFER)
}

func NewPixelTransferBuffer(info render.BufferInfo) render.Buffer {
	var id uint32
	gl.GenBuffers(1, &id)
	gl.BindBuffer(gl.PIXEL_PACK_BUFFER, id)
	gl.BufferData(gl.PIXEL_PACK_BUFFER, info.Size, nil, gl.DYNAMIC_READ)

	result := &Buffer{
		id:   id,
		kind: gl.PIXEL_PACK_BUFFER,
	}
	buffers.Track(id, result)
	return result
}

func NewUniformBuffer(info render.BufferInfo) render.Buffer {
	return newBuffer(info, gl.UNIFORM_BUFFER)
}

func newBuffer(info render.BufferInfo, kind uint32) *Buffer {
	var id uint32
	gl.GenBuffers(1, &id)
	gl.BindBuffer(kind, id)

	if info.Data != nil {
		gl.BufferData(kind, len(info.Data), gl.Ptr(&info.Data[0]), glBufferUsage(info.Dynamic))
	} else {
		gl.BufferData(kind, info.Size, nil, glBufferUsage(info.Dynamic))
	}
	result := &Buffer{
		id:   id,
		kind: kind,
	}
	buffers.Track(id, result)
	return result
}

type Buffer struct {
	render.BufferObject
	id   uint32
	kind uint32
}

func (b *Buffer) Update(info render.BufferUpdateInfo) {
	gl.BindBuffer(b.kind, b.id)
	gl.BufferSubData(b.kind, info.Offset, len(info.Data), gl.Ptr(&info.Data[0]))
}

func (b *Buffer) Fetch(info render.BufferFetchInfo) {
	gl.BindBuffer(b.kind, b.id)
	gl.GetBufferSubData(b.kind, info.Offset, len(info.Target), gl.Ptr(&info.Target[0]))
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
