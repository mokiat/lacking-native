package internal

import (
	"github.com/mokiat/lacking/core/audio"
)

type Buffer struct {
	frames []audio.Frame
	offset int
}

func NewBuffer(initialCapacity uint32) *Buffer {
	return &Buffer{
		frames: make([]audio.Frame, initialCapacity),
		offset: 0,
	}
}

func (b *Buffer) Reset() {
	b.offset = 0
}

func (b *Buffer) Allocate(count int) []audio.Frame {
	from := b.offset
	to := from + count

	if gap := to - len(b.frames); gap > 0 {
		b.frames = append(b.frames, make([]audio.Frame, gap)...)
	}

	b.offset += count
	result := b.frames[from:to]
	clear(result)
	return result
}
