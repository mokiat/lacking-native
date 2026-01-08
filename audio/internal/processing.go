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

	// SampleRate indicates the sample rate (in Hz) at which audio is being
	// processed.
	SampleRate uint32

	// FrameCount indicates the number of audio frames to be processed in this
	// call.
	FrameCount uint32
}

// ProcessingSnapshot represents a snapshot of the audio processing graph at a
// specific point in time. It contains the list of processing units and their
// assignments to output targets.
type ProcessingSnapshot struct {

	// Processings holds the list of processors to be processed in order.
	Processings []ProcessingUnit

	// Assignments holds the target (output) indices for each processing unit.
	// The number of assignments for each processing unit is specified
	// in the corresponding GraphProcessing.AssignmentCount field.
	Assignments []uint32
}

// IsDirectConnection checks if the processing unit at sourceIndex has a direct
// connection to a single target processing unit. If so, it returns the index of
// the target processing unit and true. Otherwise, it returns false.
func (s *ProcessingSnapshot) IsDirectConnection(sourceIndex uint32) (uint32, bool) {
	sourceUnit := &s.Processings[sourceIndex]
	if sourceUnit.OutputCount != 1 {
		return 0xFFFFFFFF, false
	}
	targetUnitIndex := s.Assignments[sourceUnit.AssignmentOffset]
	targetUnit := &s.Processings[targetUnitIndex]
	if targetUnit.InputCount != 1 {
		return 0xFFFFFFFF, false
	}
	return targetUnitIndex, true
}

// ProcessingUnit represents a single processing unit in the audio graph
// along with the number of output assignments it has.
type ProcessingUnit struct {

	// Processor is the audio processing unit.
	Processor Processor

	// InputCount specifies how many input connections this processor has.
	InputCount uint32

	// OutputCount specifies how many output assignments this processor has.
	OutputCount uint32

	// AssignmentOffset specifies the starting index in the Assignments slice
	// where this processor's output assignments can be found.
	AssignmentOffset uint32
}
