package internal

import (
	"time"
)

type Media struct {
	frames []mediaSample
}

func (m *Media) Length() time.Duration {
	return time.Second * time.Duration(len(m.frames)) / 44100
}

func (m *Media) Delete() {
	m.frames = nil
}

type mediaSample struct {
	Left  int16
	Right int16
}
