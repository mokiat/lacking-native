package internal

import "github.com/mokiat/lacking/core/audio"

// ProcessContext provides context information for an audio processing call.
type ProcessContext struct {

	// SampleRate indicates the sample rate (in Hz) at which audio is being
	// processed.
	SampleRate int

	// FrameCount indicates the number of audio frames to be processed in this
	// call.
	FrameCount int

	// Buffer is a reusable buffer for audio processing. It can be used to
	// allocate temporary storage for audio frames during processing. Frames
	// returned by the buffer must not be retained after the processing call
	// returns, as they may be overwritten in subsequent calls.
	Buffer *Buffer
}

// Processor represents a single processing unit in the audio graph.
type Processor interface {

	// Process processes the given input frames and produces output frames based
	// on the provided context. The number of output frames must match the FrameCount
	// specified in the context.
	//
	// WARNING: This method may be called from a real-time audio thread. It must
	// not perform any blocking operations, memory allocations, or other
	// operations that could disrupt real-time audio processing.
	Process(ctx ProcessContext, inputFrames []audio.Frame) []audio.Frame
}

type Pipeline struct {
	Units []Unit
}

func NewPipeline(initialCapacity int) *Pipeline {
	return &Pipeline{
		Units: make([]Unit, 0, initialCapacity),
	}
}

type Unit struct {
	Processor   Processor
	TargetIndex int
}
