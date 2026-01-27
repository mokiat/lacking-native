package internal

import "github.com/mokiat/lacking/audio"

func NewHighPassNode(player *Player) *HighPassNode {
	return &HighPassNode{}
}

type HighPassNode struct {
	player *Player
}

var _ Node = (*HighPassNode)(nil)
var _ audio.HighPassNode = (*HighPassNode)(nil)

func (n *HighPassNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	// TODO: Implement!
	copy(outputFrames, inputFrames)
}

func (n *HighPassNode) CutoffFrequency() float32 {
	panic("TODO")
}

func (n *HighPassNode) SetCutoffFrequency(frequency float32) {
	panic("TODO")
}

func (n *HighPassNode) Delete() {
	n.player.DeleteHighPassNode(n)
}
