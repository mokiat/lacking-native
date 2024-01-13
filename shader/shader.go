package shader

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"text/template"
)

// TODO: Move to internal package

//go:embed common/*.glsl ui/*.glsl game/*.glsl
var sources embed.FS

func shaderDir(name string) fs.FS {
	subDir, err := fs.Sub(sources, name)
	if err != nil {
		panic(fmt.Errorf("error opening %q shaders: %w", name, err))
	}
	return subDir
}

// Common returns a filesystem containing common shaders.
func Common() fs.FS {
	return shaderDir("common")
}

// UI returns a filesystem containing UI shaders.
func UI() fs.FS {
	return shaderDir("ui")
}

// Game returns a filesystem containing game shaders.
func Game() fs.FS {
	return shaderDir("game")
}

// ConstructFunc creates a shader from a template with the given name and data.
type ConstructFunc func(name string, data any) string

// Load loads all shaders from the given filesystems and returns a
// ConstructFunc that can be used to construct shaders from the loaded
// templates.
func Load(sources ...fs.FS) ConstructFunc {
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
