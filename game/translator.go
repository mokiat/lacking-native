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

	uniformLines []string
	varyingLines []string
	codeLines    []string
}

func (t *translator) Translate(shader *lsl.Shader, funcName string) translatorOutput {
	for _, declaration := range shader.Declarations {
		switch decl := declaration.(type) {
		case *lsl.UniformBlockDeclaration:
			t.translateUniformBlock(decl)
		case *lsl.FunctionDeclaration:
			t.translateFunction(decl, funcName)
		default:
			panic(fmt.Errorf("unknown declaration type: %T", declaration))
		}
	}
	return translatorOutput{
		UniformLines: t.uniformLines,
		VaryingLines: t.varyingLines,
		CodeLines:    t.codeLines,
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

func (t *translator) translateTarget(target string) string {
	if target == "#color" {
		return "fbColor0Out"
	}
	if mappedName, ok := t.nameMapping[target]; ok {
		return mappedName
	}
	panic(fmt.Errorf("unknown target: %s", target))
}

func (t *translator) translateExpression(expression lsl.Expression) string {
	switch expr := expression.(type) {
	case *lsl.Identifier:
		return t.translateIdentifier(expr)
	default:
		panic(fmt.Errorf("unknown expression type: %T", expression))
	}
}

func (t *translator) translateIdentifier(identifier *lsl.Identifier) string {
	if mappedName, ok := t.nameMapping[identifier.Name]; ok {
		return mappedName
	}
	panic(fmt.Errorf("unknown identifier: %s", identifier.Name))
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
	UniformLines []string
	VaryingLines []string
	CodeLines    []string
}
