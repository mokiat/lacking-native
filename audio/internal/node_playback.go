package internal

import (
	"sync/atomic"

	"github.com/mokiat/lacking/audio"
)

func NewPlaybackNode(player *Player, media *Media, loop bool) *PlaybackNode {
	return &PlaybackNode{
		player: player,
		media:  media,
		loop:   loop,
	}
}

type PlaybackNode struct {
	player *Player
	media  *Media
	loop   bool
	offset atomic.Uint64
}

var _ Node = (*PlaybackNode)(nil)
var _ audio.PlaybackNode = (*PlaybackNode)(nil)

func (n *PlaybackNode) Process(ctx ProcessContext, _, outputFrames FrameList) {
	length := n.media.length
	leftSamples := n.media.leftChannel.samples
	rightSamples := n.media.rightChannel.samples
	offset := n.offset.Load()
	for i := range outputFrames {
		if offset >= length {
			outputFrames[i] = Frame{}
			continue
		}
		outputFrames[i] = Frame{
			Left:  leftSamples[offset],
			Right: rightSamples[offset],
		}
		offset++
		if n.loop {
			offset %= uint64(length)
		}
	}
	n.offset.Store(offset)
}

func (n *PlaybackNode) Loop() bool {
	return n.loop
}

func (n *PlaybackNode) Done() bool {
	length := n.media.length
	offset := n.offset.Load()
	return offset >= length
}

func (n *PlaybackNode) Delete() {
	n.player.DeletePlayback(n)
}
