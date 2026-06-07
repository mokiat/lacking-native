package internal

import (
	"github.com/mokiat/lacking/core/audio"
)

// Buffer is a bump allocator for audio frames. It is intended for use on a
// single goroutine as per-callback scratch space: call Reset at the start of
// each callback, then Allocate to hand out non-overlapping slices to each
// processing node.
type Buffer struct {
	frames []audio.Frame
	offset int
}

// NewBuffer creates a Buffer with the given initial frame capacity.
func NewBuffer(initialCapacity uint32) *Buffer {
	return &Buffer{
		frames: make([]audio.Frame, initialCapacity),
		offset: 0,
	}
}

// Reset rewinds the allocator to the beginning of the buffer, making all
// previously allocated slices available for reuse.
func (b *Buffer) Reset() {
	b.offset = 0
}

// Allocate returns a zeroed slice of count frames from the buffer, growing
// the backing allocation if necessary.
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
