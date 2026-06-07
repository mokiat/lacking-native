package internal

import (
	"sync"

	"github.com/mokiat/lacking/core/audio"
)

// GainFilter is a stereo Processor that scales each frame by a gain factor.
// The gain is safe to set from any goroutine.
type GainFilter struct {
	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu   sync.Mutex
	gain float32
}

var _ Processor = (*GainFilter)(nil)

// NewGainFilter creates a GainFilter with unity gain.
func NewGainFilter() *GainFilter {
	return &GainFilter{
		gain: 1.0,
	}
}

// Gain returns the current gain factor.
func (f *GainFilter) Gain() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.gain
}

// SetGain sets the gain factor, clamped to a minimum of zero.
func (f *GainFilter) SetGain(gain float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.gain = max(0.0, gain)
}

// Process applies the gain to inputFrames and returns the result.
func (f *GainFilter) Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame {
	f.mu.Lock()
	gain := f.gain // store value locally to avoid long locks
	f.mu.Unlock()

	outputFrames := ctx.Buffer.Allocate(ctx.FrameCount)

	const gainThreshold = 0.001
	if gain < gainThreshold {
		return outputFrames
	}

	for i, frame := range inputFrames {
		outputFrames[i] = audio.Frame{
			Left:  frame.Left * gain,
			Right: frame.Right * gain,
		}
	}
	return outputFrames
}
