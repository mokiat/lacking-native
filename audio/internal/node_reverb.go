package internal

import "github.com/mokiat/lacking/audio"

func NewReverbNode(player *Player) *ReverbNode {
	return &ReverbNode{}
}

type ReverbNode struct {
	player *Player
}

var _ Node = (*ReverbNode)(nil)
var _ audio.ReverbNode = (*ReverbNode)(nil)

func (n *ReverbNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	// TODO: Implement!
	copy(outputFrames, inputFrames)
}

func (n *ReverbNode) RoomSize() float32 {
	panic("TODO")
}

func (n *ReverbNode) SetRoomSize(size float32) {
	panic("TODO")
}

func (n *ReverbNode) Delete() {
	n.player.DeleteReverbNode(n)
}
