package ui

import (
	"github.com/mokiat/lacking-native/internal/shader"
	"github.com/mokiat/lacking-native/render"
	renderapi "github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/ui"
)

var construct = shader.Load(
	shader.Common(),
	shader.UI(),
)

func NewShaderCollection() ui.ShaderCollection {
	return ui.ShaderCollection{
		ShapeShadedSet: newShadedShapeShaderSet,
		ShapeBlankSet:  newBlankShapeShaderSet,
		ContourSet:     newContourShaderSet,
		TextSet:        newTextShaderSet,
	}
}

func newShadedShapeShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("shaded_shape.vert.glsl", struct{}{}),
		FragmentCode: construct("shaded_shape.frag.glsl", struct{}{}),
	}
}

func newBlankShapeShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("blank_shape.vert.glsl", struct{}{}),
		FragmentCode: construct("blank_shape.frag.glsl", struct{}{}),
	}
}

func newContourShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("contour.vert.glsl", struct{}{}),
		FragmentCode: construct("contour.frag.glsl", struct{}{}),
	}
}

func newTextShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   construct("text.vert.glsl", struct{}{}),
		FragmentCode: construct("text.frag.glsl", struct{}{}),
	}
}
