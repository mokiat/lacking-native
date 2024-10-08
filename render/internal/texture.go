package internal

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewColorTexture2D(info render.ColorTexture2DInfo) *Texture {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating color texture 2D (%v)", info.Label)()
	}

	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	width := info.MipmapLayers[0].Width
	height := info.MipmapLayers[0].Height
	internalFormat := glInternalFormat(info.Format, info.GammaCorrection)
	if levels := int32(len(info.MipmapLayers)); levels > 1 {
		gl.TexStorage2D(gl.TEXTURE_2D, levels, internalFormat, int32(width), int32(height))
	} else {
		levels := glMipmapLevels(width, height, info.GenerateMipmaps)
		gl.TexStorage2D(gl.TEXTURE_2D, levels, internalFormat, int32(width), int32(height))
	}

	dataFormat := glDataFormat(info.Format)
	componentType := glDataComponentType(info.Format)
	for i, mipmapLayer := range info.MipmapLayers {
		if mipmapLayer.Data != nil {
			gl.TexSubImage2D(gl.TEXTURE_2D, int32(i), 0, 0, int32(mipmapLayer.Width), int32(mipmapLayer.Height), dataFormat, componentType, gl.Ptr(&mipmapLayer.Data[0]))
		}
	}

	if info.GenerateMipmaps && len(info.MipmapLayers) == 1 {
		// TODO: Move as separate command
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}

	result := &Texture{
		label:  info.Label,
		id:     id,
		kind:   gl.TEXTURE_2D,
		width:  width,
		height: height,
	}
	textures.Track(id, result)
	return result
}

func NewDepthTexture2D(info render.DepthTexture2DInfo) *Texture {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating depth texture 2D (%v)", info.Label)()
	}

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
		label:  info.Label,
		id:     id,
		kind:   gl.TEXTURE_2D,
		width:  info.Width,
		height: info.Height,
	}
	textures.Track(id, result)
	return result
}

func NewDepthTexture2DArray(info render.DepthTexture2DArrayInfo) *Texture {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating array depth texture 2D (%v)", info.Label)()
	}

	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, id)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	if info.Comparable {
		gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_COMPARE_MODE, gl.COMPARE_REF_TO_TEXTURE)
		gl.TexImage3D(gl.TEXTURE_2D_ARRAY, 0, gl.DEPTH_COMPONENT32F, int32(info.Width), int32(info.Height), int32(info.Layers), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	} else {
		gl.TexImage3D(gl.TEXTURE_2D_ARRAY, 0, gl.DEPTH_COMPONENT24, int32(info.Width), int32(info.Height), int32(info.Layers), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	}

	result := &Texture{
		label:  info.Label,
		id:     id,
		kind:   gl.TEXTURE_2D_ARRAY,
		width:  info.Width,
		height: info.Height,
		depth:  info.Layers,
	}
	textures.Track(id, result)
	return result
}

func NewStencilTexture2D(info render.StencilTexture2DInfo) *Texture {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating stencil texture 2D (%v)", info.Label)()
	}

	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.DEPTH24_STENCIL8, int32(info.Width), int32(info.Height))

	result := &Texture{
		label:  info.Label,
		id:     id,
		kind:   gl.TEXTURE_2D,
		width:  info.Width,
		height: info.Height,
	}
	textures.Track(id, result)
	return result
}

func NewDepthStencilTexture2D(info render.DepthStencilTexture2DInfo) *Texture {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating depth-stencil texture 2D (%v)", info.Label)()
	}

	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.DEPTH24_STENCIL8, int32(info.Width), int32(info.Height))

	result := &Texture{
		label:  info.Label,
		id:     id,
		kind:   gl.TEXTURE_2D,
		width:  info.Width,
		height: info.Height,
	}
	textures.Track(id, result)
	return result
}

func NewColorTextureCube(info render.ColorTextureCubeInfo) *Texture {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating color texture cube (%v)", info.Label)()
	}

	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, id)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	dimension := info.MipmapLayers[0].Dimension
	internalFormat := glInternalFormat(info.Format, info.GammaCorrection)
	if levels := int32(len(info.MipmapLayers)); levels > 1 {
		gl.TexStorage2D(gl.TEXTURE_CUBE_MAP, levels, internalFormat, int32(dimension), int32(dimension))
	} else {
		levels := glMipmapLevels(dimension, dimension, info.GenerateMipmaps)
		gl.TexStorage2D(gl.TEXTURE_CUBE_MAP, levels, internalFormat, int32(dimension), int32(dimension))
	}

	dataFormat := glDataFormat(info.Format)
	componentType := glDataComponentType(info.Format)
	for i, mipmapLayer := range info.MipmapLayers {
		if mipmapLayer.RightSideData != nil {
			gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X, int32(i), 0, 0, int32(mipmapLayer.Dimension), int32(mipmapLayer.Dimension), dataFormat, componentType, gl.Ptr(&mipmapLayer.RightSideData[0]))
		}
		if mipmapLayer.LeftSideData != nil {
			gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_X, int32(i), 0, 0, int32(mipmapLayer.Dimension), int32(mipmapLayer.Dimension), dataFormat, componentType, gl.Ptr(&mipmapLayer.LeftSideData[0]))
		}
		if mipmapLayer.BottomSideData != nil {
			gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_Y, int32(i), 0, 0, int32(mipmapLayer.Dimension), int32(mipmapLayer.Dimension), dataFormat, componentType, gl.Ptr(&mipmapLayer.BottomSideData[0]))
		}
		if mipmapLayer.TopSideData != nil {
			gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_Y, int32(i), 0, 0, int32(mipmapLayer.Dimension), int32(mipmapLayer.Dimension), dataFormat, componentType, gl.Ptr(&mipmapLayer.TopSideData[0]))
		}
		if mipmapLayer.FrontSideData != nil {
			gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_Z, int32(i), 0, 0, int32(mipmapLayer.Dimension), int32(mipmapLayer.Dimension), dataFormat, componentType, gl.Ptr(&mipmapLayer.FrontSideData[0]))
		}
		if mipmapLayer.BackSideData != nil {
			gl.TexSubImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_Z, int32(i), 0, 0, int32(mipmapLayer.Dimension), int32(mipmapLayer.Dimension), dataFormat, componentType, gl.Ptr(&mipmapLayer.BackSideData[0]))
		}
	}

	if info.GenerateMipmaps && len(info.MipmapLayers) == 1 {
		// TODO: Move as separate command
		gl.GenerateMipmap(gl.TEXTURE_CUBE_MAP)
	}

	result := &Texture{
		label:  info.Label,
		id:     id,
		kind:   gl.TEXTURE_CUBE_MAP,
		width:  dimension,
		height: dimension,
		depth:  dimension,
	}
	textures.Track(id, result)
	return result
}

type Texture struct {
	render.TextureMarker

	label  string
	id     uint32
	kind   uint32
	width  uint32
	height uint32
	depth  uint32
}

func (t *Texture) Label() string {
	return t.label
}

func (t *Texture) Width() uint32 {
	return t.width
}

func (t *Texture) Height() uint32 {
	return t.height
}

func (t *Texture) Depth() uint32 {
	return t.depth
}

func (t *Texture) Release() {
	textures.Release(t.id)
	gl.DeleteTextures(1, &t.id)
	t.id = 0
	t.kind = 0
}

func NewSampler(info render.SamplerInfo) *Sampler {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating sampler (%v)", info.Label)()
	}

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
		label: info.Label,
		id:    id,
	}
	samplers.Track(result.id, result)
	return result
}

type Sampler struct {
	render.SamplerMarker

	label string
	id    uint32
}

func (s *Sampler) Label() string {
	return s.label
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
