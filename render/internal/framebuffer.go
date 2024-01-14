package internal

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewFramebuffer(info render.FramebufferInfo) *Framebuffer {
	var id uint32
	gl.GenFramebuffers(1, &id)
	gl.BindFramebuffer(gl.FRAMEBUFFER, id)

	var activeDrawBuffers [4]bool
	var drawBuffers []uint32
	for i, attachment := range info.ColorAttachments {
		if colorAttachment, ok := attachment.(*Texture); ok {
			attachmentID := gl.COLOR_ATTACHMENT0 + uint32(i)
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, attachmentID, gl.TEXTURE_2D, colorAttachment.id, 0)
			drawBuffers = append(drawBuffers, attachmentID)
			activeDrawBuffers[i] = true
		}
	}

	if depthStencilAttachment, ok := info.DepthStencilAttachment.(*Texture); ok {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.TEXTURE_2D, depthStencilAttachment.id, 0)
	} else {
		if depthAttachment, ok := info.DepthAttachment.(*Texture); ok {
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, depthAttachment.id, 0)
		}
		if stencilAttachment, ok := info.StencilAttachment.(*Texture); ok {
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.STENCIL_ATTACHMENT, gl.TEXTURE_2D, stencilAttachment.id, 0)
		}
	}

	if len(drawBuffers) > 0 {
		gl.DrawBuffers(int32(len(drawBuffers)), &drawBuffers[0])
	} else {
		gl.DrawBuffers(int32(len(drawBuffers)), nil)
	}

	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if status != gl.FRAMEBUFFER_COMPLETE {
		logger.Error("Framebuffer (%q) is incomplete!", info.Label)
	}

	return &Framebuffer{
		id:                id,
		activeDrawBuffers: activeDrawBuffers,
	}
}

var DefaultFramebuffer = &Framebuffer{
	id:                0,
	activeDrawBuffers: [4]bool{true, false, false, false},
}

type Framebuffer struct {
	render.FramebufferObject
	id                uint32
	activeDrawBuffers [4]bool
}

func (f *Framebuffer) Release() {
	gl.DeleteFramebuffers(1, &f.id)
	f.id = 0
	f.activeDrawBuffers = [4]bool{}
}

func DetermineContentFormat(framebuffer render.Framebuffer) render.DataFormat {
	fb := framebuffer.(*Framebuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.id)
	defer func() {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	}()
	var glFormat int32
	gl.GetIntegerv(
		gl.IMPLEMENTATION_COLOR_READ_FORMAT,
		&glFormat,
	)
	if glFormat != gl.RGBA {
		return render.DataFormatUnsupported
	}
	var glType int32
	gl.GetIntegerv(
		gl.IMPLEMENTATION_COLOR_READ_TYPE,
		&glType,
	)
	switch glType {
	case gl.UNSIGNED_BYTE:
		return render.DataFormatRGBA8
	case gl.HALF_FLOAT:
		return render.DataFormatRGBA16F
	case gl.FLOAT:
		return render.DataFormatRGBA32F
	default:
		return render.DataFormatUnsupported
	}
}
