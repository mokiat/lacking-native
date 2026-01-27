package internal

import "github.com/mokiat/lacking/audio"

func NewDelayNode(player *Player) *DelayNode {
	return &DelayNode{}
}

type DelayNode struct {
	player *Player
}

var _ Node = (*DelayNode)(nil)
var _ audio.DelayNode = (*DelayNode)(nil)

func (n *DelayNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	// TODO: Implement!
	copy(outputFrames, inputFrames)
}

func (n *DelayNode) DelayTime() float32 {
	panic("TODO")
}

func (n *DelayNode) SetDelayTime(delayTime float32) {
	panic("TODO")
}

func (n *DelayNode) Delete() {
	n.player.DeleteDelayNode(n)
}
