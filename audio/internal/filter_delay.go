package internal

func NewDelayFilter(maxDelaySamples int) *DelayFilter {
	return &DelayFilter{
		maxDelaySamples: maxDelaySamples,
		buffer:          make([]float32, maxDelaySamples+1),
	}
}

type DelayFilter struct {
	maxDelaySamples int
	buffer          []float32
	delayOffset     int64
	writeOffset     int64
}

func (f *DelayFilter) Configure(delaySamples int) {
	delaySamples = min(max(0, delaySamples), f.maxDelaySamples)
	f.delayOffset = -int64(delaySamples)
}

func (f *DelayFilter) Process(input float32) float32 {
	// Note: This implementation writes and reads to and from the delay buffer
	// at the same time. It also orders samples in opposite order compared to the
	// input and output buffers.

	bufferSize := int64(len(f.buffer))

	readOffset := (f.writeOffset + f.delayOffset + bufferSize) % bufferSize

	f.buffer[f.writeOffset] = input
	f.writeOffset = (f.writeOffset + 1) % bufferSize

	return f.buffer[readOffset]
}
