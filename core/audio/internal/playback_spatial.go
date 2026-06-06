package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
)

// SpatialPlayback is a spatially positioned Processor that wraps a BasePlayback
// and applies 3D spatial audio processing after the base chain.
type SpatialPlayback struct {
	*BasePlayback
	spatialFilter *SpatialFilter
}

var _ audio.SpatialPlayback = (*SpatialPlayback)(nil)
var _ Processor = (*SpatialPlayback)(nil)

// NewSpatialPlayback wraps basePlayback in a SpatialPlayback, attaches the
// given listener for 3D positioning, and registers it with its bus.
func NewSpatialPlayback(basePlayback *BasePlayback, listener *SpatialListener) *SpatialPlayback {
	result := &SpatialPlayback{
		BasePlayback:  basePlayback,
		spatialFilter: NewSpatialFilter(listener),
	}
	basePlayback.bus.AddPlayback(result)
	return result
}

// Release removes this playback from its bus, triggering a pipeline rebuild.
func (p *SpatialPlayback) Release() {
	p.bus.RemovePlayback(p)
}

// Position returns the emitter's position in world space.
func (p *SpatialPlayback) Position() sprec.Vec3 {
	return p.spatialFilter.Position()
}

// SetPosition sets the emitter's position in world space.
func (p *SpatialPlayback) SetPosition(position sprec.Vec3) {
	p.spatialFilter.SetPosition(position)
}

// Rotation returns the emitter's orientation in world space. The emitter's
// forward (cone) direction is its +Z axis.
func (p *SpatialPlayback) Rotation() sprec.Quat {
	return p.spatialFilter.Rotation()
}

// SetRotation sets the emitter's orientation in world space.
func (p *SpatialPlayback) SetRotation(rotation sprec.Quat) {
	p.spatialFilter.SetRotation(rotation)
}

// InnerConeAngle returns the inner cone angle of the emitter.
func (p *SpatialPlayback) InnerConeAngle() sprec.Angle {
	return p.spatialFilter.InnerConeAngle()
}

// SetInnerConeAngle sets the inner cone angle of the emitter.
func (p *SpatialPlayback) SetInnerConeAngle(angle sprec.Angle) {
	p.spatialFilter.SetInnerConeAngle(angle)
}

// OuterConeAngle returns the outer cone angle of the emitter.
func (p *SpatialPlayback) OuterConeAngle() sprec.Angle {
	return p.spatialFilter.OuterConeAngle()
}

// SetOuterConeAngle sets the outer cone angle of the emitter.
func (p *SpatialPlayback) SetOuterConeAngle(angle sprec.Angle) {
	p.spatialFilter.SetOuterConeAngle(angle)
}

// OuterConeGain returns the gain applied when the listener is outside the
// outer cone.
func (p *SpatialPlayback) OuterConeGain() float32 {
	return p.spatialFilter.OuterConeGain()
}

// SetOuterConeGain sets the gain applied when the listener is outside the
// outer cone.
func (p *SpatialPlayback) SetOuterConeGain(gain float32) {
	p.spatialFilter.SetOuterConeGain(gain)
}

// Process runs the base chain then applies spatial processing to the result.
func (p *SpatialPlayback) Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame {
	outputFrames := p.BasePlayback.Process(ctx, inputFrames)
	outputFrames = p.spatialFilter.Process(ctx, outputFrames)
	return outputFrames
}
