package internal

import (
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

func NewOscillatorNode(player *Player) *OscillatorNode {
	return &OscillatorNode{
		player: player,

		angle: sprec.Radians(0.0),

		frequency: 440.0,
	}
}

type OscillatorNode struct {
	audio.Node // marker interface

	player *Player

	angle sprec.Angle

	mu        sync.Mutex
	frequency float32
}

var _ Node = (*OscillatorNode)(nil)
var _ audio.OscillatorNode = (*OscillatorNode)(nil)

func (n *OscillatorNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	frequency := n.Frequency() // store frequency locally to avoid long locks

	deltaAngle := sprec.Radians(frequency * (2.0 * sprec.Pi / float32(ctx.SampleRate)))

	for i := range outputFrames {
		sn := sprec.Sin(n.angle)
		outputFrames[i] = Frame{
			Left:  sn,
			Right: sn,
		}
		n.angle += deltaAngle
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
	n.player.DeleteOscillatorNode(n)
}
