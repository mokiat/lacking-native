package ui

import (
	"github.com/mokiat/lacking-native/shader"
	"github.com/mokiat/lacking/ui"
)

func NewShaderCollection() ui.ShaderCollection {
	return ui.ShaderCollection{
		ShapeShadedSet: newShadedShapeShaderSet,
		ShapeBlankSet:  newBlankShapeShaderSet,
		ContourSet:     newContourShaderSet,
		TextSet:        newTextShaderSet,
	}
}

func newShadedShapeShaderSet() ui.ShaderSet {
	return ui.ShaderSet{
		VertexShader:   shader.RunTemplate("shaded_shape.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("shaded_shape.frag.glsl", struct{}{}),
	}
}

func newBlankShapeShaderSet() ui.ShaderSet {
	return ui.ShaderSet{
		VertexShader:   shader.RunTemplate("blank_shape.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("blank_shape.frag.glsl", struct{}{}),
	}
}

func newContourShaderSet() ui.ShaderSet {
	return ui.ShaderSet{
		VertexShader:   shader.RunTemplate("contour.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("contour.frag.glsl", struct{}{}),
	}
}

func newTextShaderSet() ui.ShaderSet {
	return ui.ShaderSet{
		VertexShader:   shader.RunTemplate("text.vert.glsl", struct{}{}),
		FragmentShader: shader.RunTemplate("text.frag.glsl", struct{}{}),
	}
}
