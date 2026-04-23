package internal

import "github.com/mokiat/lacking/audio"

func NewConnectorNode(player *Player) *ConnectorNode {
	return &ConnectorNode{
		player: player,
	}
}

type ConnectorNode struct {
	audio.Node // marker interface

	player *Player
}

var _ Node = (*ConnectorNode)(nil)
var _ audio.ConnectorNode = (*ConnectorNode)(nil)

func (n *ConnectorNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	copy(outputFrames, inputFrames)
}

func (n *ConnectorNode) Delete() {
	n.player.DeleteConnectorNode(n)
}
