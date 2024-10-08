package internal

import (
	"errors"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

type ProgramInfo struct {
	Label           string
	VertexCode      string
	FragmentCode    string
	TextureBindings []render.TextureBinding
	UniformBindings []render.UniformBinding
}

func NewProgram(info ProgramInfo) *Program {
	if glLogger.IsDebugEnabled() {
		defer trackError("Error creating program (%v)", info.Label)()
	}

	vertexShader := newVertexShader(info.Label, info.VertexCode)
	defer vertexShader.Release()

	fragmentShader := newFragmentShader(info.Label, info.FragmentCode)
	defer fragmentShader.Release()

	program := &Program{
		label: info.Label,
		id:    gl.CreateProgram(),
	}

	gl.AttachShader(program.id, vertexShader.id)
	defer gl.DetachShader(program.id, vertexShader.id)
	gl.AttachShader(program.id, fragmentShader.id)
	defer gl.DetachShader(program.id, fragmentShader.id)

	if err := program.link(); err != nil {
		logger.Error("Program (%v) link error: %v", info.Label, err)
	}

	if len(info.TextureBindings) > 0 {
		gl.UseProgram(program.id)
		for _, binding := range info.TextureBindings {
			name := binding.Name + "\x00"
			location := gl.GetUniformLocation(program.id, StrPtr(name))
			runtime.KeepAlive(name)
			if location != -1 {
				gl.Uniform1i(location, int32(binding.Index))
			}
		}
		gl.UseProgram(0)
	}

	for _, binding := range info.UniformBindings {
		name := binding.Name + "\x00"
		location := gl.GetUniformBlockIndex(program.id, StrPtr(name))
		runtime.KeepAlive(name)
		if location != gl.INVALID_INDEX {
			gl.UniformBlockBinding(program.id, location, uint32(binding.Index))
		}
	}

	return program
}

type Program struct {
	render.ProgramMarker

	label string
	id    uint32
}

func (p *Program) Label() string {
	return p.label
}

func (p *Program) Release() {
	gl.DeleteProgram(p.id)
	p.id = 0
}

func (p *Program) link() error {
	gl.LinkProgram(p.id)
	if !p.isLinkSuccessful() {
		return errors.New(p.getInfoLog())
	}
	return nil
}

func (p *Program) isLinkSuccessful() bool {
	var status int32
	gl.GetProgramiv(p.id, gl.LINK_STATUS, &status)
	return status != gl.FALSE
}

func (p *Program) getInfoLog() string {
	var logLength int32
	gl.GetProgramiv(p.id, gl.INFO_LOG_LENGTH, &logLength)

	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetProgramInfoLog(p.id, logLength, nil, StrPtr(log))
	runtime.KeepAlive(log)
	return log
}
