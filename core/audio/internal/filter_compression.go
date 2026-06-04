package internal

import (
	"math"
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
)

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

func (f *CompressionFilter) Attack() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.attack
}

func (f *CompressionFilter) SetAttack(attack float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.attack = sprec.Clamp(attack, 0.0, 1.0)
}

func (f *CompressionFilter) Release() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.release
}

func (f *CompressionFilter) SetRelease(release float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.release = sprec.Clamp(release, 0.0, 1.0)
}

func (f *CompressionFilter) Ratio() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.ratio
}

func (f *CompressionFilter) SetRatio(ratio float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.ratio = sprec.Clamp(ratio, 1.0, 20.0)
}

func (f *CompressionFilter) Knee() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.knee
}

func (f *CompressionFilter) SetKnee(knee float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.knee = sprec.Clamp(knee, 0.0, 40.0)
}

func (f *CompressionFilter) Threshold() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.threshold
}

func (f *CompressionFilter) SetThreshold(threshold float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.threshold = sprec.Clamp(threshold, -100.0, 0.0)
}

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
