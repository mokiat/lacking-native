package internal

import (
	"github.com/mokiat/lacking/core/audio"
)

// BasePlayback holds the common fields and methods shared by DefaultPlayback
// and SpatialPlayback. It is not meant to be used directly as a Processor —
// use DefaultPlayback or SpatialPlayback instead.
type BasePlayback struct {
	bus *Bus

	source         *PlaybackSource
	gainFilter     *GainFilter
	lowPassFilter  *LowPassFilter
	highPassFilter *HighPassFilter
}

var _ Processor = (*BasePlayback)(nil)

// NewBasePlayback creates a BasePlayback for the given bus and media.
// Low-pass and high-pass filters are allocated only when enabled in settings.
func NewBasePlayback(worker Worker, bus *Bus, media *Media, settings audio.PlaybackSettings, sampleRate int) *BasePlayback {
	result := &BasePlayback{
		bus: bus,

		source:     NewPlaybackSource(worker, media, sampleRate),
		gainFilter: NewGainFilter(),
	}
	if settings.UseLowPassFilter {
		result.lowPassFilter = NewLowPassFilter(sampleRate)
	}
	if settings.UseHighPassFilter {
		result.highPassFilter = NewHighPassFilter(sampleRate)
	}
	return result
}

// Start begins playback from the given time offset in seconds.
func (p *BasePlayback) Start(at float32) {
	p.source.Start(at)
}

// Stop halts playback and resets the position to the beginning.
func (p *BasePlayback) Stop() {
	p.source.Stop()
}

// Pause suspends playback without resetting the position.
func (p *BasePlayback) Pause() {
	p.source.Pause()
}

// Resume continues a previously paused playback.
func (p *BasePlayback) Resume() {
	p.source.Resume()
}

// Looping reports whether the playback loops when it reaches the end.
func (p *BasePlayback) Looping() bool {
	return p.source.Looping()
}

// SetLooping enables or disables looping.
func (p *BasePlayback) SetLooping(loop bool) {
	p.source.SetLooping(loop)
}

// LoopStart returns the loop start position in seconds.
func (p *BasePlayback) LoopStart() float32 {
	return p.source.LoopStart()
}

// SetLoopStart sets the loop start position in seconds.
func (p *BasePlayback) SetLoopStart(loopStart float32) {
	p.source.SetLoopStart(loopStart)
}

// LoopEnd returns the loop end position in seconds.
func (p *BasePlayback) LoopEnd() float32 {
	return p.source.LoopEnd()
}

// SetLoopEnd sets the loop end position in seconds.
func (p *BasePlayback) SetLoopEnd(loopEnd float32) {
	p.source.SetLoopEnd(loopEnd)
}

// Playing reports whether this playback is currently active.
func (p *BasePlayback) Playing() bool {
	return p.source.Playing()
}

// PlaybackRate returns the playback speed multiplier; 1.0 is normal speed.
func (p *BasePlayback) PlaybackRate() float32 {
	return p.source.PlaybackRate()
}

// SetPlaybackRate sets the playback speed multiplier.
func (p *BasePlayback) SetPlaybackRate(rate float32) {
	p.source.SetPlaybackRate(rate)
}

// Gain returns the current gain of this playback.
func (p *BasePlayback) Gain() float32 {
	return p.gainFilter.Gain()
}

// SetGain sets the gain of this playback.
func (p *BasePlayback) SetGain(gain float32) {
	p.gainFilter.SetGain(gain)
}

// LowPassFilter returns the low-pass filter controls, or nil if the filter was
// not enabled at creation time.
func (p *BasePlayback) LowPassFilter() audio.FrequencyFilter {
	if p.lowPassFilter == nil {
		return nil
	}
	return p.lowPassFilter
}

// HighPassFilter returns the high-pass filter controls, or nil if the filter
// was not enabled at creation time.
func (p *BasePlayback) HighPassFilter() audio.FrequencyFilter {
	if p.highPassFilter == nil {
		return nil
	}
	return p.highPassFilter
}

// SetOnFinished sets a callback invoked when the playback reaches its end.
func (p *BasePlayback) SetOnFinished(onFinished func()) {
	p.source.SetOnFinished(onFinished)
}

// Process produces audio frames by running the source through optional
// frequency filters and gain, in that order.
func (p *BasePlayback) Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame {
	outputFrames := p.source.Process(ctx, inputFrames)
	if p.lowPassFilter != nil {
		outputFrames = p.lowPassFilter.Process(ctx, outputFrames)
	}
	if p.highPassFilter != nil {
		outputFrames = p.highPassFilter.Process(ctx, outputFrames)
	}
	outputFrames = p.gainFilter.Process(ctx, outputFrames)
	return outputFrames
}
