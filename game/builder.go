package game

import (
	"github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
	renderapi "github.com/mokiat/lacking/render"
)

func NewShaderBuilder() graphics.ShaderBuilder {
	return &shaderBuilder{}
}

type shaderBuilder struct{}

func (b *shaderBuilder) BuildCode(constraints graphics.ShaderConstraints, shader *lsl.Shader) renderapi.ProgramCode {
	settings := shaderSettings{
		LoadGeometryPreset: constraints.LoadGeometryPreset,
		LoadSkyPreset:      constraints.LoadSkyPreset,

		HasOutput0: constraints.HasOutput0,
		HasOutput1: constraints.HasOutput1,
		HasOutput2: constraints.HasOutput2,
		HasOutput3: constraints.HasOutput3,

		HasCoords:         constraints.HasCoords,
		HasNormals:        constraints.HasNormals,
		HasTangents:       constraints.HasTangents,
		HasTexCoords:      constraints.HasTexCoords,
		HasVertexColoring: constraints.HasVertexColors,
		HasArmature:       constraints.HasArmature,
	}
	return render.ProgramCode{
		VertexCode:   b.buildVertexCode(shader, settings),
		FragmentCode: b.buildFragmentCode(shader, settings),
	}
}

func (b *shaderBuilder) buildVertexCode(shader *lsl.Shader, settings shaderSettings) string {
	translator := newTranslator()
	translation := translator.Translate(shader, "#vertex")
	return construct("custom.vert.glsl", struct {
		shaderSettings
		translatorOutput
	}{
		shaderSettings:   settings,
		translatorOutput: translation,
	})
}

func (b *shaderBuilder) buildFragmentCode(shader *lsl.Shader, settings shaderSettings) string {
	translator := newTranslator()
	translation := translator.Translate(shader, "#fragment")
	return construct("custom.frag.glsl", struct {
		shaderSettings
		translatorOutput
	}{
		shaderSettings:   settings,
		translatorOutput: translation,
	})
}

type shaderSettings struct {
	LoadGeometryPreset bool
	LoadSkyPreset      bool

	HasOutput0 bool
	HasOutput1 bool
	HasOutput2 bool
	HasOutput3 bool

	HasCoords         bool
	HasNormals        bool
	HasTangents       bool
	HasTexCoords      bool
	HasVertexColoring bool
	HasArmature       bool
}
