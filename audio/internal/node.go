package internal

import "github.com/mokiat/lacking/audio"

// Node represents a node in a chain of audio elements. Each node produces
// audio data which can be synthesized, processed, or played back.
type Node interface {
	audio.Node
	Processor
}
