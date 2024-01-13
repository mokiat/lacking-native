package game

import (
	"fmt"

	"github.com/mokiat/lacking-native/internal/shader"
	"github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/shading"
	renderapi "github.com/mokiat/lacking/render"
)

var construct = shader.Load(
	shader.Common(),
	shader.Game(),
)

func NewShaderCollection() graphics.ShaderCollection {
	return graphics.ShaderCollection{
		BuildForward:        buildForward,
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

func buildForward(cfg graphics.MeshConfig, fn shading.ForwardFunc) renderapi.ProgramCode {
	var vertexSettings struct {
		UseArmature bool
	}
	if cfg.HasArmature {
		vertexSettings.UseArmature = true
	}

	var fragmentSettings struct {
		Lines []string
	}

	palette := shading.NewForwardPalette()
	fn(palette)

	paramIndex := 0
	paramNames := make(map[shading.Parameter]string)
	for _, node := range palette.Nodes() {
		switch node := node.(type) {
		case *shading.ConstVec4Node:
			outParam := node.OutVec()
			if !outParam.IsUsed() {
				continue
			}
			paramName := fmt.Sprintf("param%d", func() int {
				paramIndex++
				return paramIndex
			}())
			paramNames[outParam] = paramName

			fragmentSettings.Lines = append(fragmentSettings.Lines,
				fmt.Sprintf("vec4 %s = vec4(%f, %f, %f, %f);", paramName, node.X(), node.Y(), node.Z(), node.W()),
			)

		case *shading.MulVec4Node:
			outParam := node.OutVec()
			if !outParam.IsUsed() {
				continue
			}
			paramName := fmt.Sprintf("param%d", func() int {
				paramIndex++
				return paramIndex
			}())
			paramNames[outParam] = paramName

			fragmentSettings.Lines = append(fragmentSettings.Lines,
				fmt.Sprintf("vec4 %s = %f * %s;", paramName, node.Ratio(), paramNames[node.InVec()]),
			)

		case *shading.OutputColorNode:
			paramName, ok := paramNames[node.InColor()]
			if !ok {
				panic(fmt.Errorf("could not find param name for param of type %T", node.InColor()))
			}
			fragmentSettings.Lines = append(fragmentSettings.Lines,
				fmt.Sprintf("fbColor0Out = %s;", paramName),
			)
		default:
			panic(fmt.Errorf("unknown node type: %T", node))
		}
	}

	return render.ProgramCode{
		VertexCode:   construct("custom.vert.glsl", vertexSettings),
		FragmentCode: construct("custom.frag.glsl", fragmentSettings),
	}
}

func newShadowMappingSet(cfg graphics.ShadowMappingShaderConfig) renderapi.ProgramCode {
	var settings struct {
		UseArmature bool
	}
	if cfg.HasArmature {
		settings.UseArmature = true
	}
	return render.ProgramCode{
		VertexCode:   construct("shadow.vert.glsl", settings),
		FragmentCode: construct("shadow.frag.glsl", settings),
	}
}

func newPBRGeometrySet(cfg graphics.PBRGeometryShaderConfig) renderapi.ProgramCode {
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
	return render.ProgramCode{
		VertexCode:   construct("pbr_geometry.vert.glsl", settings),
		FragmentCode: construct("pbr_geometry.frag.glsl", settings),
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

func newSkyboxShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("skybox.vert.glsl", struct{}{}),
		FragmentCode: construct("skybox.frag.glsl", struct{}{}),
	}
}

func newSkycolorShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("skycolor.vert.glsl", struct{}{}),
		FragmentCode: construct("skycolor.frag.glsl", struct{}{}),
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

func newPostprocessingShaderSet(cfg graphics.PostprocessingShaderConfig) renderapi.ProgramCode {
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
	return render.ProgramCode{
		VertexCode:   construct("postprocess.vert.glsl", settings),
		FragmentCode: construct("postprocess.frag.glsl", settings),
	}
}
