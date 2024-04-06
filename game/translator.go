package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/graphics/lsl"
)

func newTranslator() *translator {
	return &translator{
		nameMapping: make(map[string]string),
	}
}

type translator struct {
	nameMapping map[string]string
	nameIndex   uint32

	textureLines []string
	uniformLines []string
	varyingLines []string
	codeLines    []string
}

func (t *translator) Translate(shader *lsl.Shader, funcName string) translatorOutput {
	for _, declaration := range shader.Declarations {
		switch decl := declaration.(type) {
		case *lsl.TextureBlockDeclaration:
			t.translateTextureBlock(decl)
		case *lsl.UniformBlockDeclaration:
			t.translateUniformBlock(decl)
		case *lsl.FunctionDeclaration:
			t.translateFunction(decl, funcName)
		default:
			panic(fmt.Errorf("unknown declaration type: %T", declaration))
		}
	}
	return translatorOutput{
		TextureLines: t.textureLines,
		UniformLines: t.uniformLines,
		VaryingLines: t.varyingLines,
		CodeLines:    t.codeLines,
	}
}

func (t *translator) translateTextureBlock(decl *lsl.TextureBlockDeclaration) {
	for _, field := range decl.Fields {
		name := t.translateFieldName(field.Name)
		var textureLine string
		switch field.Type {
		case lsl.TypeNameSampler2D:
			textureLine = fmt.Sprintf("uniform sampler2D %s;", name)
		case lsl.TypeNameSamplerCube:
			textureLine = fmt.Sprintf("uniform samplerCube %s;", name)
		default:
			panic(fmt.Errorf("unknown texture field type: %s", field.Type))
		}
		t.textureLines = append(t.textureLines, textureLine)
	}
}

func (t *translator) translateUniformBlock(decl *lsl.UniformBlockDeclaration) {
	for _, field := range decl.Fields {
		name := t.translateFieldName(field.Name)
		var uniformLine string
		switch field.Type {
		case lsl.TypeNameVec4:
			uniformLine = fmt.Sprintf("vec4 %s;", name)
		default:
			panic(fmt.Errorf("unknown uniform type: %s", field.Type))
		}
		t.uniformLines = append(t.uniformLines, uniformLine)
	}
}

func (t *translator) translateFunction(decl *lsl.FunctionDeclaration, funcName string) {
	if decl.Name != funcName {
		return
	}
	for _, statement := range decl.Body {
		t.translateStatement(statement)
	}
}

func (t *translator) translateStatement(statement lsl.Statement) {
	switch stmt := statement.(type) {
	case *lsl.Assignment:
		t.translateAssignment(stmt)
	default:
		panic(fmt.Errorf("unknown statement type: %T", statement))
	}
}

func (t *translator) translateAssignment(assignment *lsl.Assignment) {
	target := t.translateTarget(assignment.Target)
	expression := t.translateExpression(assignment.Expression)
	t.codeLines = append(t.codeLines, fmt.Sprintf("%s = %s;", target, expression))
}

func (t *translator) translateTarget(target lsl.Expression) string {
	switch target := target.(type) {
	case *lsl.Identifier:
		if target.Name == "#color" {
			return "fbColor0Out"
		}
		if mappedName, ok := t.nameMapping[target.Name]; ok {
			return mappedName
		}
		panic(fmt.Errorf("unknown target: %s", target))

	case *lsl.FieldIdentifier:
		panic("field identifiers are not supported")

	default:
		panic(fmt.Errorf("unknown target type: %T", target))
	}
}

func (t *translator) translateExpression(expression lsl.Expression) string {
	switch expr := expression.(type) {
	case *lsl.Identifier:
		return t.translateIdentifier(expr)
	case *lsl.FunctionCall:
		return t.translateFunctionCall(expr)
	default:
		panic(fmt.Errorf("unknown expression type: %T", expression))
	}
}

func (t *translator) translateIdentifier(identifier *lsl.Identifier) string {
	if identifier.Name == "#direction" {
		return "texCoordInOut"
	}
	if mappedName, ok := t.nameMapping[identifier.Name]; ok {
		return mappedName
	}
	panic(fmt.Errorf("unknown identifier: %s", identifier.Name))
}

func (t *translator) translateFunctionCall(call *lsl.FunctionCall) string {
	switch call.Name {
	case "sample":
		return t.translateTextureCall(call)
	default:
		panic(fmt.Errorf("unknown function call: %s", call.Name))
	}
}

func (t *translator) translateTextureCall(call *lsl.FunctionCall) string {
	isArgumentTypes := func(_ ...string) bool {
		return true // FIXME
	}
	switch {
	case isArgumentTypes(lsl.TypeNameSamplerCube, lsl.TypeNameVec3):
		return fmt.Sprintf("texture(%s, %s)",
			t.translateExpression(call.Arguments[0]),
			t.translateExpression(call.Arguments[1]),
		)
	default:
		panic(fmt.Errorf("unknown texture call overload: %s", call.Name))
	}
}

func (t *translator) translateFieldName(name string) string {
	if mappedName, ok := t.nameMapping[name]; ok {
		return mappedName
	}
	mappedName := t.nextName()
	t.nameMapping[name] = mappedName
	return mappedName
}

func (t *translator) nextName() string {
	name := fmt.Sprintf("variable%d", t.nameIndex)
	t.nameIndex++
	return name
}

type translatorOutput struct {
	TextureLines []string
	UniformLines []string
	VaryingLines []string
	CodeLines    []string
}
