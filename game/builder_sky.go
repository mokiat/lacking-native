package game

import (
	"github.com/mokiat/lacking-native/internal/shader/translator"
	"github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
	renderapi "github.com/mokiat/lacking/render"
)

func (b *shaderBuilder) BuildSkyCode(constraints graphics.SkyConstraints, shader *lsl.Shader) renderapi.ProgramCode {
	// TODO: Verify matching varyings between vertex and fragment
	return render.ProgramCode{
		VertexCode:   b.buildSkyVertexCode(constraints, shader),
		FragmentCode: b.buildSkyFragmentCode(constraints, shader),
	}
}

func (b *shaderBuilder) buildSkyVertexCode(_ graphics.SkyConstraints, _ *lsl.Shader) string {
	var vertexSettings struct{}

	// TODO: Add support for varying

	// TODO: Add support for position output

	// TODO: Do actual shader translation
	return construct("custom_sky.vert.glsl", vertexSettings)
}

func (b *shaderBuilder) buildSkyFragmentCode(_ graphics.SkyConstraints, shader *lsl.Shader) string {
	var fragmentSettings struct {
		TextureLines []string
		UniformLines []string
		VaryingLines []string
		CodeLines    []string
	}

	lines := translator.Translate(shader, translator.ShaderStageFragment)
	fragmentSettings.TextureLines = lines.TextureLines
	fragmentSettings.UniformLines = lines.UniformLines
	fragmentSettings.VaryingLines = lines.VaryingLines
	fragmentSettings.CodeLines = lines.CodeLines
	return construct("custom_sky.frag.glsl", fragmentSettings)
}
