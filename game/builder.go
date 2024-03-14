package game

import (
	"github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/shading"
	renderapi "github.com/mokiat/lacking/render"
)

func NewShaderBuilder() graphics.ShaderBuilder {
	return &shaderBuilder{}
}

type shaderBuilder struct{}

func (b *shaderBuilder) BuildGeometryCode(constraints graphics.GeometryConstraints, vertex shading.GenericBuilderFunc, fragment shading.GenericBuilderFunc) renderapi.ProgramCode {
	panic("TODO")
}

func (b *shaderBuilder) BuildShadowCode(constraints graphics.ShadowConstraints, vertex shading.GenericBuilderFunc, fragment shading.GenericBuilderFunc) renderapi.ProgramCode {
	panic("TODO")
}

func (b *shaderBuilder) BuildForwardCode(constraints graphics.ForwardConstraints, vertex shading.GenericBuilderFunc, fragment shading.GenericBuilderFunc) renderapi.ProgramCode {
	// TODO: Verify matching varyings between vertex and fragment
	return render.ProgramCode{
		VertexCode:   b.buildForwardVertexCode(constraints, vertex),
		FragmentCode: b.buildForwardFragmentCode(constraints, fragment),
	}
}

func (b *shaderBuilder) buildForwardVertexCode(constraints graphics.ForwardConstraints, _ shading.GenericBuilderFunc) string {
	var vertexSettings struct {
		UseArmature bool
	}
	if constraints.HasArmature {
		vertexSettings.UseArmature = true
	}

	// TODO: Add support for varying

	// TODO: Add support for position output

	// TODO: Do actual shader translation
	return construct("custom_forward.vert.glsl", vertexSettings)
}

func (b *shaderBuilder) buildForwardFragmentCode(_ graphics.ForwardConstraints, builderFunc shading.GenericBuilderFunc) string {
	var fragmentSettings struct {
		UniformLines []string
		VaryingLines []string
		CodeLines    []string
		// TODO: Add more fields here, specific to forward fragment shader based
		// on the constraints.
	}

	translator := newTranslator()
	lines := translator.Translate(builderFunc)

	fragmentSettings.UniformLines = lines.UniformLines
	fragmentSettings.VaryingLines = lines.VaryingLines
	fragmentSettings.CodeLines = lines.CodeLines

	return construct("custom_forward.frag.glsl", fragmentSettings)
}
