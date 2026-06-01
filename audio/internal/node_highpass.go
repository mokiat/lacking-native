package internal

import (
	"sync"

	"github.com/mokiat/lacking/audio"
)

func newHighPassNode(player *Player) *HighPassNode {
	filterL := NewHighPassFilter()
	filterL.Configure(audio.DefaultCutoffFrequency, defaultSampleRate)

	filterR := NewHighPassFilter()
	filterR.Configure(audio.DefaultCutoffFrequency, defaultSampleRate)

	return &HighPassNode{
		player: player,

		cutoffFrequency: audio.DefaultCutoffFrequency,

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
	n.mu.Lock()
	fc := n.cutoffFrequency // store value locally to avoid long locks
	n.mu.Unlock()
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
