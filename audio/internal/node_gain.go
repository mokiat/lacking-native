package internal

import (
	"sync"

	"github.com/mokiat/lacking/audio"
)

func NewGainNode(player *Player) *GainNode {
	return &GainNode{
		player: player,

		gain: audio.DefaultGain,
	}
}

type GainNode struct {
	audio.Node // marker interface

	player *Player

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu   sync.Mutex
	gain float32
}

var _ Node = (*GainNode)(nil)
var _ audio.GainNode = (*GainNode)(nil)

func (n *GainNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	n.mu.Lock()
	gain := n.gain // store value locally to avoid long locks
	n.mu.Unlock()

	const gainThreshold = 0.001
	if gain < gainThreshold {
		return // no output, just silence
	}

	for i, frame := range inputFrames {
		outputFrames[i] = Frame{
			Left:  frame.Left * gain,
			Right: frame.Right * gain,
		}
	}
}

func (n *GainNode) Gain() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.gain
}

func (n *GainNode) SetGain(gain float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.gain = max(0.0, gain)
}

func (n *GainNode) Delete() {
	n.player.DeleteGainNode(n)
}
