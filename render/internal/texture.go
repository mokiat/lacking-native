package internal

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewColorTexture2D(info render.ColorTexture2DInfo) *Texture {
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	levels := glMipmapLevels(info.Width, info.Height, info.GenerateMipmaps)
	internalFormat := glInternalFormat(info.Format, info.GammaCorrection)
	gl.TexStorage2D(gl.TEXTURE_2D, levels, internalFormat, int32(info.Width), int32(info.Height))

	if info.Data != nil {
		dataFormat := glDataFormat(info.Format)
		componentType := glDataComponentType(info.Format)
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(info.Width), int32(info.Height), dataFormat, componentType, gl.Ptr(info.Data))
		if info.GenerateMipmaps {
			gl.GenerateMipmap(gl.TEXTURE_2D)
		}
	}

	result := &Texture{
		id:   id,
		kind: gl.TEXTURE_2D,
	}
	textures.Track(id, result)
	return result
}

func NewDepthTexture2D(info render.DepthTexture2DInfo) *Texture {
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	if info.Comparable {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_COMPARE_MODE, gl.COMPARE_REF_TO_TEXTURE)
		gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.DEPTH_COMPONENT32F, int32(info.Width), int32(info.Height))
	} else {
		gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.DEPTH_COMPONENT24, int32(info.Width), int32(info.Height))
	}

	result := &Texture{
		id:   id,
		kind: gl.TEXTURE_2D,
	}
	textures.Track(id, result)
	return result
}

func NewStencilTexture2D(info render.StencilTexture2DInfo) *Texture {
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.DEPTH24_STENCIL8, int32(info.Width), int32(info.Height))

	result := &Texture{
		id:   id,
		kind: gl.TEXTURE_2D,
	}
	textures.Track(id, result)
	return result
}

func NewDepthStencilTexture2D(info render.DepthStencilTexture2DInfo) *Texture {
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.DEPTH24_STENCIL8, int32(info.Width), int32(info.Height))

	result := &Texture{
		id:   id,
		kind: gl.TEXTURE_2D,
	}
	textures.Track(id, result)
	return result
}

func NewColorTextureCube(info render.ColorTextureCubeInfo) *Texture {
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, id)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	levels := glMipmapLevels(info.Dimension, info.Dimension, info.GenerateMipmaps)
	internalFormat := glInternalFormat(info.Format, info.GammaCorrection)
	gl.TexStorage2D(gl.TEXTURE_CUBE_MAP, levels, internalFormat, int32(info.Dimension), int32(info.Dimension))

	dataFormat := glDataFormat(info.Format)
	componentType := glDataComponentType(info.Format)
	if info.RightSideData != nil {
		gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X, 0, 0, 0, int32(info.Dimension), int32(info.Dimension), dataFormat, componentType, gl.Ptr(&info.RightSideData[0]))
	}
	if info.LeftSideData != nil {
		gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_X, 0, 0, 0, int32(info.Dimension), int32(info.Dimension), dataFormat, componentType, gl.Ptr(&info.LeftSideData[0]))
	}
	if info.BottomSideData != nil {
		gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_Y, 0, 0, 0, int32(info.Dimension), int32(info.Dimension), dataFormat, componentType, gl.Ptr(&info.BottomSideData[0]))
	}
	if info.TopSideData != nil {
		gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_Y, 0, 0, 0, int32(info.Dimension), int32(info.Dimension), dataFormat, componentType, gl.Ptr(&info.TopSideData[0]))
	}
	if info.FrontSideData != nil {
		gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_Z, 0, 0, 0, int32(info.Dimension), int32(info.Dimension), dataFormat, componentType, gl.Ptr(&info.FrontSideData[0]))
	}
	if info.BackSideData != nil {
		gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_Z, 0, 0, 0, int32(info.Dimension), int32(info.Dimension), dataFormat, componentType, gl.Ptr(&info.BackSideData[0]))
	}

	// TODO: Move as separate command
	// if info.Mipmapping {
	// 	gl.GenerateTextureMipmap(id)
	// }

	result := &Texture{
		id:   id,
		kind: gl.TEXTURE_CUBE_MAP,
	}
	textures.Track(id, result)
	return result
}

type Texture struct {
	render.TextureMarker
	id   uint32
	kind uint32
}

func (t *Texture) Release() {
	textures.Release(t.id)
	gl.DeleteTextures(1, &t.id)
	t.id = 0
	t.kind = 0
}

func NewSampler(info render.SamplerInfo) *Sampler {
	var id uint32
	gl.GenSamplers(1, &id)
	gl.SamplerParameteri(id, gl.TEXTURE_WRAP_S, glWrap(info.Wrapping))
	gl.SamplerParameteri(id, gl.TEXTURE_WRAP_T, glWrap(info.Wrapping))
	gl.SamplerParameteri(id, gl.TEXTURE_WRAP_R, glWrap(info.Wrapping))

	gl.SamplerParameteri(id, gl.TEXTURE_MIN_FILTER, glFilter(info.Filtering, info.Mipmapping))
	gl.SamplerParameteri(id, gl.TEXTURE_MAG_FILTER, glFilter(info.Filtering, false)) // no mipmaps when magnification

	if info.Comparison.Specified {
		gl.SamplerParameteri(id, gl.TEXTURE_COMPARE_MODE, gl.COMPARE_REF_TO_TEXTURE)
		gl.SamplerParameteri(id, gl.TEXTURE_COMPARE_FUNC, int32(glEnumFromComparison(info.Comparison.Value)))
	}

	result := &Sampler{
		id: id,
	}
	samplers.Track(result.id, result)
	return result
}

type Sampler struct {
	render.SamplerMarker
	id uint32
}

func (s *Sampler) Release() {
	samplers.Release(s.id)
	gl.DeleteSamplers(1, &s.id)
	s.id = 0
}

func glWrap(wrap render.WrapMode) int32 {
	switch wrap {
	case render.WrapModeClamp:
		return gl.CLAMP_TO_EDGE
	case render.WrapModeRepeat:
		return gl.REPEAT
	case render.WrapModeMirroredRepeat:
		return gl.MIRRORED_REPEAT
	default:
		return gl.CLAMP_TO_EDGE
	}
}

func glFilter(filter render.FilterMode, mipmaps bool) int32 {
	switch filter {
	case render.FilterModeNearest:
		if mipmaps {
			return gl.NEAREST_MIPMAP_NEAREST
		}
		return gl.NEAREST
	case render.FilterModeLinear, render.FilterModeAnisotropic:
		if mipmaps {
			return gl.LINEAR_MIPMAP_LINEAR
		}
		return gl.LINEAR
	default:
		return gl.NEAREST
	}
}

func glMipmapLevels(width, height uint32, mipmapping bool) int32 {
	if !mipmapping {
		return 1
	}
	count := int32(1)
	for width > 1 || height > 1 {
		width /= 2
		height /= 2
		count++
	}
	return count
}

func glInternalFormat(format render.DataFormat, gammaCorrection bool) uint32 {
	switch format {
	case render.DataFormatRGBA8:
		if gammaCorrection {
			return gl.SRGB8_ALPHA8
		}
		return gl.RGBA8
	case render.DataFormatRGBA16F:
		return gl.RGBA16F
	case render.DataFormatRGBA32F:
		return gl.RGBA32F
	default:
		return gl.RGBA8
	}
}

func glDataFormat(format render.DataFormat) uint32 {
	switch format {
	default:
		return gl.RGBA
	}
}

func glDataComponentType(format render.DataFormat) uint32 {
	switch format {
	case render.DataFormatRGBA8:
		return gl.UNSIGNED_BYTE
	case render.DataFormatRGBA16F:
		return gl.HALF_FLOAT
	case render.DataFormatRGBA32F:
		return gl.FLOAT
	default:
		return gl.UNSIGNED_BYTE
	}
}
