package internal

import "github.com/mokiat/gomath/sprec"

func NewFeedbackCombFilter(maxDelaySamples int) *FeedbackCombFilter {
	return &FeedbackCombFilter{
		maxDelaySamples: maxDelaySamples,
		buffer:          make([]float32, maxDelaySamples+1),
	}
}

type FeedbackCombFilter struct {
	maxDelaySamples int
	buffer          []float32
	delayOffset     int64
	writeOffset     int64
	feedback        float32
	damp            float32
	filterStore     float32
}

func (f *FeedbackCombFilter) Configure(feedback, damp float32, delaySamples int) {
	f.feedback = feedback
	f.damp = damp
	delaySamples = min(max(0, delaySamples), f.maxDelaySamples)
	f.delayOffset = -int64(delaySamples)
}

func (f *FeedbackCombFilter) Process(input float32) float32 {
	bufferSize := int64(len(f.buffer))

	readOffset := (f.writeOffset + f.delayOffset + bufferSize) % bufferSize
	output := f.buffer[readOffset]

	f.filterStore = sprec.Mix(output, f.filterStore, f.damp)
	f.buffer[f.writeOffset] = input + f.filterStore*f.feedback
	f.writeOffset = (f.writeOffset + 1) % bufferSize

	return output
}
