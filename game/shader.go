package game

import (
	"fmt"

	"github.com/mokiat/lacking-native/shader"
	"github.com/mokiat/lacking/game/graphics"
)

func NewShaderCollection() graphics.ShaderCollection {
	return graphics.ShaderCollection{
		ShadowMappingSet:    newShadowMappingSet,
		PBRGeometrySet:      newPBRGeometrySet,
		DirectionalLightSet: newDirectionalLightShaderSet,
		AmbientLightSet:     newAmbientLightShaderSet,
		PointLightSet:       newPointLightShaderSet,
		SpotLightSet:        newSpotLightShaderSet,
		SkyboxSet:           newSkyboxShaderSet,
		SkycolorSet:         newSkycolorShaderSet,
		DebugSet:            newDebugShaderSet,
		ExposureSet:         newExposureShaderSet,
		PostprocessingSet:   newPostprocessingShaderSet,
	}
}

func newShadowMappingSet(cfg graphics.ShadowMappingShaderConfig) graphics.ShaderSet {
	var settings struct {
		UseArmature bool
	}
	if cfg.HasArmature {
		settings.UseArmature = true
	}
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("shadow.vert.glsl", settings),
		FragmentShader: shader.RunTemplate("shadow.frag.glsl", settings),
	}
}

func newPBRGeometrySet(cfg graphics.PBRGeometryShaderConfig) graphics.ShaderSet {
	var settings struct {
		UseArmature       bool
		UseAlphaTest      bool
		UseVertexColoring bool
		UseTexturing      bool
		UseAlbedoTexture  bool
	}
	if cfg.HasArmature {
		settings.UseArmature = true
	}
	if cfg.HasAlphaTesting {
		settings.UseAlphaTest = true
	}
	if cfg.HasVertexColors {
		settings.UseVertexColoring = true
	}
	if cfg.HasAlbedoTexture {
		settings.UseTexturing = true
		settings.UseAlbedoTexture = true
	}
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("pbr_geometry.vert.glsl", settings),
		FragmentShader: shader.RunTemplate("pbr_geometry.frag.glsl", settings),
	}
}

func newAmbientLightShaderSet() graphics.ShaderSet {
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("ambient_light.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("ambient_light.frag.glsl", struct{}{}),
	}
}

func newPointLightShaderSet() graphics.ShaderSet {
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("point_light.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("point_light.frag.glsl", struct{}{}),
	}
}

func newSpotLightShaderSet() graphics.ShaderSet {
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("spot_light.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("spot_light.frag.glsl", struct{}{}),
	}
}

func newDirectionalLightShaderSet() graphics.ShaderSet {
	var settings struct {
		UseShadowMapping bool
	}
	settings.UseShadowMapping = true // TODO
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("directional_light.vert.glsl", settings),
		FragmentShader: shader.RunTemplate("directional_light.frag.glsl", settings),
	}
}

func newSkyboxShaderSet() graphics.ShaderSet {
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("skybox.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("skybox.frag.glsl", struct{}{}),
	}
}

func newSkycolorShaderSet() graphics.ShaderSet {
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("skycolor.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("skycolor.frag.glsl", struct{}{}),
	}
}

func newDebugShaderSet() graphics.ShaderSet {
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("debug.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("debug.frag.glsl", struct{}{}),
	}
}

func newExposureShaderSet() graphics.ShaderSet {
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("exposure.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("exposure.frag.glsl", struct{}{}),
	}
}

func newPostprocessingShaderSet(cfg graphics.PostprocessingShaderConfig) graphics.ShaderSet {
	var settings struct {
		UseReinhard    bool
		UseExponential bool
	}
	switch cfg.ToneMapping {
	case graphics.ReinhardToneMapping:
		settings.UseReinhard = true
	case graphics.ExponentialToneMapping:
		settings.UseExponential = true
	default:
		panic(fmt.Errorf("unknown tone mapping mode: %s", cfg.ToneMapping))
	}
	return graphics.ShaderSet{
		VertexShader:   shader.RunTemplate("postprocess.vert.glsl", settings),
		FragmentShader: shader.RunTemplate("postprocess.frag.glsl", settings),
	}
}
