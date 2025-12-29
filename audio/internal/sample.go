package internal

// Sample represents a single audio sample with left and right channel data.
type Sample struct {

	// Left is the left channel value.
	Left float64

	// Right is the right channel value.
	Right float64
}

// Add adds another sample to this sample.
func (s *Sample) Add(other Sample) {
	s.Left += other.Left
	s.Right += other.Right
}

// SampleContext provides context information for audio sample processing.
type SampleContext struct {

	// SampleRate is the number of samples per second.
	SampleRate int
}

// Rate returns the sample rate as a float64.
func (c SampleContext) Rate() float64 {
	return float64(c.SampleRate)
}

// SampleLength returns the length of a single sample in seconds.
func (c SampleContext) SampleLength() float64 {
	return 1.0 / float64(c.SampleRate)
}
