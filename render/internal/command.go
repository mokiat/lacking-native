package internal

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewCommandQueue() *CommandQueue {
	return &CommandQueue{
		data: make([]byte, 1024*1024),
	}
}

type CommandQueue struct {
	data        []byte
	writeOffset uintptr
	readOffset  uintptr
}

func (q *CommandQueue) Reset() {
	q.readOffset = 0
	q.writeOffset = 0
}

func (q *CommandQueue) BindPipeline(pipeline render.Pipeline) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindBindPipeline,
	})
	intPipeline := pipeline.(*Pipeline)
	PushCommand(q, CommandBindPipeline{
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

func (q *CommandQueue) Uniform1f(location render.UniformLocation, value float32) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindUniform1f,
	})
	PushCommand(q, CommandUniform1f{
		Location: location.(int32),
		Value:    value,
	})
}

func (q *CommandQueue) Uniform3f(location render.UniformLocation, values [3]float32) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindUniform3f,
	})
	PushCommand(q, CommandUniform3f{
		Location: location.(int32),
		Values:   values,
	})
}

func (q *CommandQueue) Uniform4f(location render.UniformLocation, values [4]float32) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindUniform4f,
	})
	PushCommand(q, CommandUniform4f{
		Location: location.(int32),
		Values:   values,
	})
}

func (q *CommandQueue) UniformBufferUnit(index int, buffer render.Buffer) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindUniformBufferUnit,
	})
	PushCommand(q, CommandUniformBufferUnit{
		Index:    uint32(index),
		BufferID: buffer.(*Buffer).id,
	})
}

func (q *CommandQueue) UniformBufferUnitRange(index int, buffer render.Buffer, offset, size int) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindUniformBufferUnitRange,
	})
	PushCommand(q, CommandUniformBufferUnitRange{
		Index:    uint32(index),
		BufferID: buffer.(*Buffer).id,
		Offset:   uint32(offset),
		Size:     uint32(size),
	})
}

func (q *CommandQueue) TextureUnit(index int, texture render.Texture) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindTextureUnit,
	})
	PushCommand(q, CommandTextureUnit{
		Index:     uint32(index),
		TextureID: texture.(*Texture).id,
	})
}

func (q *CommandQueue) Draw(vertexOffset, vertexCount, instanceCount int) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindDraw,
	})
	PushCommand(q, CommandDraw{
		VertexOffset:  int32(vertexOffset),
		VertexCount:   int32(vertexCount),
		InstanceCount: int32(instanceCount),
	})
}

func (q *CommandQueue) DrawIndexed(indexOffset, indexCount, instanceCount int) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindDrawIndexed,
	})
	PushCommand(q, CommandDrawIndexed{
		IndexOffset:   int32(indexOffset),
		IndexCount:    int32(indexCount),
		InstanceCount: int32(instanceCount),
	})
}

func (q *CommandQueue) CopyContentToBuffer(info render.CopyContentToBufferInfo) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindCopyContentToBuffer,
	})
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
	PushCommand(q, CommandCopyContentToBuffer{
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

func (q *CommandQueue) UpdateBufferData(buffer render.Buffer, info render.BufferUpdateInfo) {
	PushCommand(q, CommandHeader{
		Kind: CommandKindUpdateBufferData,
	})
	PushCommand(q, CommandUpdateBufferData{
		BufferID: buffer.(*Buffer).id,
		Offset:   uint32(info.Offset),
		Count:    uint32(len(info.Data)),
	})
	PushData(q, info.Data)
}

func (q *CommandQueue) Release() {
	q.data = nil
}

func (q *CommandQueue) ensure(size int) {
	requiredSize := int(q.writeOffset) + size
	currentSize := len(q.data)
	if requiredSize > currentSize {
		newSize := currentSize * 2
		for newSize < requiredSize {
			newSize *= 2
		}
		newData := make([]byte, newSize)
		copy(newData, q.data)
		q.data = newData
	}
}

func MoreCommands(queue *CommandQueue) bool {
	return queue.writeOffset > queue.readOffset
}

func PushCommand[T any](queue *CommandQueue, command T) {
	size := unsafe.Sizeof(command)
	queue.ensure(int(size))
	target := (*T)(unsafe.Add(unsafe.Pointer(&queue.data[0]), queue.writeOffset))
	*target = command
	queue.writeOffset += size
}

func PushData(queue *CommandQueue, data []byte) {
	queue.ensure(len(data))
	copy(queue.data[queue.writeOffset:], data)
	queue.writeOffset += uintptr(len(data))
}

func PopCommand[T any](queue *CommandQueue) T {
	target := (*T)(unsafe.Add(unsafe.Pointer(&queue.data[0]), queue.readOffset))
	command := *target
	queue.readOffset += unsafe.Sizeof(command)
	return command
}

func PopData(queue *CommandQueue, count uint32) []byte {
	result := queue.data[queue.readOffset : queue.readOffset+uintptr(count)]
	queue.readOffset += uintptr(count)
	return result
}

type CommandKind uint8

const (
	CommandKindBindPipeline CommandKind = iota
	CommandKindUniform1f
	CommandKindUniform3f
	CommandKindUniform4f
	CommandKindUniformBufferUnit
	CommandKindUniformBufferUnitRange
	CommandKindTextureUnit
	CommandKindDraw
	CommandKindDrawIndexed
	CommandKindCopyContentToBuffer
	CommandKindUpdateBufferData
)

type CommandHeader struct {
	Kind CommandKind
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

type CommandUniform1f struct {
	Location int32
	Value    float32
}

type CommandUniform3f struct {
	Location int32
	Values   [3]float32
}

type CommandUniform4f struct {
	Location int32
	Values   [4]float32
}

type CommandUniformBufferUnit struct {
	Index    uint32
	BufferID uint32
}

type CommandUniformBufferUnitRange struct {
	Index    uint32
	BufferID uint32
	Offset   uint32
	Size     uint32
}

type CommandTextureUnit struct {
	Index     uint32
	TextureID uint32
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

type CommandCopyContentToBuffer struct {
	BufferID     uint32
	X            int32
	Y            int32
	Width        int32
	Height       int32
	Format       uint32
	XType        uint32
	BufferOffset uint32
}

type CommandUpdateBufferData struct {
	BufferID uint32
	Offset   uint32
	Count    uint32
}
