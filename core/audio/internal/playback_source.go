package internal

import "github.com/mokiat/lacking/core/audio"

type PlaybackSource struct {
}

var _ Processor = (*PlaybackSource)(nil)

func NewPlaybackSource(media *Media, sampleRate int) *PlaybackSource {
	return &PlaybackSource{}
}

func (s *PlaybackSource) Start(at float32) {
	panic("TODO")
}

func (s *PlaybackSource) Stop() {
	panic("TODO")
}

func (s *PlaybackSource) Pause() {
	panic("TODO")
}

func (s *PlaybackSource) Resume() {
	panic("TODO")
}

func (s *PlaybackSource) Looping() bool {
	panic("TODO")
}

func (s *PlaybackSource) SetLooping(loop bool) {
	panic("TODO")
}

func (s *PlaybackSource) LoopStart() float32 {
	panic("TODO")
}

func (s *PlaybackSource) SetLoopStart(loopStart float32) {
	panic("TODO")
}

func (s *PlaybackSource) LoopEnd() float32 {
	panic("TODO")
}

func (s *PlaybackSource) SetLoopEnd(loopEnd float32) {
	panic("TODO")
}

func (s *PlaybackSource) Playing() bool {
	panic("TODO")
}

func (s *PlaybackSource) PlaybackRate() float32 {
	panic("TODO")
}

func (s *PlaybackSource) SetPlaybackRate(rate float32) {
	panic("TODO")
}

func (s *PlaybackSource) SetOnFinished(onFinished func()) {
	panic("TODO")
}

func (s *PlaybackSource) Process(ctx ProcessContext, _ []audio.Frame) []audio.Frame {
	panic("TODO")
}
