package internal

import (
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

func NewLowPassNode(player *Player) *LowPassNode {
	return &LowPassNode{
		player: player,

		cutoffFrequency: 350.0,
	}
}

type LowPassNode struct {
	player *Player

	mu              sync.Mutex
	cutoffFrequency float32

	prevFrame Frame
}

var _ Node = (*LowPassNode)(nil)
var _ audio.LowPassNode = (*LowPassNode)(nil)

func (n *LowPassNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	fc := n.CutoffFrequency()
	fs := float32(n.player.SampleRate())

	// This uses a first-order low-pass filter algorithm.
	alpha := 2.0 * sprec.Pi * fc / (2.0*sprec.Pi*fc + fs)

	for i := range outputFrames {
		input := inputFrames[i]
		outputFrames[i] = Frame{
			Left:  sprec.Mix(n.prevFrame.Left, input.Left, alpha),
			Right: sprec.Mix(n.prevFrame.Right, input.Right, alpha),
		}
		n.prevFrame = outputFrames[i]
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
	n.cutoffFrequency = sprec.Clamp(0.0, frequency, 44100.0)
}

func (n *LowPassNode) Delete() {
	n.player.DeleteLowPassNode(n)
}
