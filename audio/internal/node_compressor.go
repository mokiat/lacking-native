package internal

import "github.com/mokiat/lacking/audio"

func NewCompressorNode(player *Player) *CompressorNode {
	return &CompressorNode{}
}

type CompressorNode struct {
	player *Player
}

var _ Node = (*CompressorNode)(nil)
var _ audio.CompressorNode = (*CompressorNode)(nil)

func (n *CompressorNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	// TODO: Implement!
	copy(outputFrames, inputFrames)
}

func (n *CompressorNode) Threshold() float32 {
	panic("TODO")
}

func (n *CompressorNode) SetThreshold(threshold float32) {
	panic("TODO")
}

func (n *CompressorNode) Delete() {
	n.player.DeleteCompressorNode(n)
}
