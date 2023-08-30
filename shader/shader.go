package shader

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed *.glsl
var sources embed.FS

var rootTemplate = template.Must(template.
	New("root").
	Delims("/*", "*/").
	ParseFS(sources, "*.glsl"),
)

var buffer = new(bytes.Buffer)

func RunTemplate(name string, data any) string {
	tmpl := rootTemplate.Lookup(name)
	if tmpl == nil {
		panic(fmt.Errorf("template %q not found", name))
	}

	buffer.Reset()
	if err := tmpl.Execute(buffer, data); err != nil {
		panic(fmt.Errorf("template exec error: %w", err))
	}
	return buffer.String()
}
