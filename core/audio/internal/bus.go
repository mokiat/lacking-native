package internal

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/core/audio"
)

// Bus is a Processor that groups playbacks and applies shared gain,
// compression, and reverb effects. It triggers a pipeline invalidation
// whenever its topology changes.
type Bus struct {
	masterBus *MasterBus
	playbacks *ds.List[Processor]

	gainFilter        *GainFilter
	compressionFilter *CompressionFilter
	reverbFilter      *ReverbFilter

	paused bool

	invalidate func()
}

var _ audio.Bus = (*Bus)(nil)
var _ Processor = (*Bus)(nil)

// NewBus creates a Bus registered with the given master bus. Compression and
// reverb are only allocated when enabled in settings. invalidate is called
// whenever the bus topology changes so the pipeline can be rebuilt.
func NewBus(masterBus *MasterBus, settings audio.BusSettings, sampleRate int, invalidate func()) *Bus {
	var compressionFilter *CompressionFilter
	if settings.UseCompression {
		compressionFilter = NewCompressionFilter()
	}

	var reverbFilter *ReverbFilter
	if settings.UseReverb {
		reverbFilter = NewReverbFilter(sampleRate)
	}

	result := &Bus{
		masterBus: masterBus,
		playbacks: ds.EmptyList[Processor](),

		gainFilter:        NewGainFilter(),
		compressionFilter: compressionFilter,
		reverbFilter:      reverbFilter,

		invalidate: invalidate,
	}

	masterBus.AddBus(result)
	invalidate()
	return result
}

// Release removes the bus from the master bus and triggers a pipeline
// invalidation.
func (b *Bus) Release() {
	b.masterBus.RemoveBus(b)
	b.invalidate()
}

// AddPlayback registers a playback with this bus.
func (b *Bus) AddPlayback(playback Processor) {
	b.playbacks.Add(playback)
	b.invalidate()
}

// RemovePlayback unregisters a playback from this bus.
func (b *Bus) RemovePlayback(playback Processor) {
	b.playbacks.Remove(playback)
	b.invalidate()
}

// Playbacks returns the playbacks currently registered with this bus.
func (b *Bus) Playbacks() []Processor {
	return b.playbacks.Unbox()
}

// Gain returns the current gain of the bus.
func (b *Bus) Gain() float32 {
	return b.gainFilter.Gain()
}

// SetGain sets the gain of the bus.
func (b *Bus) SetGain(gain float32) {
	b.gainFilter.SetGain(gain)
}

// Compression returns the compression controls of the bus, or nil if
// compression was not enabled at creation time.
func (b *Bus) Compression() audio.Compression {
	if b.compressionFilter == nil {
		return nil
	}
	return b.compressionFilter
}

// Reverb returns the reverb controls of the bus, or nil if reverb was not
// enabled at creation time.
func (b *Bus) Reverb() audio.Reverb {
	if b.reverbFilter == nil {
		return nil
	}
	return b.reverbFilter
}

// Pause marks the bus as paused and triggers a pipeline invalidation.
func (b *Bus) Pause() {
	b.paused = true
	b.invalidate()
}

// Resume marks the bus as playing and triggers a pipeline invalidation.
func (b *Bus) Resume() {
	b.paused = false
	b.invalidate()
}

// IsPlaying reports whether the bus is currently unpaused.
func (b *Bus) IsPlaying() bool {
	return !b.paused
}

// Process applies gain, compression, and reverb to inputFrames and returns
// the result.
func (b *Bus) Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame {
	outputFrames := b.gainFilter.Process(ctx, inputFrames)
	if b.compressionFilter != nil {
		outputFrames = b.compressionFilter.Process(ctx, outputFrames)
	}
	if b.reverbFilter != nil {
		outputFrames = b.reverbFilter.Process(ctx, outputFrames)
	}
	return outputFrames
}
