package internal

import (
	"math"
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
)

// CompressionFilter is a stereo Processor that applies dynamic range
// compression using a soft-knee gain computer and per-sample IIR gain
// smoothing. All parameters are safe to set from any goroutine.
type CompressionFilter struct {
	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu        sync.Mutex
	attack    float32
	release   float32
	ratio     float32
	knee      float32
	threshold float32

	// The following fields are used only during processing.
	currentGain float32
}

var _ audio.Compression = (*CompressionFilter)(nil)
var _ Processor = (*CompressionFilter)(nil)

// NewCompressionFilter creates a CompressionFilter with moderate defaults:
// 3 ms attack, 250 ms release, 12:1 ratio, -24 dB threshold, 30 dB knee.
func NewCompressionFilter() *CompressionFilter {
	return &CompressionFilter{
		attack:    0.003,
		release:   0.25,
		ratio:     12.0,
		knee:      30.0,
		threshold: -24.0,

		currentGain: 1.0,
	}
}

// Attack returns the attack time in seconds.
func (f *CompressionFilter) Attack() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.attack
}

// SetAttack sets the attack time in seconds, clamped to [0, 1].
func (f *CompressionFilter) SetAttack(attack float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.attack = sprec.Clamp(attack, 0.0, 1.0)
}

// Release returns the release time in seconds.
func (f *CompressionFilter) Release() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.release
}

// SetRelease sets the release time in seconds, clamped to [0, 1].
func (f *CompressionFilter) SetRelease(release float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.release = sprec.Clamp(release, 0.0, 1.0)
}

// Ratio returns the compression ratio.
func (f *CompressionFilter) Ratio() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.ratio
}

// SetRatio sets the compression ratio, clamped to [1, 20].
func (f *CompressionFilter) SetRatio(ratio float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.ratio = sprec.Clamp(ratio, 1.0, 20.0)
}

// Knee returns the knee width in decibels.
func (f *CompressionFilter) Knee() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.knee
}

// SetKnee sets the knee width in decibels, clamped to [0, 40].
func (f *CompressionFilter) SetKnee(knee float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.knee = sprec.Clamp(knee, 0.0, 40.0)
}

// Threshold returns the threshold level in decibels.
func (f *CompressionFilter) Threshold() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.threshold
}

// SetThreshold sets the threshold level in decibels, clamped to [-100, 0].
func (f *CompressionFilter) SetThreshold(threshold float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.threshold = sprec.Clamp(threshold, -100.0, 0.0)
}

// Process applies dynamic range compression to inputFrames and returns the result.
func (f *CompressionFilter) Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame {
	f.mu.Lock()
	attack := f.attack       // store value locally to avoid long locks
	release := f.release     // store value locally to avoid long locks
	ratio := f.ratio         // store value locally to avoid long locks
	knee := f.knee           // store value locally to avoid long locks
	threshold := f.threshold // store value locally to avoid long locks
	f.mu.Unlock()

	sampleRate := float32(ctx.SampleRate)
	invRatioNegOne := (1.0 / ratio) - 1.0
	attackCoeff := float32(math.Exp(-1.0 / (float64(sampleRate * attack))))
	releaseCoeff := float32(math.Exp(-1.0 / (float64(sampleRate * release))))

	outputFrames := ctx.Buffer.Allocate(ctx.FrameCount)

	for i, frame := range inputFrames {
		peak := max(sprec.Abs(frame.Left), sprec.Abs(frame.Right))
		peak = max(1.0e-8, peak) // avoid log of zero
		peakDB := audio.GainToDB(peak)

		reductionDB := float32(0.0)
		if peakDB > threshold { // needs compression
			z := peakDB - threshold
			if peakDB >= (threshold + knee) { // hard compression
				reductionDB = invRatioNegOne * (z - knee/2.0)
			} else { // soft compression
				reductionDB = invRatioNegOne * (z * z) / (2.0 * knee)
			}
		}

		targetGain := audio.DBToGain(reductionDB)
		if targetGain < f.currentGain {
			f.currentGain = sprec.Mix(targetGain, f.currentGain, attackCoeff)
		} else {
			f.currentGain = sprec.Mix(targetGain, f.currentGain, releaseCoeff)
		}

		outputFrames[i] = audio.Frame{
			Left:  frame.Left * f.currentGain,
			Right: frame.Right * f.currentGain,
		}
	}

	return outputFrames
}
