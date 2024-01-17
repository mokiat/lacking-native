package internal

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewVertexArray(info render.VertexArrayInfo) *VertexArray {
	var id uint32
	gl.GenVertexArrays(1, &id)
	gl.BindVertexArray(id)

	for _, attribute := range info.Attributes {
		binding := info.Bindings[attribute.Binding]
		if vertexBuffer, ok := binding.VertexBuffer.(*Buffer); ok {
			gl.BindBuffer(vertexBuffer.kind, vertexBuffer.id)
		}
		gl.EnableVertexAttribArray(uint32(attribute.Location))
		count, compType, normalized, integer := glAttribParams(attribute.Format)
		if integer {
			gl.VertexAttribIPointer(uint32(attribute.Location), count, compType, int32(binding.Stride), gl.PtrOffset(attribute.Offset))
		} else {
			gl.VertexAttribPointer(uint32(attribute.Location), count, compType, normalized, int32(binding.Stride), gl.PtrOffset(attribute.Offset))
		}
	}
	if indexBuffer, ok := info.IndexBuffer.(*Buffer); ok {
		gl.BindBuffer(indexBuffer.kind, indexBuffer.id)
	}
	gl.BindVertexArray(0)

	return &VertexArray{
		id:          id,
		indexFormat: glIndexFormat(info.IndexFormat),
	}
}

type VertexArray struct {
	render.VertexArrayMarker
	id          uint32
	indexFormat uint32
}

func (a *VertexArray) Release() {
	gl.DeleteVertexArrays(1, &a.id)
	a.id = 0
}

func glAttribParams(format render.VertexAttributeFormat) (int32, uint32, bool, bool) {
	switch format {
	case render.VertexAttributeFormatR32F:
		return 1, gl.FLOAT, false, false
	case render.VertexAttributeFormatRG32F:
		return 2, gl.FLOAT, false, false
	case render.VertexAttributeFormatRGB32F:
		return 3, gl.FLOAT, false, false
	case render.VertexAttributeFormatRGBA32F:
		return 4, gl.FLOAT, false, false

	case render.VertexAttributeFormatR16F:
		return 1, gl.HALF_FLOAT, false, false
	case render.VertexAttributeFormatRG16F:
		return 2, gl.HALF_FLOAT, false, false
	case render.VertexAttributeFormatRGB16F:
		return 3, gl.HALF_FLOAT, false, false
	case render.VertexAttributeFormatRGBA16F:
		return 4, gl.HALF_FLOAT, false, false

	case render.VertexAttributeFormatR16S:
		return 1, gl.SHORT, false, false
	case render.VertexAttributeFormatRG16S:
		return 2, gl.SHORT, false, false
	case render.VertexAttributeFormatRGB16S:
		return 3, gl.SHORT, false, false
	case render.VertexAttributeFormatRGBA16S:
		return 4, gl.SHORT, false, false

	case render.VertexAttributeFormatR16SN:
		return 1, gl.SHORT, true, false
	case render.VertexAttributeFormatRG16SN:
		return 2, gl.SHORT, true, false
	case render.VertexAttributeFormatRGB16SN:
		return 3, gl.SHORT, true, false
	case render.VertexAttributeFormatRGBA16SN:
		return 4, gl.SHORT, true, false

	case render.VertexAttributeFormatR16U:
		return 1, gl.UNSIGNED_SHORT, false, false
	case render.VertexAttributeFormatRG16U:
		return 2, gl.UNSIGNED_SHORT, false, false
	case render.VertexAttributeFormatRGB16U:
		return 3, gl.UNSIGNED_SHORT, false, false
	case render.VertexAttributeFormatRGBA16U:
		return 4, gl.UNSIGNED_SHORT, false, false

	case render.VertexAttributeFormatR16UN:
		return 1, gl.UNSIGNED_SHORT, true, false
	case render.VertexAttributeFormatRG16UN:
		return 2, gl.UNSIGNED_SHORT, true, false
	case render.VertexAttributeFormatRGB16UN:
		return 3, gl.UNSIGNED_SHORT, true, false
	case render.VertexAttributeFormatRGBA16UN:
		return 4, gl.UNSIGNED_SHORT, true, false

	case render.VertexAttributeFormatR8S:
		return 1, gl.BYTE, false, false
	case render.VertexAttributeFormatRG8S:
		return 2, gl.BYTE, false, false
	case render.VertexAttributeFormatRGB8S:
		return 3, gl.BYTE, false, false
	case render.VertexAttributeFormatRGBA8S:
		return 4, gl.BYTE, false, false

	case render.VertexAttributeFormatR8SN:
		return 1, gl.BYTE, true, false
	case render.VertexAttributeFormatRG8SN:
		return 2, gl.BYTE, true, false
	case render.VertexAttributeFormatRGB8SN:
		return 3, gl.BYTE, true, false
	case render.VertexAttributeFormatRGBA8SN:
		return 4, gl.BYTE, true, false

	case render.VertexAttributeFormatR8U:
		return 1, gl.UNSIGNED_BYTE, false, false
	case render.VertexAttributeFormatRG8U:
		return 2, gl.UNSIGNED_BYTE, false, false
	case render.VertexAttributeFormatRGB8U:
		return 3, gl.UNSIGNED_BYTE, false, false
	case render.VertexAttributeFormatRGBA8U:
		return 4, gl.UNSIGNED_BYTE, false, false

	case render.VertexAttributeFormatR8UN:
		return 1, gl.UNSIGNED_BYTE, true, false
	case render.VertexAttributeFormatRG8UN:
		return 2, gl.UNSIGNED_BYTE, true, false
	case render.VertexAttributeFormatRGB8UN:
		return 3, gl.UNSIGNED_BYTE, true, false
	case render.VertexAttributeFormatRGBA8UN:
		return 4, gl.UNSIGNED_BYTE, true, false

	case render.VertexAttributeFormatR8IU:
		return 1, gl.UNSIGNED_BYTE, false, true
	case render.VertexAttributeFormatRG8IU:
		return 2, gl.UNSIGNED_BYTE, false, true
	case render.VertexAttributeFormatRGB8IU:
		return 3, gl.UNSIGNED_BYTE, false, true
	case render.VertexAttributeFormatRGBA8IU:
		return 4, gl.UNSIGNED_BYTE, false, true

	default:
		panic(fmt.Errorf("unknown attribute format: %d", format))
	}
}

func glIndexFormat(format render.IndexFormat) uint32 {
	switch format {
	case render.IndexFormatUnsignedShort:
		return gl.UNSIGNED_SHORT
	case render.IndexFormatUnsignedInt:
		return gl.UNSIGNED_INT
	default:
		panic(fmt.Errorf("unknown index format: %d", format))
	}
}
