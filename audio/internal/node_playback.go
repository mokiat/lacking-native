package internal

import (
	"sync"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/audio"
)

func NewPlaybackNode(player *Player, media *Media) *PlaybackNode {
	return &PlaybackNode{
		player:     player,
		samples:    media.frames,
		sampleRate: media.sampleRate,

		change:    opt.Unspecified[playbackChange](),
		loopStart: 0,
		loopEnd:   uint64(len(media.frames)),
		loop:      false,
		playing:   false,

		offset: 0,
	}
}

type PlaybackNode struct {
	audio.Node // marker interface

	player     *Player
	samples    []audio.Frame
	sampleRate int

	// The following fields are protected by the mutex and can be accessed from
	// any thread.
	mu        sync.Mutex
	change    opt.T[playbackChange]
	loopStart uint64
	loopEnd   uint64
	loop      bool
	playing   bool

	// The following fields are used only during processing
	// and should be changed only from the processing thread.
	offset uint64
}

var _ Node = (*PlaybackNode)(nil)
var _ audio.PlaybackNode = (*PlaybackNode)(nil)

func (n *PlaybackNode) Process(ctx ProcessContext, _, outputFrames FrameList) {
	n.mu.Lock()
	if change, ok := n.change.Unwrap(); ok {
		n.playing = change.playing
		if change.changeOffset {
			n.offset = change.offset
		}
		n.change = opt.Unspecified[playbackChange]()
	}
	loopStart := n.loopStart // store value locally to avoid long locks
	loopEnd := n.loopEnd     // store value locally to avoid long locks
	loop := n.loop           // store value locally to avoid long locks
	playing := n.playing     // store value locally to avoid long locks
	n.mu.Unlock()

	if !playing {
		return
	}

	length := uint64(len(n.samples))
	loopEnd = min(loopEnd, length)
	if loop && (loopStart >= loopEnd) {
		loopStart = 0
		loopEnd = length
	}

	offset := n.offset
	for i := range outputFrames {
		if offset < length {
			outputFrames[i] = Frame(n.samples[offset])
		}
		offset++
		if loop && (offset >= loopEnd) {
			offset = loopStart
		}
	}
	n.offset = offset

	n.mu.Lock()
	if !n.change.Specified { // don't change playing state if there is a pending change
		n.playing = offset < length
	}
	n.mu.Unlock()
}

func (n *PlaybackNode) Start(startTime float32) {
	startOffset := audio.SampleCount(startTime, n.player.SampleRate())

	n.mu.Lock()
	defer n.mu.Unlock()

	n.change = opt.V(playbackChange{
		offset:       uint64(startOffset),
		changeOffset: true,
		playing:      true,
	})
}

func (n *PlaybackNode) Stop() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.change = opt.V(playbackChange{
		offset:       0,
		changeOffset: true,
		playing:      false,
	})
}

func (n *PlaybackNode) Resume() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.change.Specified {
		n.change.Value.playing = true
	} else {
		n.change = opt.V(playbackChange{
			playing: true,
		})
	}
}

func (n *PlaybackNode) Pause() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.change.Specified {
		n.change.Value.playing = false
	} else {
		n.change = opt.V(playbackChange{
			playing: false,
		})
	}
}

func (n *PlaybackNode) IsPlaying() bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.playing
}

func (n *PlaybackNode) IsLoop() bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.loop
}

func (n *PlaybackNode) SetLoop(loop bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.loop = loop
}

func (n *PlaybackNode) LoopStart() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return audio.Seconds(int(n.loopStart), n.sampleRate)
}

func (n *PlaybackNode) SetLoopStart(loopStart float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.loopStart = uint64(audio.SampleCount(loopStart, n.sampleRate))
}

func (n *PlaybackNode) LoopEnd() float32 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return audio.Seconds(int(n.loopEnd), n.sampleRate)
}

func (n *PlaybackNode) SetLoopEnd(loopEnd float32) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.loopEnd = uint64(audio.SampleCount(loopEnd, n.sampleRate))
}

func (n *PlaybackNode) Delete() {
	n.Stop()
	n.player.DeletePlaybackNode(n)
}

type playbackChange struct {
	offset       uint64
	changeOffset bool
	playing      bool
}
