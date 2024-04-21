package translator

import (
	"fmt"

	"github.com/mokiat/lacking/game/graphics/lsl"
)

type ShaderStage string

const (
	ShaderStageVertex   ShaderStage = "vertex"
	ShaderStageFragment ShaderStage = "fragment"
)

func Translate(shader *lsl.Shader, stage ShaderStage) Output {
	return newTranslator(stage).Translate(shader)
}

func newTranslator(stage ShaderStage) *translator {
	return &translator{
		stage:       stage,
		nameMapping: make(map[string]string),
	}
}

type translator struct {
	stage ShaderStage

	nameMapping map[string]string
	nameIndex   uint32

	textureLines []string
	uniformLines []string
	varyingLines []string
	codeLines    []string
}

func (t *translator) Translate(shader *lsl.Shader) Output {
	for _, declaration := range shader.Declarations {
		switch decl := declaration.(type) {
		case *lsl.TextureBlockDeclaration:
			t.translateTextureBlock(decl)
		case *lsl.UniformBlockDeclaration:
			t.translateUniformBlock(decl)
		case *lsl.VaryingBlockDeclaration:
			t.translateVaryingBlock(decl)
		case *lsl.FunctionDeclaration:
			t.translateFunction(decl)
		default:
			panic(fmt.Errorf("unknown declaration type: %T", declaration))
		}
	}
	return Output{
		TextureLines: t.textureLines,
		UniformLines: t.uniformLines,
		VaryingLines: t.varyingLines,
		CodeLines:    t.codeLines,
	}
}

func (t *translator) translateTextureBlock(decl *lsl.TextureBlockDeclaration) {
	for _, field := range decl.Fields {
		name := t.translateName(field.Name)
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
		name := t.translateName(field.Name)
		var uniformLine string
		switch field.Type {
		case lsl.TypeNameFloat:
			uniformLine = fmt.Sprintf("float %s;", name)
		case lsl.TypeNameVec2:
			uniformLine = fmt.Sprintf("vec2 %s;", name)
		case lsl.TypeNameVec3:
			uniformLine = fmt.Sprintf("vec3 %s;", name)
		case lsl.TypeNameVec4:
			uniformLine = fmt.Sprintf("vec4 %s;", name)
		default:
			panic(fmt.Errorf("unknown uniform type: %s", field.Type))
		}
		t.uniformLines = append(t.uniformLines, uniformLine)
	}
}

func (t *translator) translateVaryingBlock(decl *lsl.VaryingBlockDeclaration) {
	for _, field := range decl.Fields {
		name := t.translateName(field.Name)
		var varyingLine string
		switch field.Type {
		case lsl.TypeNameFloat:
			varyingLine = fmt.Sprintf("float %s;", name)
		case lsl.TypeNameVec2:
			varyingLine = fmt.Sprintf("vec2 %s;", name)
		case lsl.TypeNameVec3:
			varyingLine = fmt.Sprintf("vec3 %s;", name)
		case lsl.TypeNameVec4:
			varyingLine = fmt.Sprintf("vec4 %s;", name)
		default:
			panic(fmt.Errorf("unknown uniform type: %s", field.Type))
		}
		t.varyingLines = append(t.varyingLines, varyingLine)
	}
}

func (t *translator) translateFunction(decl *lsl.FunctionDeclaration) {
	var specialFunctionName string
	switch t.stage {
	case ShaderStageVertex:
		specialFunctionName = "#vertex"
	case ShaderStageFragment:
		specialFunctionName = "#fragment"
	}
	if decl.Name == specialFunctionName {
		for _, statement := range decl.Body {
			t.translateStatement(statement)
		}
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
			return "color"
		}
		if target.Name == "#metallic" {
			return "metallic"
		}
		if target.Name == "#roughness" {
			return "roughness"
		}
		if mappedName, ok := t.nameMapping[target.Name]; ok {
			return mappedName
		}
		panic(fmt.Errorf("unknown target: %s", target.Name))

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
		return "varyingDirection" // FIXME: Should be handled by the sky shader rewriter
	}
	if identifier.Name == "#uv" {
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
	case "rgb":
		return t.translateRGBCall(call)
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

func (t *translator) translateRGBCall(call *lsl.FunctionCall) string {
	isArgumentTypes := func(_ ...string) bool {
		return true // FIXME
	}
	switch {
	case isArgumentTypes(lsl.TypeNameSamplerCube, lsl.TypeNameVec3):
		return fmt.Sprintf("(%s).xyz",
			t.translateExpression(call.Arguments[0]),
		)
	default:
		panic(fmt.Errorf("unknown texture call overload: %s", call.Name))
	}
}

func (t *translator) translateName(name string) string {
	if mappedName, ok := t.nameMapping[name]; ok {
		return mappedName
	}
	mappedName := t.nextName()
	t.nameMapping[name] = mappedName
	return mappedName
}

func (t *translator) nextName() string {
	name := fmt.Sprintf("userIdentifier%d", t.nameIndex)
	t.nameIndex++
	return name
}
