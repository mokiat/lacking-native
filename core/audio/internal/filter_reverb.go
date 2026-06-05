package internal

import (
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
)

// ReverbFilter is a stereo Processor that applies a Freeverb-style reverb
// effect using parallel feedback comb filters followed by series all-pass
// diffusers. All parameters are safe to set from any goroutine.
type ReverbFilter struct {
	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu       sync.Mutex
	roomSize float32
	damping  float32
	dry      float32
	wet      float32

	// The following fields are used only during processing.
	combDelaysL [reverbCombFilterCount]int
	combDelaysR [reverbCombFilterCount]int
	combsL      [reverbCombFilterCount]*FeedbackCombSampleFilter
	combsR      [reverbCombFilterCount]*FeedbackCombSampleFilter
	allPassesL  [reverbAllPassFilterCount]*AllPassSampleFilter
	allPassesR  [reverbAllPassFilterCount]*AllPassSampleFilter
}

var _ audio.Reverb = (*ReverbFilter)(nil)
var _ Processor = (*ReverbFilter)(nil)

// NewReverbFilter creates a ReverbFilter with moderate room size and damping
// defaults for the given sample rate.
func NewReverbFilter(sampleRate int) *ReverbFilter {
	combDelaysL := [reverbCombFilterCount]int{}
	combDelaysR := [reverbCombFilterCount]int{}
	for i := range reverbCombFilterCount {
		spread := reverbCombFilterSpread[i]
		combDelaysL[i] = audio.SampleCount(reverbCombFilterDelays[i], sampleRate)
		combDelaysR[i] = audio.SampleCount(reverbCombFilterDelays[i]+spread, sampleRate)
	}

	combsL := [reverbCombFilterCount]*FeedbackCombSampleFilter{}
	combsR := [reverbCombFilterCount]*FeedbackCombSampleFilter{}
	for i := range reverbCombFilterCount {
		combsL[i] = NewFeedbackCombSampleFilter(combDelaysL[i])
		combsL[i].Configure(0.0, 0.0, combDelaysL[i])
		combsR[i] = NewFeedbackCombSampleFilter(combDelaysR[i])
		combsR[i].Configure(0.0, 0.0, combDelaysR[i])
	}

	allPassDelaysL := [reverbAllPassFilterCount]int{}
	allPassDelaysR := [reverbAllPassFilterCount]int{}
	for i := range reverbAllPassFilterCount {
		spread := reverbAllPassFilterSpread[i]
		allPassDelaysL[i] = audio.SampleCount(reverbAllPassFilterDelays[i], sampleRate)
		allPassDelaysR[i] = audio.SampleCount(reverbAllPassFilterDelays[i]+spread, sampleRate)
	}

	allPassesL := [reverbAllPassFilterCount]*AllPassSampleFilter{}
	allPassesR := [reverbAllPassFilterCount]*AllPassSampleFilter{}
	for i := range reverbAllPassFilterCount {
		allPassesL[i] = NewAllPassSampleFilter(allPassDelaysL[i])
		allPassesL[i].Configure(reverbAllPassFilterFeedback, allPassDelaysL[i])
		allPassesR[i] = NewAllPassSampleFilter(allPassDelaysR[i])
		allPassesR[i].Configure(reverbAllPassFilterFeedback, allPassDelaysR[i])
	}

	return &ReverbFilter{
		roomSize: 0.3,
		damping:  0.5,
		dry:      1.0,
		wet:      0.5,

		combDelaysL: combDelaysL,
		combDelaysR: combDelaysR,
		combsL:      combsL,
		combsR:      combsR,
		allPassesL:  allPassesL,
		allPassesR:  allPassesR,
	}
}

// RoomSize returns the room size in the range [0, 1].
func (f *ReverbFilter) RoomSize() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.roomSize
}

// SetRoomSize sets the room size, clamped to [0, 1].
func (f *ReverbFilter) SetRoomSize(size float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.roomSize = sprec.Clamp(size, 0.0, 1.0)
}

// Damping returns the high-frequency damping factor in the range [0, 1].
func (f *ReverbFilter) Damping() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.damping
}

// SetDamping sets the high-frequency damping factor, clamped to [0, 1].
func (f *ReverbFilter) SetDamping(damping float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.damping = sprec.Clamp(damping, 0.0, 1.0)
}

// Dry returns the dry (unprocessed) mix level in the range [0, 1].
func (f *ReverbFilter) Dry() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.dry
}

// SetDry sets the dry mix level, clamped to [0, 1].
func (f *ReverbFilter) SetDry(dry float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.dry = sprec.Clamp(dry, 0.0, 1.0)
}

// Wet returns the wet (reverberated) mix level in the range [0, 1].
func (f *ReverbFilter) Wet() float32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.wet
}

// SetWet sets the wet mix level, clamped to [0, 1].
func (f *ReverbFilter) SetWet(wet float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.wet = sprec.Clamp(wet, 0.0, 1.0)
}

// Process applies the reverb effect to inputFrames and returns the result.
func (f *ReverbFilter) Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame {
	f.mu.Lock()
	roomSize := f.roomSize // store value locally to avoid long locks
	damping := f.damping   // store value locally to avoid long locks
	dryAmount := f.dry     // store value locally to avoid long locks
	wetAmount := f.wet     // store value locally to avoid long locks
	f.mu.Unlock()

	for i := range reverbCombFilterCount {
		feedback := 0.3 + roomSize*0.6
		f.combsL[i].Configure(feedback, damping, f.combDelaysL[i])
		f.combsR[i].Configure(feedback, damping, f.combDelaysR[i])
	}

	outputFrames := ctx.Buffer.Allocate(ctx.FrameCount)

	for i, frame := range inputFrames {
		outputLeft := float32(0.0)
		outputRight := float32(0.0)
		for c := range reverbCombFilterCount {
			outputLeft += f.combsL[c].ProcessSample(frame.Left)
			outputRight += f.combsR[c].ProcessSample(frame.Right)
		}
		outputLeft *= reverbCombFilterScale
		outputRight *= reverbCombFilterScale

		for j := range reverbAllPassFilterCount {
			outputLeft = f.allPassesL[j].ProcessSample(outputLeft)
			outputRight = f.allPassesR[j].ProcessSample(outputRight)
		}

		wetLeft := sprec.Mix(outputLeft, outputRight, 0.2)
		wetRight := sprec.Mix(outputRight, outputLeft, 0.2)

		outputFrames[i] = audio.Frame{
			Left:  frame.Left*dryAmount + wetLeft*wetAmount,
			Right: frame.Right*dryAmount + wetRight*wetAmount,
		}
	}

	return outputFrames
}

const (
	reverbCombFilterCount       = 8
	reverbCombFilterScale       = 1.0 / float32(reverbCombFilterCount)
	reverbAllPassFilterCount    = 4
	reverbAllPassFilterFeedback = 0.5
)

var (
	// Delay times back-calculated from Freeverb's original sample counts at 44100 Hz.
	// They are stored as seconds so they scale correctly at other sample rates.
	reverbCombFilterDelays = [reverbCombFilterCount]float32{
		0.025306, // 1116 samples @ 44100 Hz
		0.026939, // 1188 samples @ 44100 Hz
		0.028957, // 1277 samples @ 44100 Hz
		0.030748, // 1356 samples @ 44100 Hz
		0.032245, // 1422 samples @ 44100 Hz
		0.033810, // 1491 samples @ 44100 Hz
		0.035306, // 1557 samples @ 44100 Hz
		0.036667, // 1617 samples @ 44100 Hz
	}
	reverbCombFilterSpread = [reverbCombFilterCount]float32{
		0.000522, // ~23 samples @ 44100 Hz
		0.000522,
		0.000522,
		0.000522,
		0.000522,
		0.000522,
		0.000522,
		0.000522,
	}
	reverbAllPassFilterDelays = [reverbAllPassFilterCount]float32{
		0.012608, // 556 samples @ 44100 Hz
		0.010000, // 441 samples @ 44100 Hz
		0.007732, // 341 samples @ 44100 Hz
		0.005102, // 225 samples @ 44100 Hz
	}
	reverbAllPassFilterSpread = [reverbAllPassFilterCount]float32{
		0.000522, // ~23 samples @ 44100 Hz
		0.000522,
		0.000522,
		0.000522,
	}
)
