package internal

import (
	"sync"

	"github.com/mokiat/lacking/audio"
)

func NewPlaybackNode(player *Player, media *Media, loop bool) *PlaybackNode {
	return &PlaybackNode{
		player:     player,
		samples:    media.samples,
		sampleRate: media.sampleRate,

		state: playbackState{
			loopStart: 0,
			loopEnd:   uint64(len(media.samples)),
			offset:    0,
			revision:  0,
			loop:      loop,
			playing:   false,
		},
	}
}

type PlaybackNode struct {
	audio.Node // marker interface

	player     *Player
	samples    []audio.Sample
	sampleRate int

	mu    sync.Mutex
	state playbackState
}

var _ Node = (*PlaybackNode)(nil)
var _ audio.PlaybackNode = (*PlaybackNode)(nil)

func (n *PlaybackNode) Process(ctx ProcessContext, _, outputFrames FrameList) {
	state := n.fetchState()
	if !state.playing {
		return
	}

	offset := state.offset
	length := uint64(len(n.samples))

	for i := range outputFrames {
		if offset < length {
			outputFrames[i] = Frame(n.samples[offset])
		}
		offset++
		if state.loop && (offset >= state.loopEnd) {
			offset = state.loopStart
		}
	}

	state.offset = offset
	state.playing = offset < length
	n.replaceState(state)
}

func (n *PlaybackNode) Start(startTime float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.state.playing = true
	n.state.offset = 0
	n.state.revision++
}

func (n *PlaybackNode) Stop() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.state.playing = false
	n.state.offset = 0
	n.state.revision++
}

func (n *PlaybackNode) Resume() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.state.playing = true
	n.state.revision++
}

func (n *PlaybackNode) Pause() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.state.playing = false
	n.state.revision++
}

func (n *PlaybackNode) IsPlaying() bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.state.playing
}

func (n *PlaybackNode) IsLoop() bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.state.loop
}

func (n *PlaybackNode) SetLoop(loop bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Note: Looping does not change the revision, as otherwise it will
	// cause offset changes done by the processor in the meantime to be dropped.
	n.state.loop = loop
}

func (n *PlaybackNode) LoopStart() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return float32(n.state.loopStart) / float32(n.sampleRate)
}

func (n *PlaybackNode) SetLoopStart(loopStart float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Note: Looping does not change the revision, as otherwise it will
	// cause offset changes done by the processor in the meantime to be dropped.
	n.state.loopStart = uint64(loopStart * float32(n.sampleRate))
}

func (n *PlaybackNode) LoopEnd() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return float32(n.state.loopEnd) / float32(n.sampleRate)
}

func (n *PlaybackNode) SetLoopEnd(loopEnd float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Note: Looping does not change the revision, as otherwise it will
	// cause offset changes done by the processor in the meantime to be dropped.
	n.state.loopEnd = uint64(loopEnd * float32(n.sampleRate))
}

func (n *PlaybackNode) Delete() {
	n.Stop()
	n.player.DeletePlaybackNode(n)
}

func (n *PlaybackNode) fetchState() playbackState {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.state
}

func (n *PlaybackNode) replaceState(candidate playbackState) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if candidate.revision == n.state.revision {
		n.state = candidate
	}
}

type playbackState struct {
	offset    uint64
	revision  uint64
	loopStart uint64
	loopEnd   uint64
	loop      bool
	playing   bool
}
