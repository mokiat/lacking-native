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
	var qualifier string
	switch t.stage {
	case ShaderStageVertex:
		qualifier = "out"
	case ShaderStageFragment:
		qualifier = "in"
	}
	for _, field := range decl.Fields {
		name := t.translateName(field.Name)
		var varyingLine string
		switch field.Type {
		case lsl.TypeNameFloat:
			varyingLine = fmt.Sprintf("smooth %s float %s;", qualifier, name)
		case lsl.TypeNameVec2:
			varyingLine = fmt.Sprintf("smooth %s vec2 %s;", qualifier, name)
		case lsl.TypeNameVec3:
			varyingLine = fmt.Sprintf("smooth %s vec3 %s;", qualifier, name)
		case lsl.TypeNameVec4:
			varyingLine = fmt.Sprintf("smooth %s vec4 %s;", qualifier, name)
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
	case *lsl.Discard:
		t.translateDiscard()
	case *lsl.Assignment:
		t.translateAssignment(stmt)
	case *lsl.VariableDeclaration:
		t.translateVariableDeclaration(stmt)
	case *lsl.Conditional:
		t.translateConditional(stmt)
	default:
		panic(fmt.Errorf("unknown statement type: %T", statement))
	}
}

func (t *translator) translateDiscard() {
	t.codeLines = append(t.codeLines, "discard;")
}

func (t *translator) translateConditional(conditional *lsl.Conditional) {
	t.codeLines = append(t.codeLines, "if (")
	t.codeLines = append(t.codeLines, t.translateExpression(conditional.Condition))
	t.codeLines = append(t.codeLines, ") {")
	for _, statement := range conditional.Then {
		t.translateStatement(statement)
	}
	switch {
	case conditional.ElseIf != nil:
		t.codeLines = append(t.codeLines, "} else")
		t.translateConditional(conditional.ElseIf) // TODO: Verify if this works, since it puts it on a new line
	case conditional.Else != nil:
		t.codeLines = append(t.codeLines, "} else {")
		for _, statement := range conditional.Else {
			t.translateStatement(statement)
		}
		t.codeLines = append(t.codeLines, "}")
	default:
		t.codeLines = append(t.codeLines, "}")
	}
}

func (t *translator) translateAssignment(assignment *lsl.Assignment) {
	target := t.translateExpression(assignment.Target)
	expression := t.translateExpression(assignment.Expression)
	switch assignment.Operator {
	case "=":
		t.codeLines = append(t.codeLines, fmt.Sprintf("%s = %s;", target, expression))
	case "+=":
		t.codeLines = append(t.codeLines, fmt.Sprintf("%s += %s;", target, expression))
	case "*=":
		t.codeLines = append(t.codeLines, fmt.Sprintf("%s *= %s;", target, expression))
	default:
		panic(fmt.Errorf("unknown assignment operator: %s", assignment.Operator))
	}
}

func (t *translator) translateVariableDeclaration(declaration *lsl.VariableDeclaration) {
	varName := t.translateName(declaration.Name)
	varType := t.translateType(declaration.Type)
	if declaration.Assignment != nil {
		expression := t.translateExpression(declaration.Assignment)
		t.codeLines = append(t.codeLines, fmt.Sprintf("%s %s = %s;", varType, varName, expression))
	} else {
		t.codeLines = append(t.codeLines, fmt.Sprintf("%s %s;", varType, varName))
	}
}

func (t *translator) translateExpression(expression lsl.Expression) string {
	switch expr := expression.(type) {
	// TODO: Add support for bool literals
	case *lsl.Identifier:
		return t.translateIdentifier(expr)
	case *lsl.FieldIdentifier:
		return t.translateFieldIdentifier(expr)
	case *lsl.FunctionCall:
		return t.translateFunctionCall(expr)
	case *lsl.BinaryExpression:
		return t.translateBinaryExpression(expr)
	case *lsl.FloatLiteral:
		return t.translateFloatLiteral(expr)
	default:
		panic(fmt.Errorf("unknown expression type: %T", expression))
	}
}

func (t *translator) translateBinaryExpression(expression *lsl.BinaryExpression) string {
	left := t.translateExpression(expression.Left)
	right := t.translateExpression(expression.Right)
	operator := t.translateOperator(expression.Operator)
	return fmt.Sprintf("%s %s %s", left, operator, right)
}

func (t *translator) translateOperator(operator string) string {
	return operator
}

func (t *translator) translateIdentifier(identifier *lsl.Identifier) string {
	if identifier.Name == "#color" {
		return "color"
	}
	if identifier.Name == "#metallic" {
		return "metallic"
	}
	if identifier.Name == "#roughness" {
		return "roughness"
	}
	if identifier.Name == "#direction" {
		return "varyingDirection" // FIXME: Should be handled by the sky shader rewriter
	}
	if identifier.Name == "#normal" {
		return "normalInOut" // TODO: varyingNormal
	}
	if identifier.Name == "#uv" || identifier.Name == "#vertexUV" {
		return "texCoordInOut"
	}
	if identifier.Name == "#vertexColor" {
		return "colorInOut"
	}
	if identifier.Name == "#time" {
		return "lackingTime"
	}
	return t.translateName(identifier.Name)
}

func (t *translator) translateFieldIdentifier(identifier *lsl.FieldIdentifier) string {
	obj := t.translateIdentifier(&lsl.Identifier{
		Name: identifier.ObjName,
	})
	field := identifier.FieldName
	return fmt.Sprintf("%s.%s", obj, field)
}

func (t *translator) translateFloatLiteral(literal *lsl.FloatLiteral) string {
	return fmt.Sprintf("%f", literal.Value)
}

func (t *translator) translateFunctionCall(call *lsl.FunctionCall) string {
	switch call.Name {
	case "sample":
		return t.translateTextureCall(call)
	case "rgb":
		return t.translateRGBCall(call)
	case "cos":
		return t.translateCosCall(call)
	case "sin":
		return t.translateSinCall(call)
	case "mix":
		return t.translateMixCall(call)
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

func (t *translator) translateCosCall(call *lsl.FunctionCall) string {
	isArgumentTypes := func(_ ...string) bool {
		return true // FIXME
	}
	switch {
	case isArgumentTypes(lsl.TypeNameFloat):
		return fmt.Sprintf("cos(%s)",
			t.translateExpression(call.Arguments[0]),
		)
	default:
		panic(fmt.Errorf("unknown texture call overload: %s", call.Name))
	}
}

func (t *translator) translateSinCall(call *lsl.FunctionCall) string {
	isArgumentTypes := func(_ ...string) bool {
		return true // FIXME
	}
	switch {
	case isArgumentTypes(lsl.TypeNameFloat):
		return fmt.Sprintf("sin(%s)",
			t.translateExpression(call.Arguments[0]),
		)
	default:
		panic(fmt.Errorf("unknown texture call overload: %s", call.Name))
	}
}

func (t *translator) translateMixCall(call *lsl.FunctionCall) string {
	isArgumentTypes := func(_ ...string) bool {
		return true // FIXME
	}
	switch {
	case isArgumentTypes(lsl.TypeNameFloat, lsl.TypeNameFloat, lsl.TypeNameFloat):
		return fmt.Sprintf("mix(%s, %s, %s)",
			t.translateExpression(call.Arguments[0]),
			t.translateExpression(call.Arguments[1]),
			t.translateExpression(call.Arguments[2]),
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

func (t *translator) translateType(typeName string) string {
	switch typeName {
	case lsl.TypeNameFloat:
		return "float"
	case lsl.TypeNameVec2:
		return "vec2"
	case lsl.TypeNameVec3:
		return "vec3"
	case lsl.TypeNameVec4:
		return "vec4"
	default:
		panic(fmt.Errorf("unknown type: %s", typeName))
	}
}

func (t *translator) nextName() string {
	name := fmt.Sprintf("userIdentifier%d", t.nameIndex)
	t.nameIndex++
	return name
}
