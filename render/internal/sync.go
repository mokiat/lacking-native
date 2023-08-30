package internal

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/mokiat/lacking/render"
)

func NewFence() *Fence {
	return &Fence{
		id: gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0),
	}
}

type Fence struct {
	render.FenceObject
	id uintptr
}

func (f *Fence) Status() render.FenceStatus {
	var status, count int32
	gl.GetSynciv(f.id, gl.SYNC_STATUS, 1, &count, &status)
	switch status {
	case gl.SIGNALED:
		return render.FenceStatusSuccess
	case gl.UNSIGNALED:
		return render.FenceStatusNotReady
	default:
		return render.FenceStatusDeviceLost
	}
}

func (f *Fence) Delete() {
	gl.DeleteSync(f.id)
	f.id = 0
}
