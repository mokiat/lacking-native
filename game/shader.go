package game

import (
	"fmt"

	"github.com/mokiat/lacking-native/internal/shader"
	"github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking/game/graphics"
	renderapi "github.com/mokiat/lacking/render"
)

var construct = shader.Load(
	shader.Common(),
	shader.Game(),
)

func NewShaderCollection() graphics.ShaderCollection {
	return graphics.ShaderCollection{
		AmbientLightSet:     newAmbientLightShaderSet,
		PointLightSet:       newPointLightShaderSet,
		SpotLightSet:        newSpotLightShaderSet,
		DirectionalLightSet: newDirectionalLightShaderSet,
		DebugSet:            newDebugShaderSet,
		ExposureSet:         newExposureShaderSet,
		BloomDownsampleSet:  newBloomDownsampleShaderSet,
		BloomBlurSet:        newBloomBlurShaderSet,
		PostprocessingSet:   newPostprocessingShaderSet,
	}
}

func newAmbientLightShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("ambient_light.vert.glsl", struct{}{}),
		FragmentCode: construct("ambient_light.frag.glsl", struct{}{}),
	}
}

func newPointLightShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("point_light.vert.glsl", struct{}{}),
		FragmentCode: construct("point_light.frag.glsl", struct{}{}),
	}
}

func newSpotLightShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("spot_light.vert.glsl", struct{}{}),
		FragmentCode: construct("spot_light.frag.glsl", struct{}{}),
	}
}

func newDirectionalLightShaderSet() renderapi.ProgramCode {
	var settings struct {
		UseShadowMapping bool
	}
	settings.UseShadowMapping = true // TODO
	return render.ProgramCode{
		VertexCode:   construct("directional_light.vert.glsl", settings),
		FragmentCode: construct("directional_light.frag.glsl", settings),
	}
}

func newDebugShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("debug.vert.glsl", struct{}{}),
		FragmentCode: construct("debug.frag.glsl", struct{}{}),
	}
}

func newExposureShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("exposure.vert.glsl", struct{}{}),
		FragmentCode: construct("exposure.frag.glsl", struct{}{}),
	}
}

func newBloomDownsampleShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("bloom_downsample.vert.glsl", struct{}{}),
		FragmentCode: construct("bloom_downsample.frag.glsl", struct{}{}),
	}
}

func newBloomBlurShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("bloom_blur.vert.glsl", struct{}{}),
		FragmentCode: construct("bloom_blur.frag.glsl", struct{}{}),
	}
}

func newPostprocessingShaderSet(cfg graphics.PostprocessingShaderConfig) renderapi.ProgramCode {
	var settings struct {
		UseReinhard    bool
		UseExponential bool
		UseBloom       bool
	}
	switch cfg.ToneMapping {
	case graphics.ReinhardToneMapping:
		settings.UseReinhard = true
	case graphics.ExponentialToneMapping:
		settings.UseExponential = true
	default:
		panic(fmt.Errorf("unknown tone mapping mode: %s", cfg.ToneMapping))
	}
	settings.UseBloom = cfg.Bloom
	return render.ProgramCode{
		VertexCode:   construct("postprocess.vert.glsl", settings),
		FragmentCode: construct("postprocess.frag.glsl", settings),
	}
}
