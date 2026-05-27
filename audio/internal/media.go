package internal

import (
	"time"

	"github.com/mokiat/lacking/audio"
)

type Media struct {
	sampleRate int
	frames     []audio.Frame
}

var _ audio.Media = (*Media)(nil)

func (m *Media) Length() time.Duration {
	frameCount := len(m.frames)
	return (time.Second * time.Duration(frameCount)) / time.Duration(m.sampleRate)
}

func (m *Media) Delete() {
	m.frames = nil
}
