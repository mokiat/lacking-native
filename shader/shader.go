package shader

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"text/template"
)

// TODO: Move to internal package

//go:embed *.glsl common/*.glsl ui/*.glsl game/*.glsl
var sources embed.FS

func shaderDir(name string) fs.FS {
	subDir, err := fs.Sub(sources, name)
	if err != nil {
		panic(fmt.Errorf("error opening %q shaders: %w", name, err))
	}
	return subDir
}

func Common() fs.FS {
	return shaderDir("common")
}

func UI() fs.FS {
	return shaderDir("ui")
}

func Game() fs.FS {
	return shaderDir("game")
}

var rootTemplate = template.Must(template.
	New("root").
	Delims("/*", "*/").
	ParseFS(sources, "*.glsl"),
)

var buffer = new(bytes.Buffer)

// Deprecated: Use ShaderSource instead.
func RunTemplate(name string, data any) string {
	tmpl := rootTemplate.Lookup(name)
	if tmpl == nil {
		panic(fmt.Errorf("template %q not found", name))
	}

	buffer.Reset()
	if err := tmpl.Execute(buffer, data); err != nil {
		panic(fmt.Errorf("template %q exec error: %w", name, err))
	}
	return buffer.String()
}

// TODO: Move to internal package

type ConstructShaderFunc func(name string, data any) string

func LoadShaders(sources ...fs.FS) ConstructShaderFunc {
	rootTemplate := template.New("root").Delims("/*", "*/")
	for _, source := range sources {
		rootTemplate = template.Must(rootTemplate.ParseFS(source, "*.glsl"))
	}

	buffer := new(bytes.Buffer)
	return func(name string, data any) string {
		tmpl := rootTemplate.Lookup(name)
		if tmpl == nil {
			panic(fmt.Errorf("template %q not found", name))
		}

		buffer.Reset()
		if err := tmpl.Execute(buffer, data); err != nil {
			panic(fmt.Errorf("template %q exec error: %w", name, err))
		}
		return buffer.String()
	}
}
