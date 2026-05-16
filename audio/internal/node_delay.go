package internal

import (
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

const maxDelaySeconds = 1.0

func NewDelayNode(player *Player) *DelayNode {

	maxDelaySamples := int(maxDelaySeconds * float32(player.SampleRate()))
	filterL := NewDelayFilter(maxDelaySamples)
	filterR := NewDelayFilter(maxDelaySamples)

	return &DelayNode{
		player: player,

		delayTime: audio.DefaultDelay,

		filterL: filterL,
		filterR: filterR,
	}
}

type DelayNode struct {
	audio.Node // marker interface

	player *Player

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu        sync.Mutex
	delayTime float32

	// The following filters are used only during processing
	// and should be configured only from the processing thread.
	filterL *DelayFilter
	filterR *DelayFilter
}

var _ Node = (*DelayNode)(nil)
var _ audio.DelayNode = (*DelayNode)(nil)

func (n *DelayNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	n.mu.Lock()
	delay := sprec.Clamp(n.delayTime, 0.0, maxDelaySeconds) // store value locally to avoid long locks
	n.mu.Unlock()

	const delayThreshold = 0.00001
	if delay < delayThreshold {
		copy(outputFrames, inputFrames)
		return // no delay, just pass through
	}

	delaySamples := audio.SampleCount(delay, n.player.SampleRate())
	n.filterL.Configure(delaySamples)
	n.filterR.Configure(delaySamples)

	for i := range outputFrames {
		input := inputFrames[i]
		outputFrames[i] = Frame{
			Left:  n.filterL.Process(input.Left),
			Right: n.filterR.Process(input.Right),
		}
	}
}

func (n *DelayNode) DelayTime() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.delayTime
}

func (n *DelayNode) SetDelayTime(delayTime float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.delayTime = delayTime
}

func (n *DelayNode) Delete() {
	n.player.DeleteDelayNode(n)
}
