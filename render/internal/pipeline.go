package internal

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewPipeline(info render.PipelineInfo) *Pipeline {
	intProgram := info.Program.(*Program)
	intVertexArray := info.VertexArray.(*VertexArray)

	pipeline := &Pipeline{
		ProgramID: intProgram.id,
		VertexArray: CommandBindVertexArray{
			VertexArrayID: intVertexArray.id,
			IndexFormat:   intVertexArray.indexFormat,
		},
	}

	switch info.Topology {
	case render.TopologyPoints:
		pipeline.Topology.Topology = gl.POINTS
	case render.TopologyLineStrip:
		pipeline.Topology.Topology = gl.LINE_STRIP
	case render.TopologyLineList:
		pipeline.Topology.Topology = gl.LINES
	case render.TopologyTriangleStrip:
		pipeline.Topology.Topology = gl.TRIANGLE_STRIP
	case render.TopologyTriangleList:
		pipeline.Topology.Topology = gl.TRIANGLES
	case render.TopologyTriangleFan:
		pipeline.Topology.Topology = gl.TRIANGLE_FAN
	}

	switch info.Culling {
	case render.CullModeNone:
		pipeline.CullTest.Enabled = false
	case render.CullModeBack:
		pipeline.CullTest.Enabled = true
		pipeline.CullTest.Face = gl.BACK
	case render.CullModeFront:
		pipeline.CullTest.Enabled = true
		pipeline.CullTest.Face = gl.FRONT
	case render.CullModeFrontAndBack:
		pipeline.CullTest.Enabled = true
		pipeline.CullTest.Face = gl.FRONT_AND_BACK
	}

	switch info.FrontFace {
	case render.FaceOrientationCCW:
		pipeline.FrontFace.Orientation = gl.CCW
	case render.FaceOrientationCW:
		pipeline.FrontFace.Orientation = gl.CW
	}

	pipeline.DepthTest.Enabled = info.DepthTest
	pipeline.DepthWrite.Enabled = info.DepthWrite
	pipeline.DepthComparison.Mode = glEnumFromComparison(info.DepthComparison)

	pipeline.StencilTest.Enabled = info.StencilTest

	pipeline.StencilOpFront.Face = gl.FRONT
	pipeline.StencilOpFront.StencilFail = glEnumFromStencilOp(info.StencilFront.StencilFailOp)
	pipeline.StencilOpFront.DepthFail = glEnumFromStencilOp(info.StencilFront.DepthFailOp)
	pipeline.StencilOpFront.Pass = glEnumFromStencilOp(info.StencilFront.PassOp)

	pipeline.StencilOpBack.Face = gl.BACK
	pipeline.StencilOpBack.StencilFail = glEnumFromStencilOp(info.StencilBack.StencilFailOp)
	pipeline.StencilOpBack.DepthFail = glEnumFromStencilOp(info.StencilBack.DepthFailOp)
	pipeline.StencilOpBack.Pass = glEnumFromStencilOp(info.StencilBack.PassOp)

	pipeline.StencilFuncFront.Face = gl.FRONT
	pipeline.StencilFuncFront.Func = glEnumFromComparison(info.StencilFront.Comparison)
	pipeline.StencilFuncFront.Ref = info.StencilFront.Reference
	pipeline.StencilFuncFront.Mask = info.StencilFront.ComparisonMask

	pipeline.StencilFuncBack.Face = gl.BACK
	pipeline.StencilFuncBack.Func = glEnumFromComparison(info.StencilBack.Comparison)
	pipeline.StencilFuncBack.Ref = info.StencilBack.Reference
	pipeline.StencilFuncBack.Mask = info.StencilBack.ComparisonMask

	pipeline.StencilMaskFront.Face = gl.FRONT
	pipeline.StencilMaskFront.Mask = info.StencilFront.WriteMask

	pipeline.StencilMaskBack.Face = gl.BACK
	pipeline.StencilMaskBack.Mask = info.StencilBack.WriteMask

	pipeline.ColorWrite.Mask = info.ColorWrite

	pipeline.BlendEnabled = info.BlendEnabled
	pipeline.BlendColor.Color = info.BlendColor

	pipeline.BlendEquation.ModeRGB = glEnumFromBlendOp(info.BlendOpColor)
	pipeline.BlendEquation.ModeAlpha = glEnumFromBlendOp(info.BlendOpAlpha)

	pipeline.BlendFunc.SourceFactorRGB = glEnumFromBlendFactor(info.BlendSourceColorFactor)
	pipeline.BlendFunc.DestinationFactorRGB = glEnumFromBlendFactor(info.BlendDestinationColorFactor)
	pipeline.BlendFunc.SourceFactorAlpha = glEnumFromBlendFactor(info.BlendSourceAlphaFactor)
	pipeline.BlendFunc.DestinationFactorAlpha = glEnumFromBlendFactor(info.BlendDestinationAlphaFactor)

	return pipeline
}

func glEnumFromComparison(comparison render.Comparison) uint32 {
	switch comparison {
	case render.ComparisonNever:
		return gl.NEVER
	case render.ComparisonLess:
		return gl.LESS
	case render.ComparisonEqual:
		return gl.EQUAL
	case render.ComparisonLessOrEqual:
		return gl.LEQUAL
	case render.ComparisonGreater:
		return gl.GREATER
	case render.ComparisonNotEqual:
		return gl.NOTEQUAL
	case render.ComparisonGreaterOrEqual:
		return gl.GEQUAL
	case render.ComparisonAlways:
		return gl.ALWAYS
	default:
		panic(fmt.Errorf("unknown comparison: %d", comparison))
	}
}

func glEnumFromStencilOp(op render.StencilOperation) uint32 {
	switch op {
	case render.StencilOperationKeep:
		return gl.KEEP
	case render.StencilOperationZero:
		return gl.ZERO
	case render.StencilOperationReplace:
		return gl.REPLACE
	case render.StencilOperationIncrease:
		return gl.INCR
	case render.StencilOperationIncreaseWrap:
		return gl.INCR_WRAP
	case render.StencilOperationDecrease:
		return gl.DECR
	case render.StencilOperationDecreaseWrap:
		return gl.DECR_WRAP
	case render.StencilOperationInvert:
		return gl.INVERT
	default:
		panic(fmt.Errorf("unknown op: %d", op))
	}
}

func glEnumFromBlendOp(op render.BlendOperation) uint32 {
	switch op {
	case render.BlendOperationAdd:
		return gl.FUNC_ADD
	case render.BlendOperationSubtract:
		return gl.FUNC_SUBTRACT
	case render.BlendOperationReverseSubtract:
		return gl.FUNC_REVERSE_SUBTRACT
	case render.BlendOperationMin:
		return gl.MIN
	case render.BlendOperationMax:
		return gl.MAX
	default:
		panic(fmt.Errorf("unknown op: %d", op))
	}
}

func glEnumFromBlendFactor(factor render.BlendFactor) uint32 {
	switch factor {
	case render.BlendFactorZero:
		return gl.ZERO
	case render.BlendFactorOne:
		return gl.ONE
	case render.BlendFactorSourceColor:
		return gl.SRC_COLOR
	case render.BlendFactorOneMinusSourceColor:
		return gl.ONE_MINUS_SRC_COLOR
	case render.BlendFactorDestinationColor:
		return gl.DST_COLOR
	case render.BlendFactorOneMinusDestinationColor:
		return gl.ONE_MINUS_DST_COLOR
	case render.BlendFactorSourceAlpha:
		return gl.SRC_ALPHA
	case render.BlendFactorOneMinusSourceAlpha:
		return gl.ONE_MINUS_SRC_ALPHA
	case render.BlendFactorDestinationAlpha:
		return gl.DST_ALPHA
	case render.BlendFactorOneMinusDestinationAlpha:
		return gl.ONE_MINUS_DST_ALPHA
	case render.BlendFactorConstantColor:
		return gl.CONSTANT_COLOR
	case render.BlendFactorOneMinusConstantColor:
		return gl.ONE_MINUS_CONSTANT_COLOR
	case render.BlendFactorConstantAlpha:
		return gl.CONSTANT_ALPHA
	case render.BlendFactorOneMinusConstantAlpha:
		return gl.ONE_MINUS_CONSTANT_ALPHA
	case render.BlendFactorSourceAlphaSaturate:
		return gl.SRC_ALPHA_SATURATE
	default:
		panic(fmt.Errorf("unknown factor: %d", factor))
	}
}

type Pipeline struct {
	render.PipelineMarker
	ProgramID        uint32
	Topology         CommandTopology
	CullTest         CommandCullTest
	FrontFace        CommandFrontFace
	DepthTest        CommandDepthTest
	DepthWrite       CommandDepthWrite
	DepthComparison  CommandDepthComparison
	StencilTest      CommandStencilTest
	StencilOpFront   CommandStencilOperation
	StencilOpBack    CommandStencilOperation
	StencilFuncFront CommandStencilFunc
	StencilFuncBack  CommandStencilFunc
	StencilMaskFront CommandStencilMask
	StencilMaskBack  CommandStencilMask
	ColorWrite       CommandColorWrite
	BlendEnabled     bool
	BlendColor       CommandBlendColor
	BlendEquation    CommandBlendEquation
	BlendFunc        CommandBlendFunc
	VertexArray      CommandBindVertexArray
}

func (p *Pipeline) Release() {
}
