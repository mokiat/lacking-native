package game

import (
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/glsl"
	"github.com/mokiat/lacking/game/graphics/lsl"
	renderapi "github.com/mokiat/lacking/render"
)

func NewShaderBuilder() graphics.ShaderBuilder {
	return &shaderBuilder{}
}

type shaderBuilder struct{}

func (b *shaderBuilder) BuildCode(constraints graphics.ShaderConstraints, shader *lsl.Shader) renderapi.ProgramCode {
	translator := glsl.NewTranslator("410", false)
	result := translator.Translate(shader, constraints)

	// fmt.Println("--------------- VERTEX SHADER ----------------")
	// fmt.Println()
	// fmt.Println(result.VertexCode)
	// fmt.Println()
	// fmt.Println("-----------------------------------------------")

	// fmt.Println("--------------- FRAGMENT SHADER ----------------")
	// fmt.Println()
	// fmt.Println(result.FragmentCode)
	// fmt.Println()
	// fmt.Println("-----------------------------------------------")

	return result
}
