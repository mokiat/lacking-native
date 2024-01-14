package render

import "github.com/mokiat/lacking/render"

// ProgramCode is an implementation of render.ProgramCode that can be used
// with this render API implementation.
type ProgramCode struct {
	render.ProgramCode

	// VertexCode specifies the vertex shader code.
	VertexCode string

	// FragmentCode specifies the fragment shader code.
	FragmentCode string
}
