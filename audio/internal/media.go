package internal

import (
	"time"

	"github.com/mokiat/lacking/audio"
)

type Media struct {
	sampleRate int
	samples    []audio.Sample
}

var _ audio.Media = (*Media)(nil)

func (m *Media) Length() time.Duration {
	frameCount := len(m.samples)
	return (time.Second * time.Duration(frameCount)) / time.Duration(m.sampleRate)
}

func (m *Media) Delete() {
	m.samples = nil
}
