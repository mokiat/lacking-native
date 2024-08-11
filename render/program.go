package render

import (
	"strings"

	"github.com/mokiat/lacking/render"
)

// ProgramCode is an implementation of render.ProgramCode that can be used
// with this render API implementation.
type ProgramCode struct {
	render.ProgramCodeMarker

	// VertexCode specifies the vertex shader code.
	VertexCode string

	// FragmentCode specifies the fragment shader code.
	FragmentCode string
}

func (p ProgramCode) String() string {
	vertex := strings.ReplaceAll(p.VertexCode, "\n\n", "\n")
	fragment := strings.ReplaceAll(p.FragmentCode, "\n\n", "\n")
	return vertex + "\n\n---\n\n" + fragment
}
