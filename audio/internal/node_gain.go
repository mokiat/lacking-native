package internal

import (
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

func NewGainNode(player *Player) *GainNode {
	return &GainNode{
		player: player,
		gain:   1.0,
	}
}

type GainNode struct {
	player *Player

	mu   sync.Mutex
	gain float32
}

var _ Node = (*GainNode)(nil)
var _ audio.GainNode = (*GainNode)(nil)

func (n *GainNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	gain := n.Gain() // store gain locally to avoid long locks

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
	n.gain = sprec.Clamp(gain, 0.0, 1.0)
}

func (n *GainNode) Delete() {
	n.player.DeleteGain(n)
}
