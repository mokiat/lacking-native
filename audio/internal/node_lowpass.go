package internal

import (
	"sync"

	"github.com/mokiat/lacking/audio"
)

func NewLowPassNode(player *Player) *LowPassNode {
	filterL := NewLowPassFilter()
	filterL.Configure(audio.DefaultCutoffFrequency, defaultSampleRate)

	filterR := NewLowPassFilter()
	filterR.Configure(audio.DefaultCutoffFrequency, defaultSampleRate)

	return &LowPassNode{
		player: player,

		cutoffFrequency: audio.DefaultCutoffFrequency,

		filterL: filterL,
		filterR: filterR,
	}
}

type LowPassNode struct {
	audio.Node // marker interface

	player *Player

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu              sync.Mutex
	cutoffFrequency float32

	// The following filters are used only during processing
	// and should be configured only from the processing thread.
	filterL *LowPassFilter
	filterR *LowPassFilter
}

var _ Node = (*LowPassNode)(nil)
var _ audio.LowPassNode = (*LowPassNode)(nil)

func (n *LowPassNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	n.mu.Lock()
	fc := n.cutoffFrequency // store value locally to avoid long locks
	n.mu.Unlock()
	fs := n.player.SampleRate()

	const threshold = 0.1
	if fc <= threshold {
		return // nothing passes through
	}
	nyquist := float32(fs) / 2.0
	if fc >= (nyquist - threshold) {
		copy(outputFrames, inputFrames)
		return // everything passes through
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

func (n *LowPassNode) CutoffFrequency() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.cutoffFrequency
}

func (n *LowPassNode) SetCutoffFrequency(frequency float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.cutoffFrequency = frequency
}

func (n *LowPassNode) Delete() {
	n.player.DeleteLowPassNode(n)
}
