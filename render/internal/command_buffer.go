package internal

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewCommandBuffer(initialCapacity uint) *CommandBuffer {
	return &CommandBuffer{
		data: make([]byte, max(1024, initialCapacity)),
	}
}

type CommandBuffer struct {
	render.CommandBufferMarker

	data        []byte
	writeOffset uintptr
	readOffset  uintptr

	isRenderPassActive bool
}

func (b *CommandBuffer) HasMoreCommands() bool {
	return b.writeOffset > b.readOffset
}

func (b *CommandBuffer) Reset() {
	b.readOffset = 0
	b.writeOffset = 0
}

func (b *CommandBuffer) CopyFramebufferToBuffer(info render.CopyFramebufferToBufferInfo) {
	b.verifyIsRenderPass()
	var format, xtype uint32
	switch info.Format {
	case render.DataFormatRGBA8:
		format = gl.RGBA
		xtype = gl.UNSIGNED_BYTE
	case render.DataFormatRGBA16F:
		format = gl.RGBA
		xtype = gl.HALF_FLOAT
	case render.DataFormatRGBA32F:
		format = gl.RGBA
		xtype = gl.FLOAT
	default:
		panic(fmt.Errorf("unsupported data format %v", info.Format))
	}
	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindCopyFramebufferToBuffer,
	})
	writeCommandChunk(b, CommandCopyFramebufferToBuffer{
		BufferID:     info.Buffer.(*Buffer).id,
		X:            int32(info.X),
		Y:            int32(info.Y),
		Width:        int32(info.Width),
		Height:       int32(info.Height),
		Format:       format,
		XType:        xtype,
		BufferOffset: uint32(info.Offset),
	})
}

func (b *CommandBuffer) CopyFramebufferToTexture(info render.CopyFramebufferToTextureInfo) {
	b.verifyIsRenderPass()
	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindCopyFramebufferToTexture,
	})
	writeCommandChunk(b, CommandCopyFramebufferToTexture{
		TextureID:       info.Texture.(*Texture).id,
		TextureLevel:    int32(info.TextureLevel),
		TextureX:        int32(info.TextureX),
		TextureY:        int32(info.TextureY),
		FramebufferX:    int32(info.FramebufferX),
		FramebufferY:    int32(info.FramebufferY),
		Width:           int32(info.Width),
		Height:          int32(info.Height),
		GenerateMipmaps: info.GenerateMipmaps,
	})
}

func (b *CommandBuffer) BeginRenderPass(info render.RenderPassInfo) {
	b.verifyNotRenderPass()
	b.isRenderPassActive = true

	var colors [4]CommandColorAttachment
	for i := range colors {
		colors[i].LoadOp = CommandLoadOperationFromRender(info.Colors[i].LoadOp)
		colors[i].StoreOp = CommandStoreOperationFromRender(info.Colors[i].StoreOp)
		colors[i].ClearValue = info.Colors[i].ClearValue
	}

	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindBeginRenderPass,
	})
	writeCommandChunk(b, CommandBeginRenderPass{
		FramebufferID:     info.Framebuffer.(*Framebuffer).id,
		ViewportX:         int32(info.Viewport.X),
		ViewportY:         int32(info.Viewport.Y),
		ViewportWidth:     int32(info.Viewport.Width),
		ViewportHeight:    int32(info.Viewport.Height),
		Colors:            colors,
		DepthLoadOp:       CommandLoadOperationFromRender(info.DepthLoadOp),
		DepthStoreOp:      CommandStoreOperationFromRender(info.DepthStoreOp),
		DepthClearValue:   info.DepthClearValue,
		DepthBias:         info.DepthBias,
		DepthSlopeBias:    info.DepthSlopeBias,
		StencilLoadOp:     CommandLoadOperationFromRender(info.StencilLoadOp),
		StencilStoreOp:    CommandStoreOperationFromRender(info.StencilStoreOp),
		StencilClearValue: int32(info.StencilClearValue),
	})
}

func (b *CommandBuffer) SetViewport(x, y, width, height uint32) {
	b.verifyIsRenderPass()
	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindSetViewport,
	})
	writeCommandChunk(b, CommandSetViewport{
		X:      int32(x),
		Y:      int32(y),
		Width:  int32(width),
		Height: int32(height),
	})
}

func (b *CommandBuffer) BindPipeline(pipeline render.Pipeline) {
	b.verifyIsRenderPass()
	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindBindPipeline,
	})
	intPipeline := pipeline.(*Pipeline)
	writeCommandChunk(b, CommandBindPipeline{
		ProgramID:        intPipeline.ProgramID,
		Topology:         intPipeline.Topology,
		CullTest:         intPipeline.CullTest,
		FrontFace:        intPipeline.FrontFace,
		DepthTest:        intPipeline.DepthTest,
		DepthWrite:       intPipeline.DepthWrite,
		DepthComparison:  intPipeline.DepthComparison,
		StencilTest:      intPipeline.StencilTest,
		StencilOpFront:   intPipeline.StencilOpFront,
		StencilOpBack:    intPipeline.StencilOpBack,
		StencilFuncFront: intPipeline.StencilFuncFront,
		StencilFuncBack:  intPipeline.StencilFuncBack,
		StencilMaskFront: intPipeline.StencilMaskFront,
		StencilMaskBack:  intPipeline.StencilMaskBack,
		ColorWrite:       intPipeline.ColorWrite,
		BlendEnabled:     intPipeline.BlendEnabled,
		BlendColor:       intPipeline.BlendColor,
		BlendEquation:    intPipeline.BlendEquation,
		BlendFunc:        intPipeline.BlendFunc,
		VertexArray:      intPipeline.VertexArray,
	})
}

func (b *CommandBuffer) TextureUnit(index uint, texture render.Texture) {
	b.verifyIsRenderPass()
	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindTextureUnit,
	})
	writeCommandChunk(b, CommandTextureUnit{
		Index:     uint32(index),
		TextureID: texture.(*Texture).id,
	})
}

func (b *CommandBuffer) SamplerUnit(index uint, sampler render.Sampler) {
	b.verifyIsRenderPass()
	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindSamplerUnit,
	})
	samplerID := uint32(0) // disable
	if sampler != nil {
		samplerID = sampler.(*Sampler).id
	}
	writeCommandChunk(b, CommandSamplerUnit{
		Index:     uint32(index),
		SamplerID: samplerID,
	})
}

func (b *CommandBuffer) UniformBufferUnit(index uint, buffer render.Buffer, offset, size uint32) {
	b.verifyIsRenderPass()
	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindUniformBufferUnit,
	})
	writeCommandChunk(b, CommandUniformBufferUnit{
		Index:    uint32(index),
		BufferID: buffer.(*Buffer).id,
		Offset:   offset,
		Size:     size,
	})
}

func (b *CommandBuffer) Draw(vertexOffset, vertexCount, instanceCount uint32) {
	b.verifyIsRenderPass()
	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindDraw,
	})
	writeCommandChunk(b, CommandDraw{
		VertexOffset:  int32(vertexOffset),
		VertexCount:   int32(vertexCount),
		InstanceCount: int32(instanceCount),
	})
}

func (b *CommandBuffer) DrawIndexed(indexByteOffset, indexCount, instanceCount uint32) {
	b.verifyIsRenderPass()
	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindDrawIndexed,
	})
	writeCommandChunk(b, CommandDrawIndexed{
		IndexByteOffset: int32(indexByteOffset),
		IndexCount:      int32(indexCount),
		InstanceCount:   int32(instanceCount),
	})
}

func (b *CommandBuffer) EndRenderPass() {
	b.verifyIsRenderPass()
	b.isRenderPassActive = false
	writeCommandChunk(b, CommandHeader{
		Kind: CommandKindEndRenderPass,
	})
	writeCommandChunk(b, CommandEndRenderPass{})
}

func (b *CommandBuffer) verifyIsRenderPass() {
	if !b.isRenderPassActive {
		panic("needs to be called from inside a render pass")
	}
}

func (b *CommandBuffer) verifyNotRenderPass() {
	if b.isRenderPassActive {
		panic("cannot be called from inside a render pass")
	}
}

func (b *CommandBuffer) ensure(size int) {
	requiredSize := int(b.writeOffset) + size
	currentSize := len(b.data)
	if requiredSize > currentSize {
		logger.Warn("Command buffer capacity reached! Will grow to accomodate %d bytes.", requiredSize)
		newSize := currentSize * 2
		for newSize < requiredSize {
			newSize *= 2
		}
		newData := make([]byte, newSize)
		copy(newData, b.data)
		b.data = newData
	}
}

func writeCommandChunk[T any](buffer *CommandBuffer, command T) {
	size := unsafe.Sizeof(command)
	buffer.ensure(int(size))
	target := (*T)(unsafe.Add(unsafe.Pointer(&buffer.data[0]), buffer.writeOffset))
	*target = command
	buffer.writeOffset += size
}

func readCommandChunk[T any](buffer *CommandBuffer) T {
	target := (*T)(unsafe.Add(unsafe.Pointer(&buffer.data[0]), buffer.readOffset))
	command := *target
	buffer.readOffset += unsafe.Sizeof(command)
	return command
}
