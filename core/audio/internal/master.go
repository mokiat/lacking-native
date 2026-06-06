package internal

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/core/audio"
)

// MasterBus is a Processor that applies master gain and compression to the
// mixed output of all child buses. It also tracks the set of active buses for
// use during pipeline construction.
type MasterBus struct {
	buses *ds.List[*Bus]

	gainFilter        *GainFilter
	compressionFilter *CompressionFilter
}

var _ audio.MasterBus = (*MasterBus)(nil)
var _ Processor = (*MasterBus)(nil)

// NewMasterBus creates a MasterBus with unity gain and a default compression
// filter.
func NewMasterBus() *MasterBus {
	return &MasterBus{
		buses: ds.EmptyList[*Bus](),

		gainFilter:        NewGainFilter(),
		compressionFilter: NewCompressionFilter(),
	}
}

// AddBus registers a child bus with the master bus.
func (b *MasterBus) AddBus(bus *Bus) {
	b.buses.Add(bus)
}

// RemoveBus unregisters a child bus from the master bus.
func (b *MasterBus) RemoveBus(bus *Bus) {
	b.buses.Remove(bus)
}

// Gain returns the master gain.
func (b *MasterBus) Gain() float32 {
	return b.gainFilter.Gain()
}

// SetGain sets the master gain.
func (b *MasterBus) SetGain(gain float32) {
	b.gainFilter.SetGain(gain)
}

// Compression returns the global compression controls.
func (b *MasterBus) Compression() audio.Compression {
	return b.compressionFilter
}

// Process applies master gain and compression to inputFrames and returns the
// result.
func (b *MasterBus) Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame {
	outputFrames := b.gainFilter.Process(ctx, inputFrames)
	outputFrames = b.compressionFilter.Process(ctx, outputFrames)
	return outputFrames
}
