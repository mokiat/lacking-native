package internal

import "github.com/mokiat/lacking/render"

type CommandKind uint8

const (
	CommandKindCopyFramebufferToBuffer CommandKind = iota
	CommandKindCopyFramebufferToTexture
	CommandKindBeginRenderPass
	CommandKindEndRenderPass
	CommandKindBindPipeline
	CommandKindUniformBufferUnit
	CommandKindTextureUnit
	CommandKindSamplerUnit
	CommandKindDraw
	CommandKindDrawIndexed
)

type CommandHeader struct {
	Kind CommandKind
}

type CommandCopyFramebufferToBuffer struct {
	BufferID     uint32
	X            int32
	Y            int32
	Width        int32
	Height       int32
	Format       uint32
	XType        uint32
	BufferOffset uint32
}

type CommandCopyFramebufferToTexture struct {
	TextureID       uint32
	TextureLevel    int32
	TextureX        int32
	TextureY        int32
	FramebufferX    int32
	FramebufferY    int32
	Width           int32
	Height          int32
	GenerateMipmaps bool
}

type CommandBeginRenderPass struct {
	FramebufferID     uint32
	ViewportX         int32
	ViewportY         int32
	ViewportWidth     int32
	ViewportHeight    int32
	Colors            [4]CommandColorAttachment
	DepthLoadOp       CommandLoadOperation
	DepthStoreOp      CommandStoreOperation
	DepthClearValue   float32
	StencilLoadOp     CommandLoadOperation
	StencilStoreOp    CommandStoreOperation
	StencilClearValue int32
}

type CommandColorAttachment struct {
	LoadOp     CommandLoadOperation
	StoreOp    CommandStoreOperation
	ClearValue [4]float32
}

type CommandLoadOperation uint8

func CommandLoadOperationFromRender(value render.LoadOperation) CommandLoadOperation {
	return CommandLoadOperation(value)
}

func CommandLoadOperationToRender(value CommandLoadOperation) render.LoadOperation {
	return render.LoadOperation(value)
}

type CommandStoreOperation uint8

func CommandStoreOperationFromRender(value render.StoreOperation) CommandStoreOperation {
	return CommandStoreOperation(value)
}

func CommandStoreOperationToRender(value CommandStoreOperation) render.StoreOperation {
	return render.StoreOperation(value)
}

type CommandEndRenderPass struct {
}

type CommandBindPipeline struct {
	ProgramID        uint32 // not dynamic
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
	BlendEnabled     bool // not dynamic
	BlendEquation    CommandBlendEquation
	BlendFunc        CommandBlendFunc
	BlendColor       CommandBlendColor
	VertexArray      CommandBindVertexArray
}

type CommandTopology struct {
	Topology uint32
}

type CommandCullTest struct {
	Enabled bool
	Face    uint32
}

type CommandFrontFace struct {
	Orientation uint32
}

type CommandDepthTest struct {
	Enabled bool
}

type CommandDepthWrite struct {
	Enabled bool
}

type CommandDepthComparison struct {
	Mode uint32
}

type CommandStencilTest struct {
	Enabled bool
}

type CommandStencilOperation struct {
	Face        uint32
	StencilFail uint32
	DepthFail   uint32
	Pass        uint32
}

type CommandStencilFunc struct {
	Face uint32
	Func uint32
	Ref  int32
	Mask uint32
}

type CommandStencilMask struct {
	Face uint32
	Mask uint32
}

type CommandColorWrite struct {
	Mask [4]bool
}

type CommandBlendColor struct {
	Color [4]float32
}

type CommandBlendEquation struct {
	ModeRGB   uint32
	ModeAlpha uint32
}

type CommandBlendFunc struct {
	SourceFactorRGB        uint32
	DestinationFactorRGB   uint32
	SourceFactorAlpha      uint32
	DestinationFactorAlpha uint32
}

type CommandBindVertexArray struct {
	VertexArrayID uint32
	IndexFormat   uint32
}

type CommandUniformBufferUnit struct {
	Index    uint32
	BufferID uint32
	Offset   uint32
	Size     uint32
}

type CommandTextureUnit struct {
	Index     uint32
	TextureID uint32
}

type CommandSamplerUnit struct {
	Index     uint32
	SamplerID uint32
}

type CommandDraw struct {
	VertexOffset  int32
	VertexCount   int32
	InstanceCount int32
}

type CommandDrawIndexed struct {
	IndexOffset   int32
	IndexCount    int32
	InstanceCount int32
}
