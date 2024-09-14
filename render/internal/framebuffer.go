package internal

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewFramebuffer(info render.FramebufferInfo) *Framebuffer {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating framebuffer (%v)", info.Label)()
	}

	var id uint32
	gl.GenFramebuffers(1, &id)
	gl.BindFramebuffer(gl.FRAMEBUFFER, id)

	var activeDrawBuffers [4]bool
	var drawBuffers []uint32
	for i, colorAttachment := range info.ColorAttachments {
		if !colorAttachment.Specified {
			continue
		}
		attachment := colorAttachment.Value
		texture := attachment.Texture.(*Texture)
		attachmentID := gl.COLOR_ATTACHMENT0 + uint32(i)
		switch texture.kind {
		case gl.TEXTURE_2D_ARRAY:
			gl.FramebufferTextureLayer(gl.FRAMEBUFFER, attachmentID, texture.id, int32(attachment.MipmapLayer), int32(attachment.Depth))
		default:
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, attachmentID, gl.TEXTURE_2D, texture.id, int32(attachment.MipmapLayer))
		}
		drawBuffers = append(drawBuffers, attachmentID)
		activeDrawBuffers[i] = true
	}

	if info.DepthStencilAttachment.Specified {
		attachment := info.DepthStencilAttachment.Value
		texture := attachment.Texture.(*Texture)
		switch texture.kind {
		case gl.TEXTURE_2D_ARRAY:
			gl.FramebufferTextureLayer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, texture.id, int32(attachment.MipmapLayer), int32(attachment.Depth))
		default:
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.TEXTURE_2D, texture.id, int32(attachment.MipmapLayer))
		}
	} else {
		if info.DepthAttachment.Specified {
			attachment := info.DepthAttachment.Value
			texture := attachment.Texture.(*Texture)
			switch texture.kind {
			case gl.TEXTURE_2D_ARRAY:
				gl.FramebufferTextureLayer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, texture.id, int32(attachment.MipmapLayer), int32(attachment.Depth))
			default:
				gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, texture.id, int32(attachment.MipmapLayer))
			}
		}
		if info.StencilAttachment.Specified {
			attachment := info.StencilAttachment.Value
			texture := attachment.Texture.(*Texture)
			switch texture.kind {
			case gl.TEXTURE_2D_ARRAY:
				gl.FramebufferTextureLayer(gl.FRAMEBUFFER, gl.STENCIL_ATTACHMENT, texture.id, int32(attachment.MipmapLayer), int32(attachment.Depth))
			default:
				gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.STENCIL_ATTACHMENT, gl.TEXTURE_2D, texture.id, int32(attachment.MipmapLayer))
			}
		}
	}

	if len(drawBuffers) > 0 {
		gl.DrawBuffers(int32(len(drawBuffers)), &drawBuffers[0])
	} else {
		gl.DrawBuffers(int32(len(drawBuffers)), nil)
	}

	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if status != gl.FRAMEBUFFER_COMPLETE {
		logger.Error("Framebuffer (%v) is incomplete", info.Label)
	}

	result := &Framebuffer{
		label:             info.Label,
		id:                id,
		activeDrawBuffers: activeDrawBuffers,
	}
	framebuffers.Track(result.id, result)
	return result
}

var DefaultFramebuffer = func() *Framebuffer {
	result := &Framebuffer{
		label:             "default",
		id:                0,
		activeDrawBuffers: [4]bool{true, false, false, false},
	}
	framebuffers.Track(result.id, result)
	return result
}()

type Framebuffer struct {
	render.FramebufferMarker

	label             string
	id                uint32
	activeDrawBuffers [4]bool
}

func (f *Framebuffer) Label() string {
	return f.label
}

func (f *Framebuffer) Release() {
	framebuffers.Release(f.id)
	gl.DeleteFramebuffers(1, &f.id)
	f.id = 0
	f.activeDrawBuffers = [4]bool{}
}

func DetermineContentFormat(framebuffer render.Framebuffer) render.DataFormat {
	fb := framebuffer.(*Framebuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.id)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

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
