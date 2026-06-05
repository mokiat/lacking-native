package internal

// AllPassSampleFilter implements a Schroeder all-pass filter, providing flat
// magnitude response with frequency-dependent phase shift. Typical use is in
// reverb diffusion stages.
type AllPassSampleFilter struct {
	maxDelaySamples int
	buffer          []float32
	delayOffset     int
	writeOffset     int
	feedback        float32
}

var _ SampleProcessor = (*AllPassSampleFilter)(nil)

// NewAllPassSampleFilter creates an AllPassSampleFilter with a delay line
// large enough to hold maxDelaySamples samples.
func NewAllPassSampleFilter(maxDelaySamples int) *AllPassSampleFilter {
	return &AllPassSampleFilter{
		maxDelaySamples: maxDelaySamples,
		buffer:          make([]float32, maxDelaySamples+1),
	}
}

// Configure sets the feedback coefficient g and delay length in samples.
// delaySamples is clamped to [0, maxDelaySamples].
func (f *AllPassSampleFilter) Configure(feedback float32, delaySamples int) {
	f.feedback = feedback
	delaySamples = min(max(0, delaySamples), f.maxDelaySamples)
	f.delayOffset = -delaySamples
}

// ProcessSample filters one sample and returns the all-pass output.
func (f *AllPassSampleFilter) ProcessSample(input float32) float32 {
	bufferSize := len(f.buffer)

	readOffset := (f.writeOffset + f.delayOffset + bufferSize) % bufferSize
	buffered := f.buffer[readOffset]
	output := buffered - input*f.feedback

	f.buffer[f.writeOffset] = input + output*f.feedback
	f.writeOffset = (f.writeOffset + 1) % bufferSize

	return output
}
