package internal

import (
	"math"
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

func NewCompressorNode(player *Player) *CompressorNode {
	const defaultThreshold = -24.0

	return &CompressorNode{
		player: player,

		threshold:   defaultThreshold,
		currentGain: 1.0,
	}
}

type CompressorNode struct {
	player *Player

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu        sync.Mutex
	threshold float32

	// The following fields are used only during processing.
	currentGain float32
}

var _ Node = (*CompressorNode)(nil)
var _ audio.CompressorNode = (*CompressorNode)(nil)

func (n *CompressorNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	threshold := n.Threshold() // store value locally to avoid long locks
	sampleRate := float32(n.player.SampleRate())

	const (
		ratio   = 12.0
		knee    = 30.0
		attack  = 0.003
		release = 0.25

		invRatio = 1.0 / ratio
	)

	attackCoeff := float32(math.Exp(-1.0 / (float64(sampleRate) * attack)))
	releaseCoeff := float32(math.Exp(-1.0 / (float64(sampleRate) * release)))

	for i, frame := range inputFrames {
		peak := max(sprec.Abs(frame.Left), sprec.Abs(frame.Right))
		peak = max(1.0e-8, peak) // avoid log of zero
		peakDB := audio.GainToDB(peak)

		targetDB := float32(0.0)
		if peakDB > threshold { // needs compression
			if peakDB >= (threshold + knee) { // hard compression
				z := peakDB - threshold
				targetDB = (invRatio - 1.0) * z
			} else { // soft compression
				z := peakDB - threshold
				targetDB = (invRatio - 1.0) * ((2.0 * z * z / knee) - (z * z * z / (knee * knee)))
			}
		}

		targetGain := audio.DBToGain(targetDB)
		if targetGain < n.currentGain {
			n.currentGain = sprec.Mix(targetGain, n.currentGain, attackCoeff)
		} else {
			n.currentGain = sprec.Mix(targetGain, n.currentGain, releaseCoeff)
		}

		outputFrames[i] = Frame{
			Left:  frame.Left * n.currentGain,
			Right: frame.Right * n.currentGain,
		}
	}
}

func (n *CompressorNode) Threshold() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.threshold
}

func (n *CompressorNode) SetThreshold(threshold float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.threshold = sprec.Clamp(threshold, -100.0, 0.0)
}

func (n *CompressorNode) Delete() {
	n.player.DeleteCompressorNode(n)
}
