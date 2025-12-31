package internal

import (
	"sync"

	"github.com/mokiat/lacking/audio"
)

func NewOscillatorNode(player *Player) *OscillatorNode {
	return &OscillatorNode{
		player: player,

		frequency: 440.0,
	}
}

type OscillatorNode struct {
	player *Player

	mu        sync.Mutex
	frequency float32
}

var _ Node = (*OscillatorNode)(nil)
var _ audio.OscillatorNode = (*OscillatorNode)(nil)

func (n *OscillatorNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	freqency := n.Frequency() // store frequency locally to avoid long locks

	// TODO: Implement oscillator waveform generation.
	_ = freqency

	for i := range outputFrames {
		outputFrames[i] = Frame{
			Left:  0.0,
			Right: 0.0,
		}
	}
}

func (n *OscillatorNode) Frequency() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.frequency
}

func (n *OscillatorNode) SetFrequency(frequency float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.frequency = frequency
}

func (n *OscillatorNode) Delete() {
	n.player.DeleteOscillator(n)
}
