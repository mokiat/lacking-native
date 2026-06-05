package internal

import (
	"sync"

	"github.com/mokiat/lacking/core/audio"
)

// LowPassFilter is a stereo Processor that applies a second-order Butterworth
// low-pass filter to each channel. The cutoff frequency is safe to set from
// any goroutine.
type LowPassFilter struct {
	sampleRate int

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu              sync.Mutex
	cutoffFrequency float32

	// The following filters are used only during processing
	// and should be configured only from the processing thread.
	filterL *LowPassSampleFilter
	filterR *LowPassSampleFilter
}

var _ Processor = (*LowPassFilter)(nil)

// NewLowPassFilter creates a LowPassFilter with a 350 Hz initial cutoff for
// the given sample rate.
func NewLowPassFilter(sampleRate int) *LowPassFilter {
	return &LowPassFilter{
		sampleRate: sampleRate,

		cutoffFrequency: 350.0,

		filterL: NewLowPassSampleFilter(sampleRate),
		filterR: NewLowPassSampleFilter(sampleRate),
	}
}

// Process applies the low-pass filter to inputFrames and returns the result.
func (f *LowPassFilter) Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame {
	f.mu.Lock()
	fc := f.cutoffFrequency // store value locally to avoid long locks
	f.mu.Unlock()
	fs := f.sampleRate

	outputFrames := ctx.Buffer.Allocate(ctx.FrameCount)

	const threshold = 0.1
	if fc <= threshold {
		return outputFrames // nothing passes through
	}

	nyquist := float32(fs) / 2.0
	if fc >= (nyquist - threshold) {
		copy(outputFrames, inputFrames)
		return outputFrames // everything passes through
	}

	f.filterL.Configure(fc)
	f.filterR.Configure(fc)

	for i := range outputFrames {
		input := inputFrames[i]
		outputFrames[i] = audio.Frame{
			Left:  f.filterL.ProcessSample(input.Left),
			Right: f.filterR.ProcessSample(input.Right),
		}
	}

	return outputFrames
}

// CutoffFrequency returns the cutoff frequency in Hz.
func (f *LowPassFilter) CutoffFrequency() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.cutoffFrequency
}

// SetCutoffFrequency sets the cutoff frequency in Hz.
func (f *LowPassFilter) SetCutoffFrequency(frequency float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.cutoffFrequency = frequency
}
