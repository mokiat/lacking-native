package game

import (
	"github.com/mokiat/lacking/game/graphics"
)

func NewShaderBuilder() graphics.ShaderBuilder {
	return &shaderBuilder{}
}

type shaderBuilder struct{}
