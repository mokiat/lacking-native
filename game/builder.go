package game

import (
	"fmt"

	"github.com/mokiat/lacking-native/render"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/shading"
	renderapi "github.com/mokiat/lacking/render"
)

func NewShaderBuilder() graphics.ShaderBuilder {
	return &shaderBuilder{}
}

type shaderBuilder struct{}

func (b *shaderBuilder) BuildGeometryCode(constraints graphics.GeometryConstraints, vertex shading.GenericBuilderFunc, fragment shading.GenericBuilderFunc) renderapi.ProgramCode {
	panic("TODO")
}

func (b *shaderBuilder) BuildShadowCode(constraints graphics.ShadowConstraints, vertex shading.GenericBuilderFunc, fragment shading.GenericBuilderFunc) renderapi.ProgramCode {
	panic("TODO")
}

func (b *shaderBuilder) BuildForwardCode(constraints graphics.ForwardConstraints, vertex shading.GenericBuilderFunc, fragment shading.GenericBuilderFunc) renderapi.ProgramCode {
	// TODO: Verify matching varyings between vertex and fragment
	return render.ProgramCode{
		VertexCode:   b.buildForwardVertexCode(constraints, vertex),
		FragmentCode: b.buildForwardFragmentCode(constraints, fragment),
	}
}

func (b *shaderBuilder) buildForwardVertexCode(constraints graphics.ForwardConstraints, _ shading.GenericBuilderFunc) string {
	var vertexSettings struct {
		UseArmature bool
	}
	if constraints.HasArmature {
		vertexSettings.UseArmature = true
	}
	// TODO: Do actual shader translation
	return construct("custom.vert.glsl", vertexSettings)
}

func (b *shaderBuilder) buildForwardFragmentCode(_ graphics.ForwardConstraints, builderFunc shading.GenericBuilderFunc) string {
	var fragmentSettings struct {
		UniformLines []string
		Lines        []string
	}

	builder := shading.NewBuilder()
	builderFunc(builder)

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
		fragmentSettings.UniformLines = append(fragmentSettings.UniformLines, codeLine)
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
			fragmentSettings.Lines = append(fragmentSettings.Lines, codeLine)
		}
		return name
	}

	for _, operation := range builder.Operations() {
		params := builder.Parameters(operation)
		switch operation.Type {
		case shading.OperationDefineVec1:
			target := params[0]
			ensureVariable(target)
		case shading.OperationDefineVec2:
			target := params[0]
			ensureVariable(target)
		case shading.OperationDefineVec3:
			target := params[0]
			ensureVariable(target)
		case shading.OperationDefineVec4:
			target := params[0]
			ensureVariable(target)
		case shading.OperationTypeAssignVec1:
			target := variableNames[params[0]]
			x := variableNames[params[1]]
			codeLine := fmt.Sprintf("%s = %s;", target, x)
			fragmentSettings.Lines = append(fragmentSettings.Lines, codeLine)
		case shading.OperationTypeAssignVec2:
			target := variableNames[params[0]]
			x := variableNames[params[1]]
			y := variableNames[params[2]]
			codeLine := fmt.Sprintf("%s = vec2(%s, %s);", target, x, y)
			fragmentSettings.Lines = append(fragmentSettings.Lines, codeLine)
		case shading.OperationTypeAssignVec3:
			target := variableNames[params[0]]
			x := variableNames[params[1]]
			y := variableNames[params[2]]
			z := variableNames[params[3]]
			codeLine := fmt.Sprintf("%s = vec3(%s, %s, %s);", target, x, y, z)
			fragmentSettings.Lines = append(fragmentSettings.Lines, codeLine)
		case shading.OperationTypeAssignVec4:
			target := variableNames[params[0]]
			x := variableNames[params[1]]
			y := variableNames[params[2]]
			z := variableNames[params[3]]
			w := variableNames[params[4]]
			codeLine := fmt.Sprintf("%s = vec4(%s, %s, %s, %s);", target, x, y, z, w)
			fragmentSettings.Lines = append(fragmentSettings.Lines, codeLine)
		case shading.OperationTypeForwardOutputColor:
			color := variableNames[params[0]]
			codeLine := fmt.Sprintf("fbColor0Out = %s;", color)
			fragmentSettings.Lines = append(fragmentSettings.Lines, codeLine)
		case shading.OperationTypeForwardAlphaDiscard:
			alpha := variableNames[params[0]]
			threshold := variableNames[params[1]]
			codeLine := fmt.Sprintf("if (%s < %s) discard;", alpha, threshold)
			fragmentSettings.Lines = append(fragmentSettings.Lines, codeLine)
		default:
			panic(fmt.Errorf("unknown operation type: %d", operation.Type))
		}
	}
	return construct("custom.frag.glsl", fragmentSettings)
}
