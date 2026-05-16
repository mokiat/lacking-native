package internal

import (
	"math"
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

func NewCompressorNode(player *Player) *CompressorNode {
	return &CompressorNode{
		player: player,

		attack:    audio.DefaultAttack,
		release:   audio.DefaultRelease,
		ratio:     audio.DefaultRatio,
		knee:      audio.DefaultKnee,
		threshold: audio.DefaultThreshold,

		currentGain: 1.0,
	}
}

type CompressorNode struct {
	audio.Node // marker interface

	player *Player

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu        sync.Mutex
	attack    float32
	release   float32
	ratio     float32
	knee      float32
	threshold float32

	// The following fields are used only during processing.
	currentGain float32
}

var _ Node = (*CompressorNode)(nil)
var _ audio.CompressorNode = (*CompressorNode)(nil)

func (n *CompressorNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	n.mu.Lock()
	attack := n.attack       // store value locally to avoid long locks
	release := n.release     // store value locally to avoid long locks
	ratio := n.ratio         // store value locally to avoid long locks
	knee := n.knee           // store value locally to avoid long locks
	threshold := n.threshold // store value locally to avoid long locks
	n.mu.Unlock()

	sampleRate := float32(n.player.SampleRate())
	invRatioNegOne := (1.0 / ratio) - 1.0
	attackCoeff := float32(math.Exp(-1.0 / (float64(sampleRate * attack))))
	releaseCoeff := float32(math.Exp(-1.0 / (float64(sampleRate * release))))

	for i, frame := range inputFrames {
		peak := max(sprec.Abs(frame.Left), sprec.Abs(frame.Right))
		peak = max(1.0e-8, peak) // avoid log of zero
		peakDB := audio.GainToDB(peak)

		reductionDB := float32(0.0)
		if peakDB > threshold { // needs compression
			z := peakDB - threshold
			if peakDB >= (threshold + knee) { // hard compression
				reductionDB = invRatioNegOne * (z - knee/2.0)
			} else { // soft compression
				reductionDB = invRatioNegOne * (z * z) / (2.0 * knee)
			}
		}

		targetGain := audio.DBToGain(reductionDB)
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

func (n *CompressorNode) Attack() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.attack
}

func (n *CompressorNode) SetAttack(attack float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.attack = sprec.Clamp(attack, 0.0, 1.0)
}

func (n *CompressorNode) Release() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.release
}

func (n *CompressorNode) SetRelease(release float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.release = sprec.Clamp(release, 0.0, 1.0)
}

func (n *CompressorNode) Ratio() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.ratio
}

func (n *CompressorNode) SetRatio(ratio float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ratio = sprec.Clamp(ratio, 1.0, 20.0)
}

func (n *CompressorNode) Knee() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.knee
}

func (n *CompressorNode) SetKnee(knee float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.knee = sprec.Clamp(knee, 0.0, 40.0)
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
