package internal

import (
	"math"
	"sync"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
)

// PlaybackSource is a Processor that reads decoded frames from a Media,
// performs sample-rate conversion via linear interpolation, and manages
// playback state (play, pause, stop, loop, rate). State changes from the
// game thread are communicated to the audio thread via a single pending
// playbackChange under a mutex. When playback finishes naturally the
// completion callback is dispatched through the Worker.
type PlaybackSource struct {
	worker     Worker
	sampleRate int

	frames     []audio.Frame
	framesRate int
	length     float32

	onFinished func()

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu        sync.Mutex
	change    opt.T[playbackChange]
	loopStart float32
	loopEnd   float32
	loop      bool
	rate      float32
	playing   bool

	// The following fields are used only during processing
	// and should be changed only from the processing thread.
	offset float64
}

var _ Processor = (*PlaybackSource)(nil)

// NewPlaybackSource creates a PlaybackSource that decodes from media at the
// given output sample rate. worker is used to dispatch the onFinished callback
// off the real-time audio thread.
func NewPlaybackSource(worker Worker, media *Media, sampleRate int) *PlaybackSource {
	return &PlaybackSource{
		worker:     worker,
		sampleRate: sampleRate,

		frames:     media.frames,
		framesRate: media.sampleRate,
		length:     media.Length(),

		change:    opt.Unspecified[playbackChange](),
		loopStart: 0.0,
		loopEnd:   media.Length(),
		loop:      false,
		rate:      1.0,
		playing:   false,

		offset: 0.0,
	}
}

// Start schedules playback to begin from the given position in seconds,
// clamped to [0, length].
func (s *PlaybackSource) Start(at float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.change = opt.V(playbackChange{
		position:       sprec.Clamp(at, 0.0, s.length),
		changePosition: true,
		playing:        true,
	})
}

// Stop schedules playback to halt and resets the position to the beginning.
func (s *PlaybackSource) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.change = opt.V(playbackChange{
		position:       0.0,
		changePosition: true,
		playing:        false,
	})
}

// Pause schedules playback to be suspended without resetting the position.
// If a position change is already pending, the position change is preserved.
func (s *PlaybackSource) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.change.Specified {
		s.change.Value.playing = false
	} else {
		s.change = opt.V(playbackChange{
			playing: false,
		})
	}
}

// Resume schedules a previously paused playback to continue.
// If a position change is already pending, the position change is preserved.
func (s *PlaybackSource) Resume() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.change.Specified {
		s.change.Value.playing = true
	} else {
		s.change = opt.V(playbackChange{
			playing: true,
		})
	}
}

// Looping reports whether the playback is set to loop.
func (s *PlaybackSource) Looping() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.loop
}

// SetLooping enables or disables looping.
func (s *PlaybackSource) SetLooping(loop bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.loop = loop
}

// LoopStart returns the loop start position in seconds.
func (s *PlaybackSource) LoopStart() float32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.loopStart
}

// SetLoopStart sets the loop start position in seconds.
func (s *PlaybackSource) SetLoopStart(loopStart float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.loopStart = loopStart
}

// LoopEnd returns the loop end position in seconds.
func (s *PlaybackSource) LoopEnd() float32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.loopEnd
}

// SetLoopEnd sets the loop end position in seconds.
func (s *PlaybackSource) SetLoopEnd(loopEnd float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.loopEnd = loopEnd
}

// Playing reports whether the playback is currently active.
func (s *PlaybackSource) Playing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.playing
}

// PlaybackRate returns the playback speed multiplier; 1.0 is normal speed.
func (s *PlaybackSource) PlaybackRate() float32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rate
}

// SetPlaybackRate sets the playback speed multiplier.
func (s *PlaybackSource) SetPlaybackRate(rate float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rate = rate
}

// SetOnFinished sets a callback invoked on the game thread when playback
// reaches its natural end (not on Stop, Pause, or each loop iteration).
func (s *PlaybackSource) SetOnFinished(onFinished func()) {
	s.onFinished = onFinished
}

// Process produces output frames by reading from the source media, applying
// sample-rate conversion, and advancing the playback position. Called
// exclusively from the real-time audio thread.
func (s *PlaybackSource) Process(ctx ProcessContext, _ []audio.Frame) []audio.Frame {
	s.mu.Lock()
	if change, ok := s.change.Unwrap(); ok {
		s.playing = change.playing
		if change.changePosition {
			s.offset = float64(change.position) * float64(s.framesRate)
		}
		s.change = opt.Unspecified[playbackChange]()
	}
	loopStart := float64(s.loopStart) // store value locally to avoid long locks
	loopEnd := float64(s.loopEnd)     // store value locally to avoid long locks
	loop := s.loop                    // store value locally to avoid long locks
	rate := float64(s.rate)           // store value locally to avoid long locks
	playing := s.playing              // store value locally to avoid long locks
	s.mu.Unlock()

	outputFrames := ctx.Buffer.Allocate(ctx.FrameCount)

	if !playing {
		return outputFrames
	}

	lenFrames := len(s.frames)

	// Convert loop boundaries and length from seconds to frame indices.
	loopEndFrame := min(loopEnd, float64(s.length)) * float64(s.framesRate)
	loopStartFrame := loopStart * float64(s.framesRate)
	if loop && (loopStartFrame >= loopEndFrame) {
		loopStartFrame = 0
		loopEndFrame = float64(lenFrames)
	}
	lengthFrames := float64(lenFrames)

	advanceRate := rate * float64(s.framesRate) / float64(s.sampleRate)

	offset := s.offset
	for i := range outputFrames {
		base := math.Floor(offset)
		index := int(base)
		fraction := float32(offset - base)

		if (index >= 0) && (index < lenFrames) {
			nextIndex := min(index+1, lenFrames-1)
			outputFrames[i] = audio.Frame{
				Left:  sprec.Mix(s.frames[index].Left, s.frames[nextIndex].Left, fraction),
				Right: sprec.Mix(s.frames[index].Right, s.frames[nextIndex].Right, fraction),
			}
		}

		offset += advanceRate
		if loop && (offset >= loopEndFrame) {
			offset = loopStartFrame
		}
	}
	s.offset = offset

	stillPlaying := (offset >= 0.0) && (offset < lengthFrames)
	if !stillPlaying {
		s.worker.Schedule(s.notifyFinished)
	}

	s.mu.Lock()
	if !s.change.Specified { // don't change playing state if there is a pending change
		s.playing = stillPlaying
	}
	s.mu.Unlock()

	return outputFrames
}

// notifyFinished invokes the onFinished callback if one is set. Called on the
// game thread via Worker.Schedule.
func (s *PlaybackSource) notifyFinished() {
	if s.onFinished != nil {
		s.onFinished()
	}
}

// playbackChange carries a pending state update from the game thread to the
// audio thread. Only one change is buffered at a time; a newer change
// overwrites an older one.
type playbackChange struct {
	position       float32
	changePosition bool
	playing        bool
}
