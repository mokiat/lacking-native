package internal

import (
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

func newOscillatorNode(player *Player) *OscillatorNode {
	return &OscillatorNode{
		player: player,

		frequency: audio.DefaultFrequency,

		angle: sprec.Radians(0.0),
	}
}

type OscillatorNode struct {
	audio.Node // marker interface

	player *Player

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu        sync.Mutex
	frequency float32

	// The following fields are used only during processing
	// and should be changed only from the processing thread.
	angle sprec.Angle
}

var _ Node = (*OscillatorNode)(nil)
var _ audio.OscillatorNode = (*OscillatorNode)(nil)

func (n *OscillatorNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	n.mu.Lock()
	frequency := n.frequency // store frequency locally to avoid long locks
	n.mu.Unlock()

	deltaAngle := sprec.Radians(frequency * (2.0 * sprec.Pi / float32(ctx.SampleRate)))

	for i := range outputFrames {
		sn := sprec.Sin(n.angle)
		outputFrames[i] = Frame{
			Left:  sn,
			Right: sn,
		}
		n.angle += deltaAngle
		n.angle = sprec.NormalizeAnglePos(n.angle)
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
