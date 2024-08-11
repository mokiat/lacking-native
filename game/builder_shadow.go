package game

import (
	"github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
	renderapi "github.com/mokiat/lacking/render"
)

func (b *shaderBuilder) BuildShadowCode(constraints graphics.ShadowConstraints, shader *lsl.Shader) renderapi.ProgramCode {
	// TODO: Verify matching varyings between vertex and fragment
	return render.ProgramCode{
		VertexCode:   b.buildShadowVertexCode(constraints, shader),
		FragmentCode: b.buildShadowFragmentCode(constraints, shader),
	}
}

func (b *shaderBuilder) buildShadowVertexCode(constraints graphics.ShadowConstraints, _ *lsl.Shader) string {
	// FIXME: Use actual shader
	var settings struct {
		UseArmature bool
	}
	if constraints.HasArmature {
		settings.UseArmature = true
	}
	return construct("shadow.vert.glsl", settings)
}

func (b *shaderBuilder) buildShadowFragmentCode(constraints graphics.ShadowConstraints, _ *lsl.Shader) string {
	// FIXME: Use actual shader
	var settings struct {
		UseArmature bool
	}
	if constraints.HasArmature {
		settings.UseArmature = true
	}
	return construct("shadow.frag.glsl", settings)
}
