package internal

import "github.com/mokiat/lacking/audio"

func NewLowPassNode(player *Player) *LowPassNode {
	return &LowPassNode{}
}

type LowPassNode struct {
	player *Player
}

var _ Node = (*LowPassNode)(nil)
var _ audio.LowPassNode = (*LowPassNode)(nil)

func (n *LowPassNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	// TODO: Implement!
	copy(outputFrames, inputFrames)
}

func (n *LowPassNode) CutoffFrequency() float32 {
	panic("TODO")
}

func (n *LowPassNode) SetCutoffFrequency(frequency float32) {
	panic("TODO")
}

func (n *LowPassNode) Delete() {
	n.player.DeleteLowPassNode(n)
}
