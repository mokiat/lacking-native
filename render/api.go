package render

import (
	"github.com/mokiat/lacking-native/render/internal"
	"github.com/mokiat/lacking/render"
)

func NewAPI() render.API {
	return &API{
		renderer: internal.NewRenderer(),
	}
}

type API struct {
	renderer *internal.Renderer
}

func (a *API) Capabilities() render.Capabilities {
	return render.Capabilities{
		Quality: render.QualityHigh,
	}
}

func (a *API) DefaultFramebuffer() render.Framebuffer {
	return internal.DefaultFramebuffer
}

func (a *API) CreateFramebuffer(info render.FramebufferInfo) render.Framebuffer {
	return internal.NewFramebuffer(info)
}

func (a *API) CreateProgram(info render.ProgramInfo) render.Program {
	return internal.NewProgram(internal.ProgramInfo{
		Label:           info.Label,
		VertexCode:      info.SourceCode.(ProgramCode).VertexCode,
		FragmentCode:    info.SourceCode.(ProgramCode).FragmentCode,
		TextureBindings: info.TextureBindings,
		UniformBindings: info.UniformBindings,
	})
}

func (a *API) CreateColorTexture2D(info render.ColorTexture2DInfo) render.Texture {
	return internal.NewColorTexture2D(info)
}

func (a *API) CreateColorTextureCube(info render.ColorTextureCubeInfo) render.Texture {
	return internal.NewColorTextureCube(info)
}

func (a *API) CreateDepthTexture2D(info render.DepthTexture2DInfo) render.Texture {
	return internal.NewDepthTexture2D(info)
}

func (a *API) CreateStencilTexture2D(info render.StencilTexture2DInfo) render.Texture {
	return internal.NewStencilTexture2D(info)
}

func (a *API) CreateDepthStencilTexture2D(info render.DepthStencilTexture2DInfo) render.Texture {
	return internal.NewDepthStencilTexture2D(info)
}

func (a *API) CreateVertexBuffer(info render.BufferInfo) render.Buffer {
	return internal.NewVertexBuffer(info)
}

func (a *API) CreateIndexBuffer(info render.BufferInfo) render.Buffer {
	return internal.NewIndexBuffer(info)
}

func (a *API) CreatePixelTransferBuffer(info render.BufferInfo) render.Buffer {
	return internal.NewPixelTransferBuffer(info)
}

func (a *API) CreateUniformBuffer(info render.BufferInfo) render.Buffer {
	return internal.NewUniformBuffer(info)
}

func (a *API) CreateVertexArray(info render.VertexArrayInfo) render.VertexArray {
	return internal.NewVertexArray(info)
}

func (a *API) CreatePipeline(info render.PipelineInfo) render.Pipeline {
	return internal.NewPipeline(info)
}

func (a *API) CreateCommandQueue() render.CommandQueue {
	return internal.NewCommandQueue()
}

func (a *API) DetermineContentFormat(framebuffer render.Framebuffer) render.DataFormat {
	return internal.DetermineContentFormat(framebuffer)
}

func (a *API) BeginRenderPass(info render.RenderPassInfo) {
	a.renderer.BeginRenderPass(info)
}

func (a *API) EndRenderPass() {
	a.renderer.EndRenderPass()
}

func (a *API) Invalidate() {
	a.renderer.Invalidate()
}

func (a *API) BindPipeline(pipeline render.Pipeline) {
	a.renderer.BindPipeline(pipeline)
}

func (a *API) Uniform1f(location render.UniformLocation, value float32) {
	a.renderer.Uniform1f(location, value)
}

func (a *API) Uniform1i(location render.UniformLocation, value int) {
	a.renderer.Uniform1i(location, value)
}

func (a *API) Uniform3f(location render.UniformLocation, values [3]float32) {
	a.renderer.Uniform3f(location, values)
}

func (a *API) Uniform4f(location render.UniformLocation, values [4]float32) {
	a.renderer.Uniform4f(location, values)
}

func (a *API) UniformMatrix4f(location render.UniformLocation, values [16]float32) {
	a.renderer.UniformMatrix4f(location, values)
}

func (a *API) UniformBufferUnit(index int, buffer render.Buffer) {
	a.renderer.UniformBufferUnit(index, buffer)
}

func (a *API) UniformBufferUnitRange(index int, buffer render.Buffer, offset, size int) {
	a.renderer.UniformBufferUnitRange(index, buffer, offset, size)
}

func (a *API) TextureUnit(index int, texture render.Texture) {
	a.renderer.TextureUnit(index, texture)
}

func (a *API) Draw(vertexOffset, vertexCount, instanceCount int) {
	a.renderer.Draw(vertexOffset, vertexCount, instanceCount)
}

func (a *API) DrawIndexed(indexOffset, indexCount, instanceCount int) {
	a.renderer.DrawIndexed(indexOffset, indexCount, instanceCount)
}

func (a *API) CopyContentToTexture(info render.CopyContentToTextureInfo) {
	a.renderer.CopyContentToTexture(info)
}

func (a *API) SubmitQueue(queue render.CommandQueue) {
	a.renderer.SubmitQueue(queue.(*internal.CommandQueue))
}

func (a *API) CreateFence() render.Fence {
	return internal.NewFence()
}
