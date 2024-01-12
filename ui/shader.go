package ui

import (
	"github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking-native/shader"
	renderapi "github.com/mokiat/lacking/render"
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

func newShadedShapeShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   shader.RunTemplate("shaded_shape.vert.glsl", struct{}{}),
		FragmentCode: shader.RunTemplate("shaded_shape.frag.glsl", struct{}{}),
	}
}

func newBlankShapeShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   shader.RunTemplate("blank_shape.vert.glsl", struct{}{}),
		FragmentCode: shader.RunTemplate("blank_shape.frag.glsl", struct{}{}),
	}
}

func newContourShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   shader.RunTemplate("contour.vert.glsl", struct{}{}),
		FragmentCode: shader.RunTemplate("contour.frag.glsl", struct{}{}),
	}
}

func newTextShaderSet() renderapi.ProgramCode {
	return render.ProgramCode{
		VertexCode:   shader.RunTemplate("text.vert.glsl", struct{}{}),
		FragmentCode: shader.RunTemplate("text.frag.glsl", struct{}{}),
	}
}
