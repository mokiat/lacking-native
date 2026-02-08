package internal

import (
	"sync"

	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

const maxDelayNodeSeconds = 1.0

func NewDelayNode(player *Player) *DelayNode {
	bufferSize := 1 + int(float32(player.SampleRate())*maxDelayNodeSeconds)

	return &DelayNode{
		player: player,

		delayTime: 0.0,

		buffer: make([]Frame, bufferSize),
	}
}

type DelayNode struct {
	player *Player

	mu        sync.Mutex
	delayTime float32

	buffer      []Frame
	writeOffset int64
}

var _ Node = (*DelayNode)(nil)
var _ audio.DelayNode = (*DelayNode)(nil)

func (n *DelayNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	delay := n.DelayTime() // store gain locally to avoid long locks

	const delayThreshold = 0.00001
	if delay < delayThreshold {
		copy(outputFrames, inputFrames)
		return // no delay, just pass through
	}

	bufferSize := int64(len(n.buffer))
	delayOffset := int64(float32(n.player.SampleRate()) * delay)
	delayOffset = min(delayOffset, bufferSize-1)

	writeOffset := n.writeOffset
	readOffset := gog.Ternary(writeOffset >= delayOffset,
		writeOffset-delayOffset,
		bufferSize+writeOffset-delayOffset,
	)

	// Note: This implementation writes and reads to and from the delay buffer
	// at the same time. It also orders samples in opposite order compared to the
	// input and output buffers.

	for i := range outputFrames {
		n.buffer[writeOffset] = inputFrames[i]
		outputFrames[i] = n.buffer[readOffset]
		writeOffset = (writeOffset + 1) % bufferSize
		readOffset = (readOffset + 1) % bufferSize
	}

	n.writeOffset = writeOffset
}

func (n *DelayNode) DelayTime() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.delayTime
}

func (n *DelayNode) SetDelayTime(delayTime float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.delayTime = sprec.Clamp(delayTime, 0.0, maxDelayNodeSeconds)
}

func (n *DelayNode) Delete() {
	n.player.DeleteDelayNode(n)
}
