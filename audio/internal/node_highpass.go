package internal

import (
	"sync"

	"github.com/mokiat/lacking/audio"
)

func NewHighPassNode(player *Player) *HighPassNode {
	const defaultCutoffFrequency = 350.0

	filterL := NewHighPassFilter()
	filterL.Configure(defaultCutoffFrequency, defaultSampleRate)

	filterR := NewHighPassFilter()
	filterR.Configure(defaultCutoffFrequency, defaultSampleRate)

	return &HighPassNode{
		player: player,

		cutoffFrequency: defaultCutoffFrequency,

		filterL: filterL,
		filterR: filterR,
	}
}

type HighPassNode struct {
	audio.Node // marker interface

	player *Player

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu              sync.Mutex
	cutoffFrequency float32

	// The following filters are used only during processing
	// and should be configured only from the processing thread.
	filterL *HighPassFilter
	filterR *HighPassFilter
}

var _ Node = (*HighPassNode)(nil)
var _ audio.HighPassNode = (*HighPassNode)(nil)

func (n *HighPassNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	fc := n.CutoffFrequency() // store value locally to avoid long locks
	fs := n.player.SampleRate()

	const threshold = 0.1
	if fc <= threshold {
		copy(outputFrames, inputFrames)
		return // everything passes through
	}
	nyquist := float32(fs) / 2.0
	if fc >= (nyquist - threshold) {
		return // nothing passes through
	}

	n.filterL.Configure(fc, fs)
	n.filterR.Configure(fc, fs)

	for i := range outputFrames {
		input := inputFrames[i]
		outputFrames[i] = Frame{
			Left:  n.filterL.Process(input.Left),
			Right: n.filterR.Process(input.Right),
		}
	}
}

func (n *HighPassNode) CutoffFrequency() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.cutoffFrequency
}

func (n *HighPassNode) SetCutoffFrequency(frequency float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.cutoffFrequency = frequency
}

func (n *HighPassNode) Delete() {
	n.player.DeleteHighPassNode(n)
}
