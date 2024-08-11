package game

import (
	"github.com/mokiat/lacking-native/internal/shader/translator"
	"github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/lsl"
	renderapi "github.com/mokiat/lacking/render"
)

func (b *shaderBuilder) BuildGeometryCode(constraints graphics.GeometryConstraints, shader *lsl.Shader) renderapi.ProgramCode {
	// TODO: Verify matching varyings between vertex and fragment
	return render.ProgramCode{
		VertexCode:   b.buildGeometryVertexCode(constraints, shader),
		FragmentCode: b.buildGeometryFragmentCode(constraints, shader),
	}
}

func (b *shaderBuilder) buildGeometryVertexCode(constraints graphics.GeometryConstraints, _ *lsl.Shader) string {
	var settings struct {
		UseArmature       bool
		UseTangents       bool
		UseNormals        bool
		UseTexCoords      bool
		UseVertexColoring bool
	}
	if constraints.HasArmature {
		settings.UseArmature = true
	}
	if constraints.HasNormals {
		settings.UseNormals = true
	}
	if constraints.HasTexCoords {
		settings.UseTexCoords = true
	}
	if constraints.HasVertexColors {
		settings.UseVertexColoring = true
	}
	// TODO: Add support for varying

	// TODO: Add support for position output

	// TODO: Do actual shader translation
	result := construct("custom_geometry.vert.glsl", settings)
	// fmt.Println("Vertex code:\n", result)
	return result
}

func (b *shaderBuilder) buildGeometryFragmentCode(constraints graphics.GeometryConstraints, shader *lsl.Shader) string {
	var settings struct {
		UseArmature       bool
		UseNormals        bool
		UseTexCoords      bool
		UseVertexColoring bool

		TextureLines []string
		UniformLines []string
		VaryingLines []string
		CodeLines    []string
		// TODO: Add more fields here, specific to geometry fragment shader based
		// on the constraints.
	}
	if constraints.HasArmature {
		settings.UseArmature = true
	}
	if constraints.HasNormals {
		settings.UseNormals = true
	}
	if constraints.HasTexCoords {
		settings.UseTexCoords = true
	}
	if constraints.HasVertexColors {
		settings.UseVertexColoring = true
	}

	lines := translator.Translate(shader, translator.ShaderStageFragment)
	settings.TextureLines = lines.TextureLines
	settings.UniformLines = lines.UniformLines
	settings.VaryingLines = lines.VaryingLines
	settings.CodeLines = lines.CodeLines

	result := construct("custom_geometry.frag.glsl", settings)
	// fmt.Println("Fragment code:\n", result)
	return result
}
