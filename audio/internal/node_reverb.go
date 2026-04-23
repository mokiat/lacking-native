package internal

import (
	"sync"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
)

const (
	reverbCombFilterCount       = 4
	reverbCombFilterScale       = 1.0 / float32(reverbCombFilterCount)
	reverbAllPassFilterCount    = 2
	reverbAllPassFilterFeedback = 0.7
)

var (
	reverbCombFilterDelays = [reverbCombFilterCount]float32{
		0.0297,
		0.0308,
		0.0323,
		0.0331,
	}

	reverbCombFilterSpread = [reverbCombFilterCount]float32{
		0.0003,
		0.0007,
		0.0011,
		0.0015,
	}

	reverbAllPassFilterDelays = [reverbAllPassFilterCount]float32{
		0.0043,
		0.0019,
	}

	reverbAllPassFilterSpread = [reverbAllPassFilterCount]float32{
		0.0012,
		0.0003,
	}
)

func NewReverbNode(player *Player) *ReverbNode {
	combDelaysL := [reverbCombFilterCount]int{}
	combDelaysR := [reverbCombFilterCount]int{}
	for i := range reverbCombFilterCount {
		spread := reverbCombFilterSpread[i]
		combDelaysL[i] = audio.SampleCount(reverbCombFilterDelays[i], player.SampleRate())
		combDelaysR[i] = audio.SampleCount(reverbCombFilterDelays[i]+spread, player.SampleRate())
	}

	combsL := [reverbCombFilterCount]*FeedbackCombFilter{}
	combsR := [reverbCombFilterCount]*FeedbackCombFilter{}
	for i := range reverbCombFilterCount {
		combsL[i] = NewFeedbackCombFilter(combDelaysL[i])
		combsL[i].Configure(0.0, combDelaysL[i])
		combsR[i] = NewFeedbackCombFilter(combDelaysR[i])
		combsR[i].Configure(0.0, combDelaysR[i])
	}

	allPassDelaysL := [reverbAllPassFilterCount]int{}
	allPassDelaysR := [reverbAllPassFilterCount]int{}
	for i := range reverbAllPassFilterCount {
		spread := reverbAllPassFilterSpread[i]
		allPassDelaysL[i] = audio.SampleCount(reverbAllPassFilterDelays[i], player.SampleRate())
		allPassDelaysR[i] = audio.SampleCount(reverbAllPassFilterDelays[i]+spread, player.SampleRate())
	}

	allPassesL := [reverbAllPassFilterCount]*AllPassFilter{}
	allPassesR := [reverbAllPassFilterCount]*AllPassFilter{}
	for i := range reverbAllPassFilterCount {
		allPassesL[i] = NewAllPassFilter(allPassDelaysL[i])
		allPassesL[i].Configure(reverbAllPassFilterFeedback, allPassDelaysL[i])
		allPassesR[i] = NewAllPassFilter(allPassDelaysR[i])
		allPassesR[i].Configure(reverbAllPassFilterFeedback, allPassDelaysR[i])
	}

	const defaultRoomSize = 0.3

	return &ReverbNode{
		player: player,

		roomSize: defaultRoomSize,

		combDelaysL: combDelaysL,
		combDelaysR: combDelaysR,
		combsL:      combsL,
		combsR:      combsR,
		allPassesL:  allPassesL,
		allPassesR:  allPassesR,
	}
}

type ReverbNode struct {
	audio.Node // marker interface

	player *Player

	mu       sync.Mutex
	roomSize float32

	combDelaysL [reverbCombFilterCount]int
	combDelaysR [reverbCombFilterCount]int
	combsL      [reverbCombFilterCount]*FeedbackCombFilter
	combsR      [reverbCombFilterCount]*FeedbackCombFilter
	allPassesL  [reverbAllPassFilterCount]*AllPassFilter
	allPassesR  [reverbAllPassFilterCount]*AllPassFilter
}

var _ Node = (*ReverbNode)(nil)
var _ audio.ReverbNode = (*ReverbNode)(nil)

func (n *ReverbNode) Process(ctx ProcessContext, inputFrames, outputFrames FrameList) {
	roomSize := n.RoomSize() // store value locally to avoid long locks

	for i := range reverbCombFilterCount {
		feedback := 0.3 + roomSize*0.6
		n.combsL[i].Configure(feedback, n.combDelaysL[i])
		n.combsR[i].Configure(feedback, n.combDelaysR[i])
	}

	for i, frame := range inputFrames {
		// Run through parallel comb filters.
		outputLeft := float32(0.0)
		outputRight := float32(0.0)
		for c := range reverbCombFilterCount {
			outputLeft += n.combsL[c].Process(frame.Left)
			outputRight += n.combsR[c].Process(frame.Right)
		}
		outputLeft *= reverbCombFilterScale
		outputRight *= reverbCombFilterScale

		// Run through serial all-pass filters.
		for j := range reverbAllPassFilterCount {
			outputLeft = n.allPassesL[j].Process(outputLeft)
			outputRight = n.allPassesR[j].Process(outputRight)
		}

		outputFrames[i] = Frame{
			Left:  sprec.Mix(outputLeft, outputRight, 0.2),
			Right: sprec.Mix(outputRight, outputLeft, 0.2),
		}
	}
}

func (n *ReverbNode) RoomSize() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.roomSize
}

func (n *ReverbNode) SetRoomSize(roomSize float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.roomSize = sprec.Clamp(roomSize, 0.0, 1.0)
}

func (n *ReverbNode) Delete() {
	n.player.DeleteReverbNode(n)
}
