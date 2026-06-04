package internal

import (
	"sync"

	"github.com/mokiat/lacking/core/audio"
)

type GainFilter struct {
	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu   sync.Mutex
	gain float32
}

var _ Processor = (*GainFilter)(nil)

func NewGainFilter() *GainFilter {
	return &GainFilter{
		gain: 1.0,
	}
}

func (f *GainFilter) Gain() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.gain
}

func (f *GainFilter) SetGain(gain float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.gain = max(0.0, gain)
}

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
