package internal

import (
	"sync"

	"github.com/mokiat/lacking/audio"
)

func NewPanNode(player *Player) *PanNode {
	return &PanNode{
		player: player,

		pan: 1.0,
	}
}

type PanNode struct {
	player *Player

	mu  sync.Mutex
	pan float32
}

var _ Node = (*PanNode)(nil)
var _ audio.PanNode = (*PanNode)(nil)

func (n *PanNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	pan := n.Pan() // store gain locally to avoid long locks

	// TODO: Implement proper panning algorithm.
	_ = pan

	for i, frame := range inputFrames {
		outputFrames[i] = Frame{
			Left:  frame.Left,
			Right: frame.Right,
		}
	}
}

func (n *PanNode) Pan() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.pan
}

func (n *PanNode) SetPan(pan float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.pan = pan
}

func (n *PanNode) Delete() {
	n.player.DeletePan(n)
}
