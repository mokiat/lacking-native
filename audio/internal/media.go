package internal

import (
	"time"

	"github.com/veandco/go-sdl2/mix"
)

type Media struct {
	chunk *mix.Chunk
}

func (m *Media) Length() time.Duration {
	return time.Duration(m.chunk.LengthInMs()) * time.Millisecond
}

func (m *Media) Delete() {
	if m.chunk != nil {
		m.chunk.Free()
		m.chunk = nil
	}
}
