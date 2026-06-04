package internal

import "github.com/mokiat/lacking/core/audio"

type Media struct {
	sampleRate int
	frames     []audio.Frame
}

var _ audio.Media = (*Media)(nil)

func NewMedia(data audio.MediaData) *Media {
	return &Media{
		sampleRate: data.SampleRate,
		frames:     data.Frames,
	}
}

func (m *Media) Length() float32 {
	frameCount := len(m.frames)
	return audio.Seconds(frameCount, m.sampleRate)
}

func (m *Media) Release() {
	m.frames = nil
}
