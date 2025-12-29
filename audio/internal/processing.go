package internal

// Processor represents an audio processing unit that can process audio frames
// in a real-time audio context.
type Processor interface {

	// Process processes audio frames for the given context. The inputFrames
	// slice contains frames from preceding nodes, while the outputFrames
	// slice is where the processed frames should be written to.
	//
	// WARNING: This method may be called from a real-time audio thread. It must
	// not perform any blocking operations, memory allocations, or other
	// operations that could disrupt real-time audio processing.
	Process(ctx ProcessContext, inputFrames, outputFrames FrameList)
}

// ProcessContext provides context information for an audio processing call.
type ProcessContext struct {

	// FrameCount indicates the number of audio frames to be processed in this
	// call.
	FrameCount uint32
}
