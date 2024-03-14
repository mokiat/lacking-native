package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/graphics/shading"
)

func newTranslator() *translator {
	return &translator{}
}

type translator struct{}

func (t *translator) Translate(builderFunc shading.GenericBuilderFunc) translatorOutput {
	builder := shading.NewBuilder()
	builderFunc(builder)

	var output translatorOutput

	variableNames := make(map[shading.VariableIndex]string)

	for _, uniformIndex := range builder.Uniforms() {
		uniform := builder.Variable(uniformIndex)
		name := fmt.Sprintf("materialUniform%d", len(variableNames))
		variableNames[uniformIndex] = name

		var codeLine string
		switch uniform.Type {
		case shading.VariableTypeVec1:
			codeLine = fmt.Sprintf("uniform float %s;", name)
		case shading.VariableTypeVec2:
			codeLine = fmt.Sprintf("uniform vec2 %s;", name)
		case shading.VariableTypeVec3:
			codeLine = fmt.Sprintf("uniform vec3 %s;", name)
		case shading.VariableTypeVec4:
			codeLine = fmt.Sprintf("uniform vec4 %s;", name)
		default:
			panic(fmt.Errorf("unknown uniform type: %d", uniform.Type))
		}
		output.UniformLines = append(output.UniformLines, codeLine)
	}

	ensureVariable := func(index shading.VariableIndex) string {
		name, ok := variableNames[index]
		if !ok {
			name = fmt.Sprintf("var%d", len(variableNames))
			variableNames[index] = name

			variable := builder.Variable(index)

			var codeLine string
			switch variable.Type {
			case shading.VariableTypeFloat32Value:
				codeLine = fmt.Sprintf("float %s = %f;", name, variable.AsFloat32())
			case shading.VariableTypeVec1:
				codeLine = fmt.Sprintf("float %s = 0.0;", name)
			case shading.VariableTypeVec2:
				codeLine = fmt.Sprintf("vec2 %s = vec2(0.0, 0.0);", name)
			case shading.VariableTypeVec3:
				codeLine = fmt.Sprintf("vec3 %s = vec3(0.0, 0.0, 0.0);", name)
			case shading.VariableTypeVec4:
				codeLine = fmt.Sprintf("vec4 %s = vec4(0.0, 0.0, 0.0, 0.0);", name)
			default:
				panic(fmt.Errorf("unknown variable type: %d", variable.Type))
			}
			output.CodeLines = append(output.CodeLines, codeLine)
		}
		return name
	}

	for _, operation := range builder.Operations() {
		inputParams := builder.InputParameters(operation)
		outputParams := builder.OutputParameters(operation)
		switch operation.Type {
		case shading.OperationDefineVec1:
			target := outputParams[0]
			ensureVariable(target)
		case shading.OperationDefineVec2:
			target := outputParams[0]
			ensureVariable(target)
		case shading.OperationDefineVec3:
			target := outputParams[0]
			ensureVariable(target)
		case shading.OperationDefineVec4:
			target := outputParams[0]
			ensureVariable(target)
		case shading.OperationTypeAssignVec1:
			target := variableNames[outputParams[0]]
			x := variableNames[inputParams[0]]
			codeLine := fmt.Sprintf("%s = %s;", target, x)
			output.CodeLines = append(output.CodeLines, codeLine)
		case shading.OperationTypeAssignVec2:
			target := variableNames[outputParams[0]]
			x := variableNames[inputParams[0]]
			y := variableNames[inputParams[1]]
			codeLine := fmt.Sprintf("%s = vec2(%s, %s);", target, x, y)
			output.CodeLines = append(output.CodeLines, codeLine)
		case shading.OperationTypeAssignVec3:
			target := variableNames[outputParams[0]]
			x := variableNames[inputParams[0]]
			y := variableNames[inputParams[1]]
			z := variableNames[inputParams[2]]
			codeLine := fmt.Sprintf("%s = vec3(%s, %s, %s);", target, x, y, z)
			output.CodeLines = append(output.CodeLines, codeLine)
		case shading.OperationTypeAssignVec4:
			target := variableNames[outputParams[0]]
			x := variableNames[inputParams[0]]
			y := variableNames[inputParams[1]]
			z := variableNames[inputParams[2]]
			w := variableNames[inputParams[3]]
			codeLine := fmt.Sprintf("%s = vec4(%s, %s, %s, %s);", target, x, y, z, w)
			output.CodeLines = append(output.CodeLines, codeLine)
		case shading.OperationTypeForwardOutputColor:
			color := variableNames[inputParams[0]]
			codeLine := fmt.Sprintf("fbColor0Out = %s;", color)
			output.CodeLines = append(output.CodeLines, codeLine)
		case shading.OperationTypeForwardAlphaDiscard:
			alpha := variableNames[inputParams[0]]
			threshold := variableNames[inputParams[1]]
			codeLine := fmt.Sprintf("if (%s < %s) discard;", alpha, threshold)
			output.CodeLines = append(output.CodeLines, codeLine)
		default:
			panic(fmt.Errorf("unknown operation type: %d", operation.Type))
		}
	}

	return output
}

type translatorOutput struct {
	UniformLines []string
	VaryingLines []string
	CodeLines    []string
}
