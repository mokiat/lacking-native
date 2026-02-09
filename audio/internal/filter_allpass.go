package internal

func NewAllPassFilter(maxDelaySamples int) *AllPassFilter {
	return &AllPassFilter{
		maxDelaySamples: maxDelaySamples,
		buffer:          make([]float32, maxDelaySamples+1),
	}
}

type AllPassFilter struct {
	maxDelaySamples int
	buffer          []float32
	delayOffset     int64
	writeOffset     int64
	feedback        float32
}

func (f *AllPassFilter) Configure(feedback float32, delaySamples int) {
	f.feedback = feedback
	delaySamples = min(max(0, delaySamples), f.maxDelaySamples)
	f.delayOffset = -int64(delaySamples)
}

func (f *AllPassFilter) Process(input float32) float32 {
	bufferSize := int64(len(f.buffer))

	readOffset := (f.writeOffset + f.delayOffset + bufferSize) % bufferSize
	buffered := f.buffer[readOffset]
	output := buffered - input*f.feedback

	f.buffer[f.writeOffset] = input + output*f.feedback
	f.writeOffset = (f.writeOffset + 1) % bufferSize

	return output
}
