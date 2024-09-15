package internal

import (
	"errors"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func newVertexShader(programLabel, sourceCode string) *Shader {
	shader := &Shader{
		id: gl.CreateShader(gl.VERTEX_SHADER),
	}
	shader.setSourceCode(sourceCode)
	if err := shader.compile(); err != nil {
		logger.Error("Vertex shader (%v) compilation error: %v", programLabel, err)
	}
	return shader
}

func newFragmentShader(programLabel, sourceCode string) *Shader {
	shader := &Shader{
		id: gl.CreateShader(gl.FRAGMENT_SHADER),
	}
	shader.setSourceCode(sourceCode)
	if err := shader.compile(); err != nil {
		logger.Error("Fragment shader (%v) compilation error: %v", programLabel, err)
	}
	return shader
}

type Shader struct {
	id uint32
}

func (s *Shader) Release() {
	gl.DeleteShader(s.id)
	s.id = 0
}

func (s *Shader) setSourceCode(code string) {
	terminatedCode := code + "\x00"
	sources, free := gl.Strs(terminatedCode)
	defer free()
	gl.ShaderSource(s.id, 1, sources, nil)
	runtime.KeepAlive(terminatedCode)
}

func (s *Shader) compile() error {
	gl.CompileShader(s.id)
	if !s.isCompileSuccessful() {
		return errors.New(s.getInfoLog())
	}
	return nil
}

func (s *Shader) isCompileSuccessful() bool {
	var status int32
	gl.GetShaderiv(s.id, gl.COMPILE_STATUS, &status)
	return status != gl.FALSE
}

func (s *Shader) getInfoLog() string {
	var logLength int32
	gl.GetShaderiv(s.id, gl.INFO_LOG_LENGTH, &logLength)

	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetShaderInfoLog(s.id, logLength, nil, StrPtr(log))
	runtime.KeepAlive(log)
	return log
}
