package internal

import (
	"github.com/mokiat/gomath/sprec"
)

// FeedbackCombSampleFilter implements a Schroeder feedback comb filter with
// damping. Typical use is as the primary coloring stage in reverb algorithms.
type FeedbackCombSampleFilter struct {
	maxDelaySamples int
	buffer          []float32
	delayOffset     int
	writeOffset     int
	feedback        float32
	damp            float32
	filterStore     float32
}

var _ SampleProcessor = (*FeedbackCombSampleFilter)(nil)

// NewFeedbackCombSampleFilter creates a FeedbackCombSampleFilter with a delay
// line large enough to hold maxDelaySamples samples.
func NewFeedbackCombSampleFilter(maxDelaySamples int) *FeedbackCombSampleFilter {
	return &FeedbackCombSampleFilter{
		maxDelaySamples: maxDelaySamples,
		buffer:          make([]float32, maxDelaySamples+1),
	}
}

// Configure sets the feedback coefficient, damping amount, and delay length in
// samples. delaySamples is clamped to [0, maxDelaySamples].
func (f *FeedbackCombSampleFilter) Configure(feedback, damp float32, delaySamples int) {
	f.feedback = feedback
	f.damp = damp
	delaySamples = min(max(0, delaySamples), f.maxDelaySamples)
	f.delayOffset = -delaySamples
}

// ProcessSample filters one sample and returns the comb filter output.
func (f *FeedbackCombSampleFilter) ProcessSample(input float32) float32 {
	bufferSize := len(f.buffer)

	readOffset := (f.writeOffset + f.delayOffset + bufferSize) % bufferSize
	output := f.buffer[readOffset]

	f.filterStore = sprec.Mix(output, f.filterStore, f.damp)
	f.buffer[f.writeOffset] = input + f.filterStore*f.feedback
	f.writeOffset = (f.writeOffset + 1) % bufferSize

	return output
}
