package internal

import (
	"sync"

	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

func NewPanNode(player *Player) *PanNode {
	return &PanNode{
		player: player,

		pan: 0.0,
	}
}

type PanNode struct {
	player *Player

	propMU sync.Mutex
	pan    float32
}

var _ Node = (*PanNode)(nil)
var _ audio.PanNode = (*PanNode)(nil)

func (n *PanNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	pan := n.Pan() // store gain locally to avoid long locks

	// This implementation is consistent with WebAudio's stereo panning algorithm.
	// https://webaudio.github.io/web-audio-api/#stereopanner-algorithm

	// This works by treating the pan value as an angle on a semicircle.
	// Furthermore, it always keeps one channel fully audible, while the other
	// channel is attenuated and blended based on the pan position.
	// This is a bit different from standard mono panning approaches.

	angleFraction := gog.Ternary(pan >= 0.0, pan, pan+1.0)
	panAngle := sprec.Radians(angleFraction * sprec.Pi / 2.0)
	leftGain := sprec.Cos(panAngle)
	rightGain := sprec.Sin(panAngle)
	if pan < 0.0 {
		for i, frame := range inputFrames {
			outputFrames[i] = Frame{
				Left:  (frame.Right * leftGain) + frame.Left,
				Right: (frame.Right * rightGain),
			}
		}
	} else {
		for i, frame := range inputFrames {
			outputFrames[i] = Frame{
				Left:  (frame.Left * leftGain),
				Right: (frame.Left * rightGain) + frame.Right,
			}
		}
	}
}

func (n *PanNode) Pan() float32 {
	n.propMU.Lock()
	defer n.propMU.Unlock()
	return n.pan
}

func (n *PanNode) SetPan(pan float32) {
	n.propMU.Lock()
	defer n.propMU.Unlock()
	n.pan = sprec.Clamp(pan, -1.0, 1.0)
}

func (n *PanNode) Delete() {
	n.player.DeletePan(n)
}
