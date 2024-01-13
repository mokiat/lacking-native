package internal

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewRenderer() *Renderer {
	result := &Renderer{
		framebuffer:   DefaultFramebuffer,
		isDirty:       true,
		isInvalidated: true,
		desiredState: &State{
			CullTest:                    false,
			CullFace:                    gl.BACK,
			FrontFace:                   gl.CCW,
			DepthTest:                   false,
			DepthMask:                   true,
			DepthComparison:             gl.LESS,
			StencilTest:                 false,
			StencilOpStencilFailFront:   gl.KEEP,
			StencilOpDepthFailFront:     gl.KEEP,
			StencilOpPassFront:          gl.KEEP,
			StencilOpStencilFailBack:    gl.KEEP,
			StencilOpDepthFailBack:      gl.KEEP,
			StencilOpPassBack:           gl.KEEP,
			StencilComparisonFuncFront:  gl.ALWAYS,
			StencilComparisonRefFront:   0x00,
			StencilComparisonMaskFront:  0xFF,
			StencilComparisonFuncBack:   gl.ALWAYS,
			StencilComparisonRefBack:    0x00,
			StencilComparisonMaskBack:   0xFF,
			StencilMaskFront:            0xFF,
			StencilMaskBack:             0xFF,
			ColorMask:                   render.ColorMaskTrue,
			Blending:                    false,
			BlendModeRGB:                gl.FUNC_ADD,
			BlendModeAlpha:              gl.FUNC_ADD,
			BlendSourceFactorRGB:        gl.ONE,
			BlendDestinationFactorRGB:   gl.ZERO,
			BlendSourceFactorAlpha:      gl.ONE,
			BlendDestinationFactorAlpha: gl.ZERO,
		},
		actualState: &State{},
	}
	result.Invalidate()
	return result
}

type Renderer struct {
	framebuffer           *Framebuffer
	invalidateAttachments []uint32
	program               uint32
	topology              uint32
	indexType             uint32

	isDirty       bool
	isInvalidated bool
	desiredState  *State
	actualState   *State
}

func (r *Renderer) BeginRenderPass(info render.RenderPassInfo) {
	r.validateState()

	gl.Enable(gl.PRIMITIVE_RESTART_FIXED_INDEX)
	gl.Enable(gl.CLIP_DISTANCE0)
	gl.Enable(gl.CLIP_DISTANCE1)
	gl.Enable(gl.CLIP_DISTANCE2)
	gl.Enable(gl.CLIP_DISTANCE3)

	r.framebuffer = info.Framebuffer.(*Framebuffer)
	isDefaultFramebuffer := r.framebuffer.id == 0

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.framebuffer.id)
	gl.Viewport(
		int32(info.Viewport.X),
		int32(info.Viewport.Y),
		int32(info.Viewport.Width),
		int32(info.Viewport.Height),
	)

	oldColorMask := r.actualState.ColorMask

	var colorMaskChanged bool
	for i, attachment := range info.Colors {
		if r.framebuffer.activeDrawBuffers[i] && (attachment.LoadOp == render.LoadOperationClear) {
			if !colorMaskChanged {
				r.executeCommandColorWrite(CommandColorWrite{
					Mask: render.ColorMaskTrue,
				})
				r.validateColorMask(false)
				colorMaskChanged = true
			}
			gl.ClearBufferfv(gl.COLOR, int32(i), &attachment.ClearValue[0])
		}
	}
	if colorMaskChanged {
		r.executeCommandColorWrite(CommandColorWrite{
			Mask: oldColorMask,
		})
	}

	oldDepthMask := r.actualState.DepthMask
	oldStencilMaskFront := r.actualState.StencilMaskFront
	oldStencilMaskBack := r.actualState.StencilMaskBack

	clearDepth := info.DepthLoadOp == render.LoadOperationClear
	clearStencil := info.StencilLoadOp == render.LoadOperationClear

	if clearDepth && clearStencil {
		r.executeCommandDepthWrite(CommandDepthWrite{
			Enabled: true,
		})
		r.validateDepthMask(false)
		r.executeCommandStencilMask(CommandStencilMask{
			Face: gl.FRONT_AND_BACK,
			Mask: 0xFF,
		})
		r.validateStencilMask(false)
		depthValue := info.DepthClearValue
		stencilValue := int32(info.StencilClearValue)
		gl.ClearBufferfi(gl.DEPTH_STENCIL, 0, depthValue, stencilValue)
	} else {
		if clearDepth {
			r.executeCommandDepthWrite(CommandDepthWrite{
				Enabled: true,
			})
			r.validateDepthMask(false)
			depthValue := info.DepthClearValue
			gl.ClearBufferfv(gl.DEPTH, 0, &depthValue)
		}
		if clearStencil {
			r.executeCommandStencilMask(CommandStencilMask{
				Face: gl.FRONT_AND_BACK,
				Mask: 0xFF,
			})
			r.validateStencilMask(false)
			stencilValue := int32(info.StencilClearValue)
			gl.ClearBufferiv(gl.STENCIL, 0, &stencilValue)
		}
	}

	r.invalidateAttachments = r.invalidateAttachments[:0]

	for i, attachment := range info.Colors {
		if r.framebuffer.activeDrawBuffers[i] && (attachment.StoreOp == render.StoreOperationDontCare) {
			if isDefaultFramebuffer {
				if i == 0 {
					r.invalidateAttachments = append(r.invalidateAttachments, gl.COLOR)
				}
			} else {
				r.invalidateAttachments = append(r.invalidateAttachments, gl.COLOR_ATTACHMENT0+uint32(i))
			}
		}
	}

	invalidateDepth := info.DepthStoreOp == render.StoreOperationDontCare
	invalidateStencil := info.StencilStoreOp == render.StoreOperationDontCare

	if invalidateDepth && invalidateStencil && !isDefaultFramebuffer {
		r.invalidateAttachments = append(r.invalidateAttachments, gl.DEPTH_STENCIL_ATTACHMENT)
	} else {
		if invalidateDepth {
			if isDefaultFramebuffer {
				r.invalidateAttachments = append(r.invalidateAttachments, gl.DEPTH)
			} else {
				r.invalidateAttachments = append(r.invalidateAttachments, gl.DEPTH_ATTACHMENT)
			}
		}
		if invalidateStencil {
			if isDefaultFramebuffer {
				r.invalidateAttachments = append(r.invalidateAttachments, gl.STENCIL)
			} else {
				r.invalidateAttachments = append(r.invalidateAttachments, gl.STENCIL_ATTACHMENT)
			}
		}
	}

	r.executeCommandDepthWrite(CommandDepthWrite{
		Enabled: oldDepthMask,
	})
	r.executeCommandStencilMask(CommandStencilMask{
		Face: gl.FRONT,
		Mask: uint32(oldStencilMaskFront),
	})
	r.executeCommandStencilMask(CommandStencilMask{
		Face: gl.BACK,
		Mask: uint32(oldStencilMaskBack),
	})
}

func (r *Renderer) EndRenderPass() {
	gl.Disable(gl.CLIP_DISTANCE0)
	gl.Disable(gl.CLIP_DISTANCE1)
	gl.Disable(gl.CLIP_DISTANCE2)
	gl.Disable(gl.CLIP_DISTANCE3)
	gl.Disable(gl.PRIMITIVE_RESTART_FIXED_INDEX)
	r.framebuffer = DefaultFramebuffer
}

func (r *Renderer) Invalidate() {
	r.program = 0
	r.isDirty = true
	r.isInvalidated = true
}

func (r *Renderer) CopyContentToTexture(info render.CopyContentToTextureInfo) {
	intTexture := info.Texture.(*Texture)
	gl.BindTexture(intTexture.kind, intTexture.id)
	gl.CopyTexSubImage2D(
		intTexture.kind,
		int32(info.TextureLevel),
		int32(info.TextureX),
		int32(info.TextureY),
		int32(info.FramebufferX),
		int32(info.FramebufferY),
		int32(info.Width),
		int32(info.Height),
	)
	if info.GenerateMipmaps {
		gl.GenerateMipmap(intTexture.kind)
	}
}

func (r *Renderer) SubmitQueue(queue *CommandQueue) {
	for MoreCommands(queue) {
		header := PopCommand[CommandHeader](queue)
		switch header.Kind {
		case CommandKindBindPipeline:
			command := PopCommand[CommandBindPipeline](queue)
			r.executeCommandBindPipeline(command)
		case CommandKindUniform1f:
			command := PopCommand[CommandUniform1f](queue)
			r.executeCommandUniform1f(command)
		case CommandKindUniform3f:
			command := PopCommand[CommandUniform3f](queue)
			r.executeCommandUniform3f(command)
		case CommandKindUniform4f:
			command := PopCommand[CommandUniform4f](queue)
			r.executeCommandUniform4f(command)
		case CommandKindUniformBufferUnit:
			command := PopCommand[CommandUniformBufferUnit](queue)
			r.executeCommandUniformBufferUnit(command)
		case CommandKindUniformBufferUnitRange:
			command := PopCommand[CommandUniformBufferUnitRange](queue)
			r.executeCommandUniformBufferUnitRange(command)
		case CommandKindTextureUnit:
			command := PopCommand[CommandTextureUnit](queue)
			r.executeCommandTextureUnit(command)
		case CommandKindDraw:
			command := PopCommand[CommandDraw](queue)
			r.executeCommandDraw(command)
		case CommandKindDrawIndexed:
			command := PopCommand[CommandDrawIndexed](queue)
			r.executeCommandDrawIndexed(command)
		case CommandKindCopyContentToBuffer:
			command := PopCommand[CommandCopyContentToBuffer](queue)
			r.executeCommandCopyContentToBuffer(command)
		case CommandKindUpdateBufferData:
			command := PopCommand[CommandUpdateBufferData](queue)
			data := PopData(queue, command.Count)
			r.executeCommandUpdateBufferData(command, data)
		default:
			panic(fmt.Errorf("unknown command kind: %v", header.Kind))
		}
	}
	queue.Reset()
}

func (r *Renderer) executeCommandBindPipeline(command CommandBindPipeline) {
	if r.program != command.ProgramID {
		r.program = command.ProgramID
		gl.UseProgram(command.ProgramID)
	}
	r.executeCommandTopology(command.Topology)
	r.executeCommandCullTest(command.CullTest)
	r.executeCommandFrontFace(command.FrontFace)
	r.executeCommandDepthTest(command.DepthTest)
	r.executeCommandDepthWrite(command.DepthWrite)
	if command.DepthTest.Enabled {
		r.executeCommandDepthComparison(command.DepthComparison)
	}
	r.executeCommandStencilTest(command.StencilTest)
	if command.StencilTest.Enabled {
		r.executeCommandStencilOperation(command.StencilOpFront)
		r.executeCommandStencilOperation(command.StencilOpBack)
		r.executeCommandStencilFunc(command.StencilFuncFront)
		r.executeCommandStencilFunc(command.StencilFuncBack)
		r.executeCommandStencilMask(command.StencilMaskFront)
		r.executeCommandStencilMask(command.StencilMaskBack)
	}
	r.executeCommandColorWrite(command.ColorWrite)
	r.desiredState.Blending = command.BlendEnabled
	r.isDirty = true
	if command.BlendEnabled {
		r.executeCommandBlendColor(command.BlendColor)
		r.executeCommandBlendEquation(command.BlendEquation)
		r.executeCommandBlendFunc(command.BlendFunc)
	}
	r.executeCommandBindVertexArray(command.VertexArray)
}

func (r *Renderer) executeCommandTopology(command CommandTopology) {
	r.topology = command.Topology
}

func (r *Renderer) executeCommandCullTest(command CommandCullTest) {
	r.desiredState.CullTest = command.Enabled
	if command.Enabled {
		r.desiredState.CullFace = command.Face
	}
	r.isDirty = true
}

func (r *Renderer) executeCommandFrontFace(command CommandFrontFace) {
	r.desiredState.FrontFace = command.Orientation
	r.isDirty = true
}

func (r *Renderer) executeCommandDepthTest(command CommandDepthTest) {
	r.desiredState.DepthTest = command.Enabled
	r.isDirty = true
}

func (r *Renderer) executeCommandDepthWrite(command CommandDepthWrite) {
	r.desiredState.DepthMask = command.Enabled
	r.isDirty = true
}

func (r *Renderer) executeCommandDepthComparison(command CommandDepthComparison) {
	r.desiredState.DepthComparison = command.Mode
	r.isDirty = true
}

func (r *Renderer) executeCommandStencilTest(command CommandStencilTest) {
	r.desiredState.StencilTest = command.Enabled
	r.isDirty = true
}

func (r *Renderer) executeCommandStencilOperation(command CommandStencilOperation) {
	if command.Face == gl.FRONT || command.Face == gl.FRONT_AND_BACK {
		r.desiredState.StencilOpStencilFailFront = command.StencilFail
		r.desiredState.StencilOpDepthFailFront = command.DepthFail
		r.desiredState.StencilOpPassFront = command.Pass
	}
	if command.Face == gl.BACK || command.Face == gl.FRONT_AND_BACK {
		r.desiredState.StencilOpStencilFailBack = command.StencilFail
		r.desiredState.StencilOpDepthFailBack = command.DepthFail
		r.desiredState.StencilOpPassBack = command.Pass
	}
	r.isDirty = true
}

func (r *Renderer) executeCommandStencilFunc(command CommandStencilFunc) {
	if command.Face == gl.FRONT || command.Face == gl.FRONT_AND_BACK {
		r.desiredState.StencilComparisonFuncFront = command.Func
		r.desiredState.StencilComparisonRefFront = command.Ref
		r.desiredState.StencilComparisonMaskFront = command.Mask
	}
	if command.Face == gl.BACK || command.Face == gl.FRONT_AND_BACK {
		r.desiredState.StencilComparisonFuncBack = command.Func
		r.desiredState.StencilComparisonRefBack = command.Ref
		r.desiredState.StencilComparisonMaskBack = command.Mask
	}
	r.isDirty = true
}

func (r *Renderer) executeCommandStencilMask(command CommandStencilMask) {
	if command.Face == gl.FRONT || command.Face == gl.FRONT_AND_BACK {
		r.desiredState.StencilMaskFront = command.Mask
	}
	if command.Face == gl.BACK || command.Face == gl.FRONT_AND_BACK {
		r.desiredState.StencilMaskBack = command.Mask
	}
	r.isDirty = true
}

func (r *Renderer) executeCommandColorWrite(command CommandColorWrite) {
	r.desiredState.ColorMask = command.Mask
	r.isDirty = true
}

func (r *Renderer) executeCommandBlendColor(command CommandBlendColor) {
	r.desiredState.BlendColor = command.Color
	r.isDirty = true
}

func (r *Renderer) executeCommandBlendEquation(command CommandBlendEquation) {
	r.desiredState.BlendModeRGB = command.ModeRGB
	r.desiredState.BlendModeAlpha = command.ModeAlpha
	r.isDirty = true
}

func (r *Renderer) executeCommandBlendFunc(command CommandBlendFunc) {
	r.desiredState.BlendSourceFactorRGB = command.SourceFactorRGB
	r.desiredState.BlendDestinationFactorRGB = command.DestinationFactorRGB
	r.desiredState.BlendSourceFactorAlpha = command.SourceFactorAlpha
	r.desiredState.BlendDestinationFactorAlpha = command.DestinationFactorAlpha
	r.isDirty = true
}

func (r *Renderer) executeCommandBindVertexArray(command CommandBindVertexArray) {
	gl.BindVertexArray(command.VertexArrayID)
	r.indexType = command.IndexFormat
}

func (r *Renderer) executeCommandUniform1f(command CommandUniform1f) {
	gl.Uniform1f(
		command.Location,
		command.Value,
	)
}

func (r *Renderer) executeCommandUniform3f(command CommandUniform3f) {
	gl.Uniform3f(
		command.Location,
		command.Values[0],
		command.Values[1],
		command.Values[2],
	)
}

func (r *Renderer) executeCommandUniform4f(command CommandUniform4f) {
	gl.Uniform4f(
		command.Location,
		command.Values[0],
		command.Values[1],
		command.Values[2],
		command.Values[3],
	)
}

func (r *Renderer) executeCommandUniformBufferUnit(command CommandUniformBufferUnit) {
	gl.BindBufferBase(
		gl.UNIFORM_BUFFER,
		command.Index,
		command.BufferID,
	)
}

func (r *Renderer) executeCommandUniformBufferUnitRange(command CommandUniformBufferUnitRange) {
	gl.BindBufferRange(
		gl.UNIFORM_BUFFER,
		command.Index,
		command.BufferID,
		int(command.Offset),
		int(command.Size),
	)
}

func (r *Renderer) executeCommandTextureUnit(command CommandTextureUnit) {
	texture := textures.Get(command.TextureID)
	gl.ActiveTexture(uint32(gl.TEXTURE0) + uint32(command.Index))
	gl.BindTexture(texture.kind, texture.id)
}

func (r *Renderer) executeCommandDraw(command CommandDraw) {
	r.validateState()
	gl.DrawArraysInstanced(
		r.topology,
		command.VertexOffset,
		command.VertexCount,
		command.InstanceCount,
	)
}

func (r *Renderer) executeCommandDrawIndexed(command CommandDrawIndexed) {
	r.validateState()
	gl.DrawElementsInstanced(
		r.topology,
		command.IndexCount,
		r.indexType,
		gl.PtrOffset(int(command.IndexOffset)),
		command.InstanceCount,
	)
}

func (r *Renderer) executeCommandCopyContentToBuffer(command CommandCopyContentToBuffer) {
	gl.BindBuffer(
		gl.PIXEL_PACK_BUFFER,
		command.BufferID,
	)
	gl.ReadPixels(
		command.X,
		command.Y,
		command.Width,
		command.Height,
		command.Format,
		command.XType,
		gl.PtrOffset(0),
	)
	gl.BindBuffer(
		gl.PIXEL_PACK_BUFFER,
		0,
	)
}

func (r *Renderer) executeCommandUpdateBufferData(command CommandUpdateBufferData, data []byte) {
	buffer := buffers.Get(command.BufferID)
	gl.BindBuffer(buffer.kind, buffer.id)
	gl.BufferSubData(buffer.kind, int(command.Offset), len(data), gl.Ptr(&data[0]))
	gl.BindBuffer(buffer.kind, 0)
}

func (r *Renderer) validateState() {
	if r.isDirty || r.isInvalidated {
		forcedUpdate := r.isInvalidated
		r.validateCullTest(forcedUpdate)
		r.validateCullFace(forcedUpdate)
		r.validateFrontFace(forcedUpdate)
		r.validateDepthTest(forcedUpdate)
		r.validateDepthMask(forcedUpdate)
		r.validateDepthComparison(forcedUpdate)
		r.validateStencilTest(forcedUpdate)
		r.validateStencilOperation(forcedUpdate)
		r.validateStencilComparison(forcedUpdate)
		r.validateStencilMask(forcedUpdate)
		r.validateColorMask(forcedUpdate)
		r.validateBlending(forcedUpdate)
		r.validateBlendEquation(forcedUpdate)
		r.validateBlendFunc(forcedUpdate)
		r.validateBlendColor(forcedUpdate)
	}
	r.isDirty = false
	r.isInvalidated = false
}

func (r *Renderer) validateCullTest(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.CullTest != r.desiredState.CullTest)

	if needsUpdate {
		r.actualState.CullTest = r.desiredState.CullTest
		if r.actualState.CullTest {
			gl.Enable(gl.CULL_FACE)
		} else {
			gl.Disable(gl.CULL_FACE)
		}
	}
}

func (r *Renderer) validateCullFace(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.CullFace != r.desiredState.CullFace)

	if needsUpdate {
		r.actualState.CullFace = r.desiredState.CullFace
		gl.CullFace(r.actualState.CullFace)
	}
}

func (r *Renderer) validateFrontFace(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.FrontFace != r.desiredState.FrontFace)

	if needsUpdate {
		r.actualState.FrontFace = r.desiredState.FrontFace
		gl.FrontFace(r.actualState.FrontFace)
	}
}

func (r *Renderer) validateDepthTest(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.DepthTest != r.desiredState.DepthTest)

	if needsUpdate {
		r.actualState.DepthTest = r.desiredState.DepthTest
		if r.actualState.DepthTest {
			gl.Enable(gl.DEPTH_TEST)
		} else {
			gl.Disable(gl.DEPTH_TEST)
		}
	}
}

func (r *Renderer) validateDepthMask(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.DepthMask != r.desiredState.DepthMask)

	if needsUpdate {
		r.actualState.DepthMask = r.desiredState.DepthMask
		gl.DepthMask(r.actualState.DepthMask)
	}
}

func (r *Renderer) validateDepthComparison(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.DepthComparison != r.desiredState.DepthComparison)

	if needsUpdate {
		r.actualState.DepthComparison = r.desiredState.DepthComparison
		gl.DepthFunc(r.actualState.DepthComparison)
	}
}

func (r *Renderer) validateStencilTest(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.StencilTest != r.desiredState.StencilTest)

	if needsUpdate {
		r.actualState.StencilTest = r.desiredState.StencilTest
		if r.actualState.StencilTest {
			gl.Enable(gl.STENCIL_TEST)
		} else {
			gl.Disable(gl.STENCIL_TEST)
		}
	}
}

func (r *Renderer) validateStencilOperation(forcedUpdate bool) {
	frontNeedsUpdate := forcedUpdate ||
		(r.actualState.StencilOpStencilFailFront != r.desiredState.StencilOpStencilFailFront) ||
		(r.actualState.StencilOpDepthFailFront != r.desiredState.StencilOpDepthFailFront) ||
		(r.actualState.StencilOpPassFront != r.desiredState.StencilOpPassFront)

	backNeedsUpdate := forcedUpdate ||
		(r.actualState.StencilOpStencilFailBack != r.desiredState.StencilOpStencilFailBack) ||
		(r.actualState.StencilOpDepthFailBack != r.desiredState.StencilOpDepthFailBack) ||
		(r.actualState.StencilOpPassBack != r.desiredState.StencilOpPassBack)

	if frontNeedsUpdate {
		r.actualState.StencilOpStencilFailFront = r.desiredState.StencilOpStencilFailFront
		r.actualState.StencilOpDepthFailFront = r.desiredState.StencilOpDepthFailFront
		r.actualState.StencilOpPassFront = r.desiredState.StencilOpPassFront
	}

	if backNeedsUpdate {
		r.actualState.StencilOpStencilFailBack = r.desiredState.StencilOpStencilFailBack
		r.actualState.StencilOpDepthFailBack = r.desiredState.StencilOpDepthFailBack
		r.actualState.StencilOpPassBack = r.desiredState.StencilOpPassBack
	}

	frontEqualsBack := (r.desiredState.StencilOpStencilFailFront == r.desiredState.StencilOpStencilFailBack) &&
		(r.desiredState.StencilOpDepthFailFront == r.desiredState.StencilOpDepthFailBack) &&
		(r.desiredState.StencilOpPassFront == r.desiredState.StencilOpPassBack)

	if frontNeedsUpdate && backNeedsUpdate && frontEqualsBack {
		gl.StencilOpSeparate(
			gl.FRONT_AND_BACK,
			r.actualState.StencilOpStencilFailFront,
			r.actualState.StencilOpDepthFailFront,
			r.actualState.StencilOpPassFront,
		)
	} else {
		if frontNeedsUpdate {
			gl.StencilOpSeparate(
				gl.FRONT,
				r.actualState.StencilOpStencilFailFront,
				r.actualState.StencilOpDepthFailFront,
				r.actualState.StencilOpPassFront,
			)
		}
		if backNeedsUpdate {
			gl.StencilOpSeparate(
				gl.BACK,
				r.actualState.StencilOpStencilFailBack,
				r.actualState.StencilOpDepthFailBack,
				r.actualState.StencilOpPassBack,
			)
		}
	}
}

func (r *Renderer) validateStencilComparison(forcedUpdate bool) {
	frontNeedsUpdate := forcedUpdate ||
		(r.actualState.StencilComparisonFuncFront != r.desiredState.StencilComparisonFuncFront) ||
		(r.actualState.StencilComparisonRefFront != r.desiredState.StencilComparisonRefFront) ||
		(r.actualState.StencilComparisonMaskFront != r.desiredState.StencilComparisonMaskFront)

	backNeedsUpdate := forcedUpdate ||
		(r.actualState.StencilComparisonFuncBack != r.desiredState.StencilComparisonFuncBack) ||
		(r.actualState.StencilComparisonRefBack != r.desiredState.StencilComparisonRefBack) ||
		(r.actualState.StencilComparisonMaskBack != r.desiredState.StencilComparisonMaskBack)

	if frontNeedsUpdate {
		r.actualState.StencilComparisonFuncFront = r.desiredState.StencilComparisonFuncFront
		r.actualState.StencilComparisonRefFront = r.desiredState.StencilComparisonRefFront
		r.actualState.StencilComparisonMaskFront = r.desiredState.StencilComparisonMaskFront
	}

	if backNeedsUpdate {
		r.actualState.StencilComparisonFuncBack = r.desiredState.StencilComparisonFuncBack
		r.actualState.StencilComparisonRefBack = r.desiredState.StencilComparisonRefBack
		r.actualState.StencilComparisonMaskBack = r.desiredState.StencilComparisonMaskBack
	}

	frontEqualsBack := (r.desiredState.StencilComparisonFuncFront == r.desiredState.StencilComparisonFuncBack) &&
		(r.desiredState.StencilComparisonRefFront == r.desiredState.StencilComparisonRefBack) &&
		(r.desiredState.StencilComparisonMaskFront == r.desiredState.StencilComparisonMaskBack)

	if frontNeedsUpdate && backNeedsUpdate && frontEqualsBack {
		gl.StencilFuncSeparate(
			gl.FRONT_AND_BACK,
			r.actualState.StencilComparisonFuncFront,
			r.actualState.StencilComparisonRefFront,
			r.actualState.StencilComparisonMaskFront,
		)
	} else {
		if frontNeedsUpdate {
			gl.StencilFuncSeparate(
				gl.FRONT,
				r.actualState.StencilComparisonFuncFront,
				r.actualState.StencilComparisonRefFront,
				r.actualState.StencilComparisonMaskFront,
			)
		}
		if backNeedsUpdate {
			gl.StencilFuncSeparate(
				gl.BACK,
				r.actualState.StencilComparisonFuncBack,
				r.actualState.StencilComparisonRefBack,
				r.actualState.StencilComparisonMaskBack,
			)
		}
	}
}

func (r *Renderer) validateStencilMask(forcedUpdate bool) {
	frontNeedsUpdate := forcedUpdate ||
		(r.actualState.StencilMaskFront != r.desiredState.StencilMaskFront)

	backNeedsUpdate := forcedUpdate ||
		(r.actualState.StencilMaskBack != r.desiredState.StencilMaskBack)

	if frontNeedsUpdate {
		r.actualState.StencilMaskFront = r.desiredState.StencilMaskFront
	}
	if backNeedsUpdate {
		r.actualState.StencilMaskBack = r.desiredState.StencilMaskBack
	}

	frontEqualsBack := (r.desiredState.StencilMaskFront == r.desiredState.StencilMaskBack)

	if frontNeedsUpdate && backNeedsUpdate && frontEqualsBack {
		gl.StencilMaskSeparate(
			gl.FRONT_AND_BACK,
			r.actualState.StencilMaskFront,
		)
	} else {
		if frontNeedsUpdate {
			gl.StencilMaskSeparate(
				gl.FRONT,
				r.actualState.StencilMaskFront,
			)
		}
		if backNeedsUpdate {
			gl.StencilMaskSeparate(
				gl.BACK,
				r.actualState.StencilMaskBack,
			)
		}
	}
}

func (r *Renderer) validateColorMask(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.ColorMask != r.desiredState.ColorMask)

	if needsUpdate {
		r.actualState.ColorMask = r.desiredState.ColorMask
		gl.ColorMask(
			r.actualState.ColorMask[0],
			r.actualState.ColorMask[1],
			r.actualState.ColorMask[2],
			r.actualState.ColorMask[3],
		)
	}
}

func (r *Renderer) validateBlending(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.Blending != r.desiredState.Blending)

	if needsUpdate {
		r.actualState.Blending = r.desiredState.Blending
		if r.actualState.Blending {
			gl.Enable(gl.BLEND)
		} else {
			gl.Disable(gl.BLEND)
		}
	}
}

func (r *Renderer) validateBlendColor(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.BlendColor != r.desiredState.BlendColor)

	if needsUpdate {
		r.actualState.BlendColor = r.desiredState.BlendColor
		gl.BlendColor(
			r.actualState.BlendColor[0],
			r.actualState.BlendColor[1],
			r.actualState.BlendColor[2],
			r.actualState.BlendColor[3],
		)
	}
}

func (r *Renderer) validateBlendEquation(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.BlendModeRGB != r.desiredState.BlendModeRGB) ||
		(r.actualState.BlendModeAlpha != r.desiredState.BlendModeAlpha)

	if needsUpdate {
		r.actualState.BlendModeRGB = r.desiredState.BlendModeRGB
		r.actualState.BlendModeAlpha = r.desiredState.BlendModeAlpha
		gl.BlendEquationSeparate(
			r.actualState.BlendModeRGB,
			r.actualState.BlendModeAlpha,
		)
	}
}

func (r *Renderer) validateBlendFunc(forcedUpdate bool) {
	needsUpdate := forcedUpdate ||
		(r.actualState.BlendSourceFactorRGB != r.desiredState.BlendSourceFactorRGB) ||
		(r.actualState.BlendDestinationFactorRGB != r.desiredState.BlendDestinationFactorRGB) ||
		(r.actualState.BlendSourceFactorAlpha != r.desiredState.BlendSourceFactorAlpha) ||
		(r.actualState.BlendDestinationFactorAlpha != r.desiredState.BlendDestinationFactorAlpha)

	if needsUpdate {
		r.actualState.BlendSourceFactorRGB = r.desiredState.BlendSourceFactorRGB
		r.actualState.BlendDestinationFactorRGB = r.desiredState.BlendDestinationFactorRGB
		r.actualState.BlendSourceFactorAlpha = r.desiredState.BlendSourceFactorAlpha
		r.actualState.BlendDestinationFactorAlpha = r.desiredState.BlendDestinationFactorAlpha
		gl.BlendFuncSeparate(
			r.actualState.BlendSourceFactorRGB,
			r.actualState.BlendDestinationFactorRGB,
			r.actualState.BlendSourceFactorAlpha,
			r.actualState.BlendDestinationFactorAlpha,
		)
	}
}
