package render

import (
	"github.com/mokiat/lacking-native/render/internal"
	"github.com/mokiat/lacking/render"
)

func NewAPI() render.API {
	return &API{
		limits:   internal.NewLimits(),
		renderer: internal.NewRenderer(),
	}
}

type API struct {
	limits   *internal.Limits
	renderer *internal.Renderer
}

func (a *API) Limits() render.Limits {
	return a.limits
}

func (a *API) Capabilities() render.Capabilities {
	return render.Capabilities{
		Quality: render.QualityHigh,
	}
}

func (a *API) DefaultFramebuffer() render.Framebuffer {
	return internal.DefaultFramebuffer
}

func (a *API) DetermineContentFormat(framebuffer render.Framebuffer) render.DataFormat {
	return internal.DetermineContentFormat(framebuffer)
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

func (a *API) BeginRenderPass(info render.RenderPassInfo) {
	a.renderer.BeginRenderPass(info)
}

func (a *API) EndRenderPass() {
	a.renderer.EndRenderPass()
}

func (a *API) Invalidate() {
	a.renderer.Invalidate()
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
