package internal

import "github.com/mokiat/gomath/sprec"

// HighPassSampleFilter implements a second-order Butterworth high-pass filter.
type HighPassSampleFilter struct {
	sampleRate int

	b0 float32
	b1 float32
	b2 float32
	a1 float32
	a2 float32

	inputNeg1  float32
	inputNeg2  float32
	outputNeg1 float32
	outputNeg2 float32
}

var _ SampleProcessor = (*HighPassSampleFilter)(nil)

// NewHighPassSampleFilter creates a HighPassSampleFilter configured with a
// 350 Hz cutoff for the given sample rate.
func NewHighPassSampleFilter(sampleRate int) *HighPassSampleFilter {
	result := &HighPassSampleFilter{
		sampleRate: sampleRate,
	}
	result.Configure(350.0)
	return result
}

// Configure sets the cutoff frequency in Hz for the given sample rate.
func (f *HighPassSampleFilter) Configure(cutoffFrequency float32) {
	const fcEpsilon = 0.1
	fs := float32(f.sampleRate)
	fc := min(max(fcEpsilon, cutoffFrequency), (fs/2.0)-fcEpsilon)

	// This uses a second-order high-pass filter algorithm based on the
	// Audio EQ Cookbook by Robert Bristow-Johnson.
	// Reference: https://www.w3.org/TR/audio-eq-cookbook/
	angle := sprec.Radians(2.0 * sprec.Pi * (fc / fs))
	cs := sprec.Cos(angle)
	sn := sprec.Sin(angle)
	alpha := sn / sprec.Sqrt(2.0)

	a0 := 1.0 + alpha
	f.b0 = ((1.0 + cs) / 2.0) / a0
	f.b1 = -(1.0 + cs) / a0
	f.b2 = ((1.0 + cs) / 2.0) / a0
	f.a1 = (-2.0 * cs) / a0
	f.a2 = (1.0 - alpha) / a0
}

// ProcessSample filters one sample and returns the high-pass output.
func (f *HighPassSampleFilter) ProcessSample(input float32) float32 {
	result := f.b0*input +
		f.b1*f.inputNeg1 +
		f.b2*f.inputNeg2 -
		f.a1*f.outputNeg1 -
		f.a2*f.outputNeg2

	f.inputNeg2 = f.inputNeg1
	f.inputNeg1 = input
	f.outputNeg2 = f.outputNeg1
	f.outputNeg1 = result

	return result
}
