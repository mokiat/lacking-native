package internal

import (
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

func NewHighPassNode(player *Player) *HighPassNode {
	return &HighPassNode{
		player: player,

		cutoffFrequency: 350.0,
	}
}

type HighPassNode struct {
	player *Player

	mu              sync.Mutex
	cutoffFrequency float32

	inputNeg1 Frame
	inputNeg2 Frame

	outputNeg1 Frame
	outputNeg2 Frame
}

var _ Node = (*HighPassNode)(nil)
var _ audio.HighPassNode = (*HighPassNode)(nil)

func (n *HighPassNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	fc := n.CutoffFrequency()
	fs := float32(n.player.SampleRate())

	// This uses a second-order high-pass filter algorithm based on the
	// Audio EQ Cookbook by Robert Bristow-Johnson.
	// Reference: https://www.w3.org/TR/audio-eq-cookbook/
	angle := sprec.Radians(2.0 * sprec.Pi * (fc / fs))
	cs := sprec.Cos(angle)
	sn := sprec.Sin(angle)
	alpha := sn / sprec.Sqrt(2.0)

	a0 := 1.0 + alpha
	b0 := ((1.0 + cs) / 2.0) / a0
	b1 := -(1.0 + cs) / a0
	b2 := ((1.0 + cs) / 2.0) / a0
	a1 := (-2.0 * cs) / a0
	a2 := (1.0 - alpha) / a0

	for i := range outputFrames {
		input := inputFrames[i]

		outputFrames[i] = Frame{
			Left: b0*input.Left +
				b1*n.inputNeg1.Left +
				b2*n.inputNeg2.Left -
				a1*n.outputNeg1.Left -
				a2*n.outputNeg2.Left,

			Right: b0*input.Right +
				b1*n.inputNeg1.Right +
				b2*n.inputNeg2.Right -
				a1*n.outputNeg1.Right -
				a2*n.outputNeg2.Right,
		}

		n.inputNeg2 = n.inputNeg1
		n.inputNeg1 = input

		n.outputNeg2 = n.outputNeg1
		n.outputNeg1 = outputFrames[i]
	}
}

func (n *HighPassNode) CutoffFrequency() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.cutoffFrequency
}

func (n *HighPassNode) SetCutoffFrequency(frequency float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.cutoffFrequency = sprec.Clamp(0.0, frequency, 44100.0)
}

func (n *HighPassNode) Delete() {
	n.player.DeleteHighPassNode(n)
}
